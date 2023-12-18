package consumer

import (
	"context"
	"errors"
	"fmt"
	"github.com/zhangweijie11/zRPC/config"
	"github.com/zhangweijie11/zRPC/global"
	"github.com/zhangweijie11/zRPC/protocol"
	"log"
	"net"
	"reflect"
	"time"
)

type Client interface {
	Connect(string) error
	Invoke(context.Context, *Service, interface{}, ...interface{}) (interface{}, error)
	Close()
}

type Option struct {
	Retries           int                    // 重试次数
	ConnectionTimeout time.Duration          // 超时时间
	ReadTimeout       time.Duration          // 超时时间
	WriteTimeout      time.Duration          // 超时时间
	SerializeType     protocol.SerializeType // 序列化协议
	CompressType      protocol.CompressType  // 压缩类型
	NetProtocol       string
	FailMode          FailMode
	LoadBalanceMode   LoadBalanceMode
}

var DefaultOption = Option{
	Retries:           3,
	ConnectionTimeout: 5 * time.Second,
	ReadTimeout:       3 * time.Second,
	WriteTimeout:      3 * time.Second,
	SerializeType:     protocol.Gob,
	CompressType:      protocol.None,
	NetProtocol:       "tcp",
	FailMode:          Failover,
	LoadBalanceMode:   RoundRobinBalance,
}

type RPCClient struct {
	conn   net.Conn
	option Option
}

// NewClient 初始化客户端
func NewClient(option Option) *RPCClient {
	return &RPCClient{option: option}
}

// Connect 连接客户端
func (cli *RPCClient) Connect(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, cli.option.ConnectionTimeout)
	if err != nil {
		return err
	}

	cli.conn = conn

	return nil
}

// Invoke 执行
func (cli *RPCClient) Invoke(ctx context.Context, service *Service, stub interface{}, params ...interface{}) (interface{}, error) {
	cli.makeCall(service, stub)

	return cli.wrapCall(ctx, stub, params...)
}

// Close 关闭客户端
func (cli *RPCClient) Close() {
	if cli.conn != nil {
		cli.conn.Close()
	}
}

// 通过反射生成代理函数，在代理函数中完成网络连接、请求数据序列化、网络传输、响应返回数据解析等工作
func (cli *RPCClient) makeCall(service *Service, methodPtr interface{}) {
	container := reflect.ValueOf(methodPtr).Elem()
	// 针对不同序列化协议的编解码器，默认为 GOB 协议
	coder := global.Codecs[cli.option.SerializeType]

	handler := func(req []reflect.Value) []reflect.Value {
		// 函数类型的返回值数量
		numOut := container.Type().NumOut()
		errorHandler := func(err error) []reflect.Value {
			outArgs := make([]reflect.Value, numOut)
			for i := 0; i < len(outArgs); i++ {
				outArgs[i] = reflect.Zero(container.Type().Out(i))
			}
			outArgs[len(outArgs)-1] = reflect.ValueOf(&err).Elem()
			return outArgs
		}

		inArgs := make([]interface{}, 0, len(req))
		for _, arg := range req {
			inArgs = append(inArgs, arg.Interface())
		}

		payload, err := coder.Encode(inArgs)
		if err != nil {
			log.Printf("编码出现异常：%v\n", err)
			return errorHandler(err)
		}

		msg := protocol.NewRPCMsg()
		msg.SetVersion(config.Protocol_MsgVersion)
		msg.SetMsgType(protocol.Request)
		msg.SetCompressType(cli.option.CompressType)
		msg.SetSerializeType(cli.option.SerializeType)
		msg.ServiceClass = service.Class
		msg.ServiceMethod = service.Method
		msg.Payload = payload
		err = msg.Send(cli.conn)
		if err != nil {
			log.Printf("发送数据出现异常：%v\n", err)
			return errorHandler(err)
		}

		respMsg, err := protocol.Read(cli.conn)
		if err != nil {
			return errorHandler(err)
		}

		respDecode := make([]interface{}, 0)
		err = coder.Decode(respMsg.Payload, &respDecode)
		if err != nil {
			log.Printf("解码出现异常：%v\n", err)
			return errorHandler(err)
		}

		if len(respDecode) == 0 {
			respDecode = make([]interface{}, numOut)
		}

		outArgs := make([]reflect.Value, numOut)
		for i := 0; i < numOut; i++ {
			if i != numOut {
				// 如果没有解码到值，设置为与函数返回类型对应位置相同类型的零值
				if respDecode[i] == nil {
					outArgs[i] = reflect.Zero(container.Type().Out(i))
				} else {
					outArgs[i] = reflect.ValueOf(respDecode[i])
				}
			} else {
				outArgs[i] = reflect.Zero(container.Type().Out(i))
			}
		}

		return outArgs
	}
	// 利用反射机制，根据制定的函数类型信息以及处理函数 handler，动态创建一个函数，
	// 并将这个新创建的函数设置到 container 对应的位置，覆盖原始的函数值或指针，实现动态生成的函数替换
	container.Set(reflect.MakeFunc(container.Type(), handler))
}

// 执行实际函数调用
func (cli *RPCClient) wrapCall(ctx context.Context, stub interface{}, params ...interface{}) (interface{}, error) {
	f := reflect.ValueOf(stub).Elem()
	// 判断参数的数量和函数定义的输入参数数量是否相同
	if len(params) != f.Type().NumIn() {
		return nil, errors.New(fmt.Sprintf("参数数量不一致：%d-%d", len(params), f.Type().NumIn()))
	}

	inArgs := make([]reflect.Value, len(params))
	for idx, param := range params {
		inArgs[idx] = reflect.ValueOf(param)
	}
	result := f.Call(inArgs)

	return result, nil
}
