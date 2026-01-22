package postgres

import (
	"context"
	"testing"
)

// TestNew 测试创建 PostgreSQL 配置源
func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "invalid DSN",
			dsn:     "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.dsn, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSource_Name 测试配置源名称
func TestSource_Name(t *testing.T) {
	s := &Source{}
	if got := s.Name(); got != "postgres" {
		t.Errorf("Name() = %v, want postgres", got)
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
			priority: 75,
			want:     75,
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

// TestSource_Watch 测试 Watch 功能
func TestSource_Watch(t *testing.T) {
	s := &Source{}

	// PostgreSQL 源不支持 Watch
	w := s.Watch()
	if w != nil {
		t.Error("Watch() should return nil for PostgreSQL source")
	}
}

// TestSource_Load 测试加载配置
// 注意：此测试需要运行中的 PostgreSQL 服务器和配置表
func TestSource_Load(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// 这里需要实际的数据库连接，跳过集成测试
	t.Skip("requires actual PostgreSQL connection")
}

// TestWithTable 测试表名选项
func TestWithTable(t *testing.T) {
	s := &Source{}
	opt := WithTable("my_config")
	opt(s)
	if s.table != "my_config" {
		t.Errorf("WithTable() table = %v, want my_config", s.table)
	}
}

// TestWithColumns 测试列名选项
func TestWithColumns(t *testing.T) {
	s := &Source{}
	opt := WithColumns("config_key", "config_value")
	opt(s)
	if s.keyCol != "config_key" {
		t.Errorf("WithColumns() keyCol = %v, want config_key", s.keyCol)
	}
	if s.valueCol != "config_value" {
		t.Errorf("WithColumns() valueCol = %v, want config_value", s.valueCol)
	}
}

// TestClose 测试关闭连接
func TestClose(t *testing.T) {
	// 由于需要实际的数据库连接，跳过此测试
	t.Skip("requires actual PostgreSQL connection")
}

// BenchmarkSource_Load 基准测试
func BenchmarkSource_Load(b *testing.B) {
	// 跳过基准测试，需要实际的数据库连接
	b.Skip("requires actual PostgreSQL connection")

	ctx := context.Background()
	s := &Source{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.Load(ctx)
	}
}
