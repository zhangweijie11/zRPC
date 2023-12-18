package consumer

import (
	"context"
	"errors"
	"github.com/zhangweijie11/zRPC/naming"
	"sync"
)

type RPCClientProxy struct {
	option      Option
	registry    naming.Registry
	failMode    FailMode
	mutex       sync.RWMutex
	servers     []string
	loadBalance LoadBalance
	client      Client
}

func (cp *RPCClientProxy) Call(ctx context.Context, servicePath string, stub interface{}, params ...interface{}) (interface{}, error) {
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

// 获取服务列表
func (cp *RPCClientProxy) discoveryService(ctx context.Context, appId string) ([]string, error) {
	instances, ok := cp.registry.Fetch(ctx, appId)
	if !ok {
		return nil, errors.New("service not found")
	}
	var servers []string
	for _, instance := range instances {
		servers = append(servers, instance.Addresses...)
	}
	return servers, nil
}

func NewRPCClientProxy(appId string, option Option, registry naming.Registry) *RPCClientProxy {
	rcp := &RPCClientProxy{option: option, failMode: option.FailMode, registry: registry}
	servers, err := rcp.discoveryService(context.Background(), appId)
	if err != nil {
		panic(err)
	}

	rcp.servers = servers
	rcp.loadBalance = LoadBalanceFactory(option.LoadBalanceMode, rcp.servers)
	rcp.client = NewClient(rcp.option)

	return rcp
}
