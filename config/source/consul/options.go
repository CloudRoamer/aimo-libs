package consul

import (
	"strings"

	consulapi "github.com/hashicorp/consul/api"
)

// Option Consul 配置源选项
type Option func(*Source)

// WithPrefix 设置 KV 路径前缀
// 例如: "config/prod/myapp"
func WithPrefix(prefix string) Option {
	return func(s *Source) {
		s.prefix = strings.TrimSuffix(prefix, "/")
	}
}

// WithPriority 设置优先级
func WithPriority(p int) Option {
	return func(s *Source) {
		s.priority = p
	}
}

// WithToken 设置 ACL Token
// 注意：需要在创建 Source 前通过 consulapi.Config 设置
func WithToken(token string) Option {
	return func(s *Source) {
		// Consul client 的 config 字段是私有的
		// Token 应该在创建 client 前通过 consulapi.Config 设置
		// 这里保留为空实现，实际使用时应该通过 WithConfig 选项
	}
}

// WithConfig 使用自定义 Consul 配置
func WithConfig(cfg *consulapi.Config) Option {
	return func(s *Source) {
		client, err := consulapi.NewClient(cfg)
		if err == nil {
			s.client = client
		}
	}
}

// WithSeparator 设置路径分隔符（默认为 "/"）
func WithSeparator(sep string) Option {
	return func(s *Source) {
		s.separator = sep
	}
}
