package provider

import "reflect"

type Handler interface {
	Handle(string, []interface{}) ([]interface{}, error)
}

type RPCServerHandler struct {
	rpcServer *RPCServer
	class     reflect.Value
}
