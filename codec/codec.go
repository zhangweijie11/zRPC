package codec

import (
	"bytes"
	"encoding/gob"
)

type Codec interface {
	Encode(i interface{}) ([]byte, error)
	Decode(data []byte, i interface{}) error
}

type CobCodec struct{}

// Encode 编码
func (c CobCodec) Encode(i interface{}) ([]byte, error) {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)

	if err := encoder.Encode(i); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Decode 解码
func (c CobCodec) Decode(data []byte, i interface{}) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	return decoder.Decode(i)
}
