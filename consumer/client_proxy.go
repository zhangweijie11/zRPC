package consumer

import (
	"context"
	"errors"
)

type RPCClientProxy struct {
	option Option
}

func (cp RPCClientProxy) Call(ctx context.Context, servicePath string, stub interface{}, params ...interface{}) (interface{}, error) {
	service, err := NewService(servicePath)
	if err != nil {
		return nil, err
	}

	client := NewClient(cp.option)
	addr := service.SelectAddr()
	// TODO: 长连接管理
	err = client.Connect(addr)
	if err != nil {
		return nil, err
	}

	retries := cp.option.Retries
	for retries > 0 {
		retries--
		return client.Invoke(ctx, service, stub, params...)
	}

	return nil, errors.New("error")
}

func NewRPCClientProxy(option Option) *RPCClientProxy {
	return &RPCClientProxy{option: option}
}
