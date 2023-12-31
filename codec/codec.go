package codec

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

// Codec 编码解码器
type Codec interface {
	Encode(i interface{}) ([]byte, error)
	Decode(data []byte, i interface{}) error
}

type GobCodec struct{}

// Encode 编码，针对 GOB 协议
func (c *GobCodec) Encode(i interface{}) ([]byte, error) {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)

	if err := encoder.Encode(i); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Decode 解码，针对 GOB 协议
func (c *GobCodec) Decode(data []byte, i interface{}) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	return decoder.Decode(i)
}

type JSONCodec struct{}

// Encode 编码，针对 JSON 协议
func (c *JSONCodec) Encode(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

// Decode 解码，针对 JSON 协议
func (c *JSONCodec) Decode(data []byte, i interface{}) error {
	decode := json.NewDecoder(bytes.NewBuffer(data))
	return decode.Decode(i)
}
