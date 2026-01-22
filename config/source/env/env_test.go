package env

import (
	"context"
	"os"
	"testing"
)

func TestSource_Load(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("TEST_DATABASE_HOST", "localhost")
	os.Setenv("TEST_DATABASE_PORT", "5432")
	os.Setenv("OTHER_VAR", "should_be_ignored")
	defer func() {
		os.Unsetenv("TEST_DATABASE_HOST")
		os.Unsetenv("TEST_DATABASE_PORT")
		os.Unsetenv("OTHER_VAR")
	}()

	source := New(WithPrefix("TEST_"))

	ctx := context.Background()
	values, err := source.Load(ctx)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// 检查是否正确加载带前缀的变量
	if val, ok := values["database.host"]; !ok || val.String() != "localhost" {
		t.Errorf("Expected database.host=localhost, got %v", val.String())
	}

	if val, ok := values["database.port"]; !ok || val.String() != "5432" {
		t.Errorf("Expected database.port=5432, got %v", val.String())
	}

	// 检查是否忽略了不带前缀的变量
	if _, ok := values["other.var"]; ok {
		t.Errorf("Should not load OTHER_VAR without prefix")
	}
}

func TestSource_Priority(t *testing.T) {
	source := New()
	if source.Priority() != DefaultPriority {
		t.Errorf("Expected priority %d, got %d", DefaultPriority, source.Priority())
	}

	customPriority := 200
	source = New(WithPriority(customPriority))
	if source.Priority() != customPriority {
		t.Errorf("Expected priority %d, got %d", customPriority, source.Priority())
	}
}

func TestSource_CustomKeyMapping(t *testing.T) {
	os.Setenv("MYAPP_DB_HOST", "localhost")
	defer os.Unsetenv("MYAPP_DB_HOST")

	// 自定义映射：保持大写
	source := New(
		WithPrefix("MYAPP_"),
		WithKeyMapping(func(key string) string {
			return key // 保持原样
		}),
	)

	ctx := context.Background()
	values, err := source.Load(ctx)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if _, ok := values["MYAPP_DB_HOST"]; !ok {
		t.Errorf("Custom key mapping failed")
	}
}
