package provider

import (
	"fmt"
	"github.com/zhangweijie11/zRPC/config"
	"github.com/zhangweijie11/zRPC/global"
	"github.com/zhangweijie11/zRPC/protocol"
	"io"
	"log"
	"net"
	"sync/atomic"
)

type Listener interface {
	Run()
	SetHandler(string, Handler)
	Close()
	GetAddrs() []string
	Shutdown()
}

// RPCListener RPC 服务监听器
type RPCListener struct {
	ServiceIP   string
	ServicePort int
	Handlers    map[string]Handler
	netListener net.Listener
	doneChan    chan struct{}
	shutdown    int32 // 关闭处理中标识位
	handlingNum int32 // 处理中任务数
}

// NewRPCListener 初始化监听器
func NewRPCListener(option Option) *RPCListener {
	return &RPCListener{
		ServiceIP:   option.Ip,
		ServicePort: option.Port,
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

	go rl.acceptConn()

	//for {
	//	conn, err := rl.netListener.Accept()
	//	if err != nil {
	//		continue
	//	}
	//	go rl.handleConn(conn)
	//}
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
	// 关闭挡板
	if rl.isShutdown() {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			log.Printf("服务 %s 异常r:%s\n", conn.RemoteAddr(), err)
		}
		rl.CloseConn(conn)
	}()
	for {

		//关闭挡板
		if rl.isShutdown() {
			return
		}

		//处理中任务数+1
		atomic.AddInt32(&rl.handlingNum, 1)
		//任意退出都会导致处理中任务数-1
		defer atomic.AddInt32(&rl.handlingNum, -1)

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

// GetAddrs 获取监听地址
func (rl *RPCListener) GetAddrs() []string {
	//l.nl.Addr()
	addr := fmt.Sprintf("tcp://%s:%d", rl.ServiceIP, rl.ServicePort)
	return []string{addr}
}

func (rl *RPCListener) acceptConn() {
	for {
		conn, err := rl.netListener.Accept()
		if err != nil {
			select {
			case <-rl.getDoneChan(): //挡板：server closed done
				return
			default:
			}
			return
		}
		go rl.handleConn(conn) //处理连接
	}
}

func (rl *RPCListener) getDoneChan() <-chan struct{} {
	return rl.doneChan
}

// 关闭通道
func (rl *RPCListener) closeDoneChan() {
	select {
	case <-rl.doneChan:
	default:
		close(rl.doneChan)
	}
}

// 判断服务是否关闭
func (rl *RPCListener) isShutdown() bool {
	return atomic.LoadInt32(&rl.shutdown) == 1
}

// Shutdown 关闭逻辑
func (rl *RPCListener) Shutdown() {
	atomic.CompareAndSwapInt32(&rl.shutdown, 0, 1)
	for {
		if atomic.LoadInt32(&rl.handlingNum) == 0 {
			break
		}
	}
	rl.closeDoneChan()
}
