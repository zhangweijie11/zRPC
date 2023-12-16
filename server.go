package main

import (
	"encoding/gob"
	"github.com/zhangweijie11/zRPC/global"
	"github.com/zhangweijie11/zRPC/provider"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ip := "10.100.40.18"
	port := 8080

	rpcServer := provider.NewRPCServer(ip, port)
	rpcServer.RegisterName("User", &global.UserHandler{})
	rpcServer.RegisterName("Test", &global.TestHandler{})
	gob.Register(global.User{})

	go rpcServer.Run()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-quit
	rpcServer.Close()
}
