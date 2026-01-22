package codec

import "encoding/json"

var (
	// JSON 全局 JSON 编解码器实例
	JSON Codec = &jsonCodec{}
)

type jsonCodec struct{}

func (c *jsonCodec) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *jsonCodec) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
