package provider

// Server 服务接口提供服务启停和处理方法注册
type Server interface {
	Register(string, interface{})
	Run()
	Close()
}

type RPCServer struct {
	listener Listener
}

func NewRPCServer(ip string, port int) *RPCServer {
	return &RPCServer{listener: NewRPCListener(ip, port)}
}

func (rs *RPCServer) Run() {
	go rs.listener.Run()
}

func (rs *RPCServer) Close() {
	if rs.listener != nil {
		rs.listener.Close()
	}
}
