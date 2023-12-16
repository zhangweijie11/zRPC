package protocol

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	// SplitLen 代表各部分长度，是 int32 类型（32bit），也就是 4 个字节，所以为 4
	SplitLen = 4
)

// 协议消息格式
type RPCMsg struct {
	*Header              // 协议头
	ServiceClass  string // 调用的服务类名
	ServiceMethod string // 调用的方法名
	Payload       []byte // 调用的参数
}

// NewRPCMsg 初始化消息格式
func NewRPCMsg() *RPCMsg {
	header := Header([HeaderLen]byte{})
	header[0] = magicNumber
	return &RPCMsg{Header: &header}
}

// Send 发送数据，数据格式为：协议头，总体长度，类名长度，类名，方法名长度，方法，参数长度，参数
func (msg *RPCMsg) Send(writer io.Writer) error {
	// 写入协议头
	_, err := writer.Write(msg.Header[:])
	if err != nil {
		return err
	}
	// 消息体总长度，方便一次性解析
	dataLen := SplitLen + len(msg.ServiceClass) + SplitLen + len(msg.ServiceMethod) + SplitLen + len(msg.Payload)
	// 网络传输一般使用大端字节序，字节序即为字节的组成顺序，分为大端序（最高有效位放低地址）和小端序（最低有效位放低地址），
	// CPU 一般采用小端序读写，TCP 网络传输一般采用大端序更为方便， binary.BigEndian 代码实现大端序
	// 写入消息体长度
	err = binary.Write(writer, binary.BigEndian, uint32(dataLen))

	// 写入调用的服务类名长度
	err = binary.Write(writer, binary.BigEndian, uint32(len(msg.ServiceClass)))
	if err != nil {
		return err
	}
	// 写入调用的服务类名
	err = binary.Write(writer, binary.BigEndian, []byte(msg.ServiceClass))
	if err != nil {
		return err
	}

	// 写入调用的服务方法名长度
	err = binary.Write(writer, binary.BigEndian, uint32(len(msg.ServiceMethod)))
	if err != nil {
		return err
	}
	// 写入调用的服务方法名
	err = binary.Write(writer, binary.BigEndian, []byte(msg.ServiceMethod))
	if err != nil {
		return err
	}

	// 写入调用的服务参数长度
	err = binary.Write(writer, binary.BigEndian, uint32(len(msg.Payload)))
	if err != nil {
		return err
	}
	// 写入调用的服务参数
	err = binary.Write(writer, binary.BigEndian, msg.Payload)
	if err != nil {
		return err
	}

	return err
}

// Decode 解码
func (msg *RPCMsg) Decode(r io.Reader) error {
	// 读取协议头
	_, err := io.ReadFull(r, msg.Header[:])
	if !msg.Header.CheckMagicNumber() {
		return errors.New("校验值错误！")
	}

	// 消息体长度
	headerByte := make([]byte, 4)
	_, err = io.ReadFull(r, headerByte)
	if err != nil {
		return err
	}

	// 获取消息体长度
	bodyLen := binary.BigEndian.Uint32(headerByte)
	// 一次性获取整个消息体，再依次拆解
	data := make([]byte, bodyLen)
	_, err = io.ReadFull(r, data)

	// 调用的服务类名长度
	start := 0
	end := start + SplitLen
	classLen := binary.BigEndian.Uint32(data[start:end])

	// 调用的服务类名
	start = end
	end = start + int(classLen)
	msg.ServiceClass = string(data[start:end])

	// 调用的方法名长度
	start = end
	end = start + SplitLen
	methodLen := binary.BigEndian.Uint32(data[start:end])

	// 调用的方法
	start = end
	end = start + int(methodLen)
	msg.ServiceMethod = string(data[start:end])

	// 调用的参数长度
	start = end
	end = start + SplitLen
	binary.BigEndian.Uint32(data[start:end])

	// 调用的参数
	start = end
	msg.Payload = data[start:]

	return err
}

func Read(r io.Reader) (*RPCMsg, error) {
	msg := NewRPCMsg()
	err := msg.Decode(r)
	if err != nil {
		return nil, err
	}

	return msg, err
}
