package main

import (
	"context"
	"encoding/gob"
	"github.com/zhangweijie11/zRPC/consumer"
	"github.com/zhangweijie11/zRPC/global"
	"log"
)

func main() {
	gob.Register(global.User{})

	cli := consumer.NewRPCClientProxy(consumer.DefaultOption)
	ctx := context.Background()
	var GetUserByID func(id int) (global.User, error)
	cli.Call(ctx, "UserService.User.GetUserByID", &GetUserByID)
	user, err := GetUserByID(2)
	log.Println("结果：", user, err)
	var Hello func() string
	result, err := cli.Call(ctx, "UserService.Test.Hello", &Hello)
	log.Println("结果：", result, err)
}
