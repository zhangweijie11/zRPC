package main

import (
	"encoding/gob"
	"github.com/zhangweijie11/zRPC/global"
	"github.com/zhangweijie11/zRPC/naming"
	"github.com/zhangweijie11/zRPC/provider"
	"gopkg.in/yaml.v3"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	Hostname      string   `yaml:"hostname"`
	Appid         string   `yaml:"appid"`
	Port          int      `yaml:"port"`
	Ip            string   `yaml:"ip"`
	Env           string   `yaml:"env"`
	RegistryAddrs []string `yaml:"registry_addrs"`
}

func loadConfig(path string) (*Config, error) {
	configFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := new(Config)
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func main() {
	config, err := loadConfig("config.yaml")
	if err != nil {
		panic("读取配置文件出错！")
	}
	conf := &naming.Config{
		Nodes: config.RegistryAddrs,
		Env:   config.Env,
	}
	discovery := naming.NewDiscovery(conf)
	option := provider.Option{
		Ip:       config.Ip,
		Port:     config.Port,
		Hostname: config.Hostname,
		Env:      config.Env,
		AppID:    config.Appid,
	}
	rpcServer := provider.NewRPCServer(option, discovery)
	rpcServer.RegisterName("User", &global.UserHandler{})
	rpcServer.RegisterName("Hello", &global.HelloHandler{})
	// 可以在 RPC 传输中序列化和反序列化数据
	gob.Register(global.User{})

	go rpcServer.Run()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-quit
	rpcServer.Shutdown()
}
