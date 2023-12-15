package provider

import "reflect"

type Handler interface {
	Handle(string, []interface{}) ([]interface{}, error)
}

type RPCServerHandler struct {
	rpcServer *RPCServer
	class     reflect.Value
}

func (handler *RPCServerHandler) Handle(method string, params []interface{}) ([]interface{}, error) {
	args := make([]reflect.Value, len(params))
	for i := range args {
		args[i] = reflect.ValueOf(params[i])
	}

	reflectMethod := handler.class.MethodByName(method)

	result := reflectMethod.Call(args)

	resArgs := make([]interface{}, len(result))
	for i := 0; i < len(result); i++ {
		resArgs[i] = result[i].Interface()
	}

	var err error

	if _, ok := result[len(result)-1].Interface().(error); ok {
		err = result[len(result)-1].Interface().(error)
	}

	return resArgs, err
}
