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
	SerializeType     protocol.SerializeType // 序列化协议
	CompressType      protocol.CompressType  // 压缩类型
}

var DefaultOption = Option{
	Retries:           3,
	ConnectionTimeout: 5 * time.Second,
	SerializeType:     protocol.Gob,
	CompressType:      protocol.None,
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
	// 针对不同序列化协议的编解码器
	coder := global.Codecs[cli.option.SerializeType]

	handler := func(req []reflect.Value) []reflect.Value {
		// 函数的输出参数计数
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
	container.Set(reflect.MakeFunc(container.Type(), handler))
}

// 执行实际函数调用
func (cli RPCClient) wrapCall(ctx context.Context, stub interface{}, params ...interface{}) (interface{}, error) {
	f := reflect.ValueOf(stub).Elem()
	if len(params) != f.Type().NumIn() {
		return nil, errors.New(fmt.Sprintf("参数无法使用：%d-%d", len(params), f.Type().NumIn()))
	}

	in := make([]reflect.Value, len(params))
	for idx, param := range params {
		in[idx] = reflect.ValueOf(param)
	}
	result := f.Call(in)

	return result, nil
}
