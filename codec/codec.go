package codec

import (
	"bytes"
	"encoding/gob"
)

// Codec 编码解码器
type Codec interface {
	Encode(i interface{}) ([]byte, error)
	Decode(data []byte, i interface{}) error
}

type GobCodec struct{}

// Encode 编码
func (c GobCodec) Encode(i interface{}) ([]byte, error) {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)

	if err := encoder.Encode(i); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Decode 解码
func (c GobCodec) Decode(data []byte, i interface{}) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	return decoder.Decode(i)
}
