package provider

import (
	"context"
	"errors"
	"github.com/zhangweijie11/zRPC/naming"
	"log"
	"reflect"
	"time"
)

var maxRegisterRetry int = 2

// Server 服务接口提供服务启停和处理方法注册
type Server interface {
	Register(string, interface{})
	Run()
	Close()
	Shutdown()
}

type Option struct {
	Ip           string
	Port         int
	Hostname     string
	AppID        string
	Env          string
	NetProtocol  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

var DefaultOption = Option{
	NetProtocol:  "tcp",
	ReadTimeout:  5 * time.Second,
	WriteTimeout: 5 * time.Second,
}

type RPCServer struct {
	listener   Listener
	registry   naming.Registry
	option     Option
	cancelFunc context.CancelFunc
}

// NewRPCServer 初始化 RPC 服务
func NewRPCServer(option Option, registry naming.Registry) *RPCServer {
	return &RPCServer{listener: NewRPCListener(option), registry: registry, option: option}
}

// Run 启动服务
func (rs *RPCServer) Run() {
	rs.listener.Run()
	//if err != nil {
	//	panic(err)
	//}

	err := rs.registerToNaming()
	if err != nil {
		// 注册失败关闭服务
		rs.Close()
		panic(err)
	}
}

func (rs *RPCServer) registerToNaming() error {
	instance := &naming.Instance{
		Env:       rs.option.Env,
		AppID:     rs.option.AppID,
		Hostname:  rs.option.Hostname,
		Addresses: rs.listener.GetAddrs(),
	}
	retries := maxRegisterRetry
	for retries > 0 {
		retries--
		cancel, err := rs.registry.Register(context.Background(), instance)
		if err == nil {
			rs.cancelFunc = cancel
			return nil
		}
	}
	return errors.New("register to naming server fail")
}

// Close 关闭服务
func (rs *RPCServer) Close() {
	// 从服务注册中心注销
	if rs.cancelFunc != nil {
		rs.cancelFunc()
	}

	// 关闭当前服务
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

// Shutdown 注销服务
func (rs *RPCServer) Shutdown() {
	// 从服务注册中心注销
	if rs.cancelFunc != nil {
		rs.cancelFunc()
	}

	// 关闭当前服务
	if rs.listener != nil {
		rs.listener.Shutdown()
	}
}
