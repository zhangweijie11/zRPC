package provider

import (
	"log"
	"reflect"
)

// Server 服务接口提供服务启停和处理方法注册
type Server interface {
	Register(string, interface{})
	Run()
	Close()
}

type RPCServer struct {
	listener Listener
}

// NewRPCServer 初始化 RPC 服务
func NewRPCServer(ip string, port int) *RPCServer {
	return &RPCServer{listener: NewRPCListener(ip, port)}
}

// Run 启动服务
func (rs *RPCServer) Run() {
	go rs.listener.Run()
}

// Close 关闭服务
func (rs *RPCServer) Close() {
	if rs.listener != nil {
		rs.listener.Close()
	}
}

// Register 注册服务
func (rs *RPCServer) Register(class interface{}) {
	name := reflect.Indirect(reflect.ValueOf(class)).Type().Name()
	rs.RegisterName(name, class)

}

// RegisterName 通过名字注册服务
func (rs *RPCServer) RegisterName(name string, class interface{}) {
	handler := &RPCServerHandler{
		rpcServer: nil,
		class:     reflect.ValueOf(class),
	}
	rs.listener.SetHandler(name, handler)
	log.Printf("%s 注册成功！", name)
}
