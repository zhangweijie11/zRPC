package global

import (
	"github.com/zhangweijie11/zRPC/codec"
	"github.com/zhangweijie11/zRPC/protocol"
)

var Codecs = map[protocol.SerializeType]codec.Codec{
	protocol.JSON: &codec.JSONCodec{},
	protocol.Gob:  &codec.GobCodec{},
}
