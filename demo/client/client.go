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
	// 通过 Call 方法获取远程服务的方法，将地址传递给相对应的函数，然后将相对应的函数指向远程服务中的方法
	cli.Call(ctx, "UserService.User.GetUserByID", &GetUserByID)
	user, err := GetUserByID(2)
	log.Println("结果：", user, err)
	var Hello func() string
	cli.Call(ctx, "UserService.Test.Hello", &Hello)
	result := Hello()
	log.Println("结果：", result, err)
}
