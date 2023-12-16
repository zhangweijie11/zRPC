package consumer

import (
	"errors"
	"strings"
)

type Service struct {
	AppID  string
	Class  string
	Method string
	Addrs  []string
}

// NewService 初始化服务
func NewService(servicePath string) (*Service, error) {
	arr := strings.Split(servicePath, ".")
	service := &Service{}
	if len(arr) != 3 {
		return service, errors.New("服务路径不可用！")
	}
	service.AppID = arr[0]
	service.Class = arr[1]
	service.Method = arr[2]

	return service, nil
}

func (s *Service) SelectAddr() string {
	return "10.100.40.18:8080"
}
