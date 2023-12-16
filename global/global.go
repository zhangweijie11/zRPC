package global

import (
	"fmt"
	"github.com/zhangweijie11/zRPC/codec"
	"github.com/zhangweijie11/zRPC/protocol"
)

var Codecs = map[protocol.SerializeType]codec.Codec{
	protocol.JSON: &codec.JSONCodec{},
	protocol.Gob:  &codec.GobCodec{},
}

type TestHandler struct{}

func (h *TestHandler) Hello() string {
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

func (h *UserHandler) GetUserByID(id int) (User, error) {
	if u, ok := userList[id]; ok {
		return u, nil
	}

	return User{}, fmt.Errorf("id %d 不存在！", id)
}
