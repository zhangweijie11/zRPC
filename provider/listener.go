package provider

import (
	"fmt"
	"github.com/zhangweijie11/zRPC/config"
	"github.com/zhangweijie11/zRPC/global"
	"github.com/zhangweijie11/zRPC/protocol"
	"io"
	"log"
	"net"
)

type Listener interface {
	Run()
	SetHandler(string, Handler)
	Close()
}

// RPCListener RPC 服务监听器
type RPCListener struct {
	ServiceIP   string
	ServicePort int
	Handlers    map[string]Handler
	netListener net.Listener
}

// NewRPCListener 初始化监听器
func NewRPCListener(serviceIP string, servicePort int) *RPCListener {
	return &RPCListener{
		ServiceIP:   serviceIP,
		ServicePort: servicePort,
		Handlers:    make(map[string]Handler),
		netListener: nil,
	}
}

// Run 启动监听器
func (rl *RPCListener) Run() {
	addr := fmt.Sprintf("%s:%d", rl.ServiceIP, rl.ServicePort)

	netListener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	rl.netListener = netListener

	for {
		conn, err := rl.netListener.Accept()
		if err != nil {
			continue
		}
		go rl.handleConn(conn)
	}
}

// Close 关闭监听器
func (rl *RPCListener) Close() {
	if rl.netListener != nil {
		rl.netListener.Close()
	}
}

// SetHandler 设置处理器
func (rl *RPCListener) SetHandler(name string, handler Handler) {
	if _, ok := rl.Handlers[name]; ok {
		log.Printf("%s 已经注册！", name)
		return
	}

	rl.Handlers[name] = handler
}

// CloseConn 关闭服务链接
func (rl *RPCListener) CloseConn(conn net.Conn) {
	//activeconn
	conn.Close()

	//plugin
	log.Println("服务关闭！")
}

// 处理服务链接
func (rl *RPCListener) handleConn(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("服务 %s 异常r:%s\n", conn.RemoteAddr(), err)
		}
		rl.CloseConn(conn)
	}()
	for {
		msg, err := rl.receiveData(conn)
		if err != nil || msg == nil {
			return
		}

		coder := global.Codecs[msg.Header.SerializeType()]
		if coder == nil {
			return
		}
		inArgs := make([]interface{}, 0)
		err = coder.Decode(msg.Payload, &inArgs)
		if err != nil {
			return
		}
		handler, ok := rl.Handlers[msg.ServiceClass]
		if !ok {
			return
		}
		result, err := handler.Handle(msg.ServiceMethod, inArgs)
		encodeRes, err := coder.Encode(result)
		if err != nil {
			return
		}
		err = rl.sendData(conn, encodeRes)
		if err != nil {
			return
		}
	}
}

// 接收数据
func (rl *RPCListener) receiveData(conn net.Conn) (*protocol.RPCMsg, error) {
	msg, err := protocol.Read(conn)
	if err != nil {
		if err != io.EOF {
			return nil, err
		}
	}
	return msg, nil
}

// 发送数据
func (rl *RPCListener) sendData(conn net.Conn, payload []byte) error {
	resMsg := protocol.NewRPCMsg()
	resMsg.SetVersion(config.Protocol_MsgVersion)
	resMsg.SetMsgType(protocol.Response)
	resMsg.SetCompressType(protocol.None)
	resMsg.SetSerializeType(protocol.Gob)
	resMsg.Payload = payload
	return resMsg.Send(conn)
}
