package codec

import "gopkg.in/yaml.v3"

var (
	// YAML 全局 YAML 编解码器实例
	YAML Codec = &yamlCodec{}
)

type yamlCodec struct{}

func (c *yamlCodec) Encode(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

func (c *yamlCodec) Decode(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}
