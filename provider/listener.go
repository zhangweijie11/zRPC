package provider

import (
	"fmt"
	"log"
	"net"
)

type Listener interface {
	Run()
	SetHandler(string, Handler)
	Close()
}

type RPCListener struct {
	ServiceIP   string
	ServicePort int
	Handlers    map[string]Handler
	netListener net.Listener
}

func NewRPCListener(serviceIP string, servicePort int) *RPCListener {
	return &RPCListener{
		ServiceIP:   serviceIP,
		ServicePort: servicePort,
		Handlers:    make(map[string]Handler),
		netListener: nil,
	}
}

func (rl *RPCListener) Run() {
	addr := fmt.Sprintf("%s:%d", rl.ServiceIP, rl.ServicePort)

	netListener, err := net.Listen("config.NET_TRANS_PROTOCOL", addr)
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

func (rl *RPCListener) Close() {
	if rl.netListener != nil {
		rl.netListener.Close()
	}
}

func (rl *RPCListener) SetHandler(name string, handler Handler) {
	if _, ok := rl.Handlers[name]; ok {
		log.Printf("%s 已经注册！", name)
		return
	}

	rl.Handlers[name] = handler
}
