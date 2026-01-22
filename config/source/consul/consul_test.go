package consul

import (
	"context"
	"testing"
	"time"
)

// TestNew 测试创建 Consul 配置源
func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		address string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "invalid address",
			address: "invalid://address",
			wantErr: true,
		},
		{
			name:    "with prefix option",
			address: "localhost:8500",
			opts:    []Option{WithPrefix("config/test")},
			wantErr: false,
		},
		{
			name:    "with priority option",
			address: "localhost:8500",
			opts:    []Option{WithPriority(90)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.address, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSource_Name 测试配置源名称
func TestSource_Name(t *testing.T) {
	s := &Source{
		addr:   "localhost:8500",
		prefix: "config",
	}
	want := "consul:localhost:8500/config"
	if got := s.Name(); got != want {
		t.Errorf("Name() = %v, want %v", got, want)
	}
}

// TestSource_Priority 测试优先级
func TestSource_Priority(t *testing.T) {
	tests := []struct {
		name     string
		priority int
		want     int
	}{
		{
			name:     "default priority",
			priority: DefaultPriority,
			want:     DefaultPriority,
		},
		{
			name:     "custom priority",
			priority: 90,
			want:     90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Source{priority: tt.priority}
			if got := s.Priority(); got != tt.want {
				t.Errorf("Priority() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSource_Load 测试加载配置
// 注意：此测试需要运行中的 Consul 服务器
func TestSource_Load(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// 创建配置源
	source, err := New("localhost:8500", WithPrefix("config/test"))
	if err != nil {
		t.Skipf("Consul not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 加载配置
	values, err := source.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// 验证返回的是 map
	if values == nil {
		t.Error("Load() returned nil map")
	}
}

// TestSource_Watch 测试 Watch 功能
func TestSource_Watch(t *testing.T) {
	s := &Source{prefix: "test"}

	// 第一次调用应该创建 watcher
	w1 := s.Watch()
	if w1 == nil {
		t.Error("Watch() returned nil")
	}

	// 第二次调用应该返回同一个 watcher
	w2 := s.Watch()
	if w1 != w2 {
		t.Error("Watch() should return the same watcher instance")
	}
}

// TestWithPrefix 测试前缀选项
func TestWithPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{
			name:   "simple prefix",
			prefix: "config",
			want:   "config",
		},
		{
			name:   "prefix with trailing slash",
			prefix: "config/prod/",
			want:   "config/prod",
		},
		{
			name:   "nested prefix",
			prefix: "config/prod/myapp",
			want:   "config/prod/myapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Source{}
			opt := WithPrefix(tt.prefix)
			opt(s)
			if s.prefix != tt.want {
				t.Errorf("WithPrefix() prefix = %v, want %v", s.prefix, tt.want)
			}
		})
	}
}
