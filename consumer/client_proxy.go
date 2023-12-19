package consumer

import (
	"context"
	"errors"
	"github.com/zhangweijie11/zRPC/naming"
	"strings"
	"sync"
)

type ClientProxy interface {
	Call(context.Context, string, interface{}, ...interface{}) (interface{}, error)
}

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

	err = cp.getConn()
	// 如果获取链接出现问题，并且接受失败，直接返回
	if err != nil && cp.failMode == Failfast {
		return nil, err
	}

	switch cp.failMode {
	case Failretry:
		retries := cp.option.Retries
		for retries > 0 {
			retries--
			if cp.client != nil {
				result, err := cp.client.Invoke(ctx, service, stub, params...)
				if err == nil {
					return result, err
				}
			}
		}
	case Failover:
		retries := cp.option.Retries
		for retries > 0 {
			retries--
			if cp.client != nil {
				result, err := cp.client.Invoke(ctx, service, stub, params...)
				if err == nil {
					return result, err
				}
			}
			err = cp.getConn()
		}
	case Failfast:
		if cp.client != nil {
			result, err := cp.client.Invoke(ctx, service, stub, params...)
			if err == nil {
				return result, nil
			}
			return nil, err
		}
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

func NewRPCClientProxy(appId string, option Option, registry naming.Registry) ClientProxy {
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

func (cp *RPCClientProxy) getConn() error {
	addr := strings.Replace(cp.loadBalance.Get(), cp.option.NetProtocol+"://", "", -1)
	err := cp.client.Connect(addr) //长连接管理
	if err != nil {
		return err
	}
	return nil
}
