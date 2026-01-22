package codec

import "encoding/json"

var (
	// JSON 全局 JSON 编解码器实例
	JSON Codec = &jsonCodec{}
)

type jsonCodec struct{}

func (c *jsonCodec) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (c *jsonCodec) Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
