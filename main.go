package main

import (
	"encoding/gob"
	"fmt"
	"github.com/zhangweijie11/zRPC/provider"
	"os"
	"os/signal"
	"syscall"
)

type TestHandler struct{}

func (h TestHandler) Hello() string {
	return "hello world"
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var userList = map[int]User{
	1: {
		ID:   1,
		Name: "AAA",
		Age:  11,
	},
	2: {
		ID:   2,
		Name: "BBB",
		Age:  12,
	},
}

type UserHandler struct{}

func (h *UserHandler) GetUserById(id int) (User, error) {
	if u, ok := userList[id]; ok {
		return u, nil
	}

	return User{}, fmt.Errorf("id %d 不存在！", id)
}

func main() {
	ip := "10.99.60.15"
	port := 5002

	rpcServer := provider.NewRPCServer(ip, port)
	rpcServer.RegisterName("User", &UserHandler{})
	rpcServer.RegisterName("Test", &TestHandler{})
	gob.Register(User{})

	go rpcServer.Run()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-quit
	rpcServer.Close()
}
