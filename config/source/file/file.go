package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/CloudRoamer/aimo-libs/config"
	"github.com/CloudRoamer/aimo-libs/config/codec"
)

const (
	// DefaultPriority 文件配置优先级
	DefaultPriority = 60
)

// Source 文件配置源
type Source struct {
	path     string      // 配置文件路径
	codec    codec.Codec // 编解码器
	priority int         // 优先级

	mu      sync.RWMutex
	watcher *watcher
}

// Option 配置选项
type Option func(*Source)

// WithCodec 设置编解码器
func WithCodec(c codec.Codec) Option {
	return func(s *Source) {
		s.codec = c
	}
}

// WithPriority 设置优先级
func WithPriority(p int) Option {
	return func(s *Source) {
		s.priority = p
	}
}

// New 创建文件配置源
func New(path string, opts ...Option) (*Source, error) {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	s := &Source{
		path:     path,
		priority: DefaultPriority,
	}

	// 根据文件扩展名选择默认编解码器
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		s.codec = codec.JSON
	case ".yaml", ".yml":
		s.codec = codec.YAML
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

func (s *Source) Name() string {
	return "file:" + s.path
}

func (s *Source) Priority() int {
	return s.priority
}

func (s *Source) Load(ctx context.Context) (map[string]config.Value, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解码为 map
	var raw map[string]any
	if err := s.codec.Decode(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// 扁平化嵌套结构
	result := make(map[string]config.Value)
	flatten("", raw, result)

	return result, nil
}

// flatten 将嵌套 map 扁平化为点分隔的 key
func flatten(prefix string, data map[string]any, result map[string]config.Value) {
	for k, v := range data {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]any:
			flatten(key, val, result)
		case map[any]any:
			// YAML 可能返回这种类型
			converted := make(map[string]any)
			for mk, mv := range val {
				converted[fmt.Sprintf("%v", mk)] = mv
			}
			flatten(key, converted, result)
		default:
			result[key] = config.NewValueFromInterface(v)
		}
	}
}

// Watch 返回文件监听器（本阶段暂不实现）
func (s *Source) Watch() config.Watcher {
	return nil
}
