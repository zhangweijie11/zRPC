package protocol

/*
RPC 消息格式编码设计：
协议消息头定义定长 5 字节，依次放置
魔术数（用于校验），
协议版本，
消息类型（区分请求和响应），
压缩类型，
序列化协议类型，
每个占 1 个字节（8 个 bit）。可扩展追加消息 ID 以及元数据等信息用于服务治理
*/

const (
	// 消息头长度
	HeaderLen = 5
)

const (
	// 魔术数（用于校验）
	magicNumber byte = 0x06
)

// 消息类型
type MsgType byte

const (
	// 消息类型
	Request MsgType = iota
	Response
)

type CompressType byte

const (
	// 压缩类型
	None CompressType = iota
	Gzip
)

type SerializeType byte

const (
	// 序列化类型
	Gob SerializeType = iota
	JSON
)

type Header [HeaderLen]byte

func (h *Header) CheckMagicNumber() bool {
	return h[0] == magicNumber
}

func (h *Header) Version() byte {
	return h[1]
}

func (h *Header) SetVersion(version byte) {
	h[1] = version
}

func (h *Header) MsgType() MsgType {
	return MsgType(h[2])
}

func (h *Header) SetMsgType(msgType MsgType) {
	h[2] = byte(msgType)
}

func (h *Header) CompressType() CompressType {
	return CompressType(h[3])
}

func (h *Header) SetCompressType(compressType CompressType) {
	h[3] = byte(compressType)
}

func (h *Header) SerializeType() SerializeType {
	return SerializeType(h[4])
}

func (h *Header) SetSerializeType(serializerType SerializeType) {
	h[4] = byte(serializerType)
}
