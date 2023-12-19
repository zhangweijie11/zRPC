package main

import (
	"context"
	"encoding/gob"
	"github.com/zhangweijie11/zRPC/consumer"
	"github.com/zhangweijie11/zRPC/global"
	"github.com/zhangweijie11/zRPC/naming"
	"log"
	"reflect"
)

func main() {
	nodes := []string{"localhost:8881"}
	conf := &naming.Config{Nodes: nodes, Env: "dev"}
	discovery := naming.NewDiscovery(conf)

	gob.Register(global.User{})
	cli := consumer.NewRPCClientProxy("UserService", consumer.DefaultOption, discovery)

	var GetUserById func(id int) (global.User, error)

	//wrap call
	ret, err := cli.Call(context.Background(), "User.GetUserById", &GetUserById, 1)
	if err != nil {
		log.Println("call error:", err)
	} else {
		val := ret.([]reflect.Value)
		user := val[0].Interface().(global.User)
		log.Println("rpc return result:", user)
	}

	//makefunc and call
	//u, err := GetUserById(2)

	/*var Hello func() string
	cli.Call(ctx, "Test.Hello", &Hello)
	r := Hello()
	log.Println("result:", r, err)

	var Add func(a, b int) int
	cli.Call(ctx, "Test.Add", &Add)
	w := Add(1, 2)
	log.Println("result:", w)

	var Login func(string, string) bool
	cli.Call(ctx, "User.Login", &Login)
	v := Login("kavin", "123456")
	log.Println("result:", v)
	*/
}
