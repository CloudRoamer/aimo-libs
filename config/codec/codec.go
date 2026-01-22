package codec

// Codec 定义编解码器接口
type Codec interface {
	// Encode 将结构体编码为字节流
	Encode(v interface{}) ([]byte, error)

	// Decode 将字节流解码为结构体
	Decode(data []byte, v interface{}) error
}
