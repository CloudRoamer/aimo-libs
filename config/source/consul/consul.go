package consul

import (
	"context"
	"fmt"
	"strings"
	"sync"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/CloudRoamer/aimo-libs/config"
)

const (
	DefaultPriority = 80
	DefaultPrefix   = "config"
)

// Source Consul KV 配置源
type Source struct {
	client    *consulapi.Client
	prefix    string
	priority  int
	separator string

	mu      sync.RWMutex
	watcher *watcher
}

// New 创建 Consul KV 配置源
func New(address string, opts ...Option) (*Source, error) {
	cfg := consulapi.DefaultConfig()
	cfg.Address = address

	client, err := consulapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	// 验证连接
	_, err = client.Status().Leader()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to consul at %s: %w", address, err)
	}

	s := &Source{
		client:    client,
		prefix:    DefaultPrefix,
		priority:  DefaultPriority,
		separator: "/",
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

func (s *Source) Name() string {
	return "consul"
}

func (s *Source) Priority() int {
	return s.priority
}

func (s *Source) Load(ctx context.Context) (map[string]config.Value, error) {
	kv := s.client.KV()

	pairs, _, err := kv.List(s.prefix, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys from consul: %w", err)
	}

	result := make(map[string]config.Value)
	prefixWithSep := s.prefix + "/"

	for _, pair := range pairs {
		// 移除前缀并转换分隔符
		key := strings.TrimPrefix(pair.Key, prefixWithSep)
		if key == "" || key == pair.Key {
			// 跳过前缀本身或未匹配前缀的 key
			continue
		}

		// 将路径分隔符转换为点分隔符
		key = strings.ReplaceAll(key, "/", ".")

		result[key] = config.NewValue(string(pair.Value))
	}

	return result, nil
}

// Watch 返回 Consul 监听器
func (s *Source) Watch() config.Watcher {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.watcher == nil {
		s.watcher = newWatcher(s.client, s.prefix)
	}
	return s.watcher
}
