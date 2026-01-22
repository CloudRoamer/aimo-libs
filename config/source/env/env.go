package env

import (
	"context"
	"os"
	"strings"

	"github.com/CloudRoamer/aimo-libs/config"
)

const (
	// DefaultPriority 环境变量优先级最高
	DefaultPriority = 100
)

// Source 环境变量配置源
type Source struct {
	prefix      string              // 环境变量前缀，如 "APP_"
	separator   string              // key 分隔符，默认 "_"
	priority    int                 // 优先级
	keyMapping  func(string) string // 自定义 key 映射函数
}

// Option 配置选项
type Option func(*Source)

// WithPrefix 设置环境变量前缀
func WithPrefix(prefix string) Option {
	return func(s *Source) {
		s.prefix = prefix
	}
}

// WithSeparator 设置层级分隔符
func WithSeparator(sep string) Option {
	return func(s *Source) {
		s.separator = sep
	}
}

// WithPriority 设置优先级
func WithPriority(p int) Option {
	return func(s *Source) {
		s.priority = p
	}
}

// WithKeyMapping 设置自定义 key 映射函数
// 用于将环境变量名转换为配置 key
// 例如: APP_DATABASE_HOST -> database.host
func WithKeyMapping(fn func(string) string) Option {
	return func(s *Source) {
		s.keyMapping = fn
	}
}

// New 创建环境变量配置源
func New(opts ...Option) *Source {
	s := &Source{
		prefix:    "",
		separator: "_",
		priority:  DefaultPriority,
	}

	for _, opt := range opts {
		opt(s)
	}

	// 默认 key 映射：移除前缀，将分隔符替换为点，转小写
	if s.keyMapping == nil {
		s.keyMapping = func(envKey string) string {
			key := envKey
			if s.prefix != "" {
				key = strings.TrimPrefix(key, s.prefix)
			}
			key = strings.ToLower(key)
			key = strings.ReplaceAll(key, s.separator, ".")
			return key
		}
	}

	return s
}

func (s *Source) Name() string {
	return "env"
}

func (s *Source) Priority() int {
	return s.priority
}

func (s *Source) Load(ctx context.Context) (map[string]config.Value, error) {
	result := make(map[string]config.Value)

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]

		// 如果设置了前缀，只处理匹配前缀的环境变量
		if s.prefix != "" && !strings.HasPrefix(key, s.prefix) {
			continue
		}

		configKey := s.keyMapping(key)
		result[configKey] = config.NewValue(value)
	}

	return result, nil
}

// Watch 环境变量不支持监听
func (s *Source) Watch() config.Watcher {
	return nil
}
