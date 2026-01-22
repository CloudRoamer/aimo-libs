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

func TestSource_EdgeCases(t *testing.T) {
	t.Run("empty_prefix", func(t *testing.T) {
		os.Setenv("NO_PREFIX_VAR", "value")
		defer os.Unsetenv("NO_PREFIX_VAR")

		source := New(WithPrefix(""))
		ctx := context.Background()
		values, err := source.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// 空前缀应加载所有环境变量
		if len(values) == 0 {
			t.Error("Empty prefix should load all env vars")
		}
	})

	t.Run("special_characters", func(t *testing.T) {
		os.Setenv("TEST_KEY-WITH-DASHES", "value1")
		os.Setenv("TEST_KEY.WITH.DOTS", "value2")
		os.Setenv("TEST_KEY__WITH__DOUBLE", "value3")
		defer func() {
			os.Unsetenv("TEST_KEY-WITH-DASHES")
			os.Unsetenv("TEST_KEY.WITH.DOTS")
			os.Unsetenv("TEST_KEY__WITH__DOUBLE")
		}()

		source := New(WithPrefix("TEST_"))
		ctx := context.Background()
		values, err := source.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// 验证特殊字符是否正确处理
		if len(values) < 3 {
			t.Errorf("Expected at least 3 values, got %d", len(values))
		}
	})

	t.Run("empty_value", func(t *testing.T) {
		os.Setenv("TEST_EMPTY_VAR", "")
		defer os.Unsetenv("TEST_EMPTY_VAR")

		source := New(WithPrefix("TEST_"))
		ctx := context.Background()
		values, err := source.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// 空值应该被加载
		if val, ok := values["empty.var"]; !ok {
			t.Error("Empty value should be loaded")
		} else if val.String() != "" {
			t.Errorf("Expected empty string, got %v", val.String())
		}
	})

	t.Run("numeric_value", func(t *testing.T) {
		os.Setenv("TEST_NUMBER", "12345")
		defer os.Unsetenv("TEST_NUMBER")

		source := New(WithPrefix("TEST_"))
		ctx := context.Background()
		values, err := source.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if val, ok := values["number"]; !ok {
			t.Error("Numeric value should be loaded")
		} else if val.Int(0) != 12345 {
			t.Errorf("Expected 12345, got %v", val.Int(0))
		}
	})

	t.Run("unicode_value", func(t *testing.T) {
		os.Setenv("TEST_UNICODE", "你好世界")
		defer os.Unsetenv("TEST_UNICODE")

		source := New(WithPrefix("TEST_"))
		ctx := context.Background()
		values, err := source.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if val, ok := values["unicode"]; !ok {
			t.Error("Unicode value should be loaded")
		} else if val.String() != "你好世界" {
			t.Errorf("Expected 你好世界, got %v", val.String())
		}
	})

	t.Run("whitespace_value", func(t *testing.T) {
		os.Setenv("TEST_WHITESPACE", "  spaces  ")
		defer os.Unsetenv("TEST_WHITESPACE")

		source := New(WithPrefix("TEST_"))
		ctx := context.Background()
		values, err := source.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// 空白应该被保留
		if val, ok := values["whitespace"]; !ok {
			t.Error("Whitespace value should be loaded")
		} else if val.String() != "  spaces  " {
			t.Errorf("Whitespace should be preserved")
		}
	})

	t.Run("multi_level_separator", func(t *testing.T) {
		os.Setenv("TEST_A_B_C_D", "nested")
		defer os.Unsetenv("TEST_A_B_C_D")

		source := New(WithPrefix("TEST_"))
		ctx := context.Background()
		values, err := source.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		// 验证多级嵌套转换
		if val, ok := values["a.b.c.d"]; !ok {
			t.Error("Multi-level key should be converted")
		} else if val.String() != "nested" {
			t.Errorf("Expected nested, got %v", val.String())
		}
	})
}

func TestSource_Watch(t *testing.T) {
	source := New()
	watcher := source.Watch()
	if watcher != nil {
		t.Error("Env source should not support watch")
	}
}

func TestSource_Name(t *testing.T) {
	source := New()
	if source.Name() != "env" {
		t.Errorf("Expected name 'env', got %s", source.Name())
	}
}
