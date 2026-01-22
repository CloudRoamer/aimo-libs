package config

import (
	"testing"
	"time"
)

func TestValue_String(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool", true, "true"},
		{"nil", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValueFromInterface(tt.input)
			if got := v.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValue_Int(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int
	}{
		{"int", 42, 42},
		{"string_int", "123", 123},
		{"int64", int64(999), 999},
		{"float64", float64(100.5), 100},
		{"invalid", "abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValueFromInterface(tt.input)
			if got := v.Int(0); got != tt.expected {
				t.Errorf("Int() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValue_Bool(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"bool_true", true, true},
		{"bool_false", false, false},
		{"string_true", "true", true},
		{"string_false", "false", false},
		{"string_1", "1", true},
		{"string_0", "0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValueFromInterface(tt.input)
			if got := v.Bool(false); got != tt.expected {
				t.Errorf("Bool() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValue_Duration(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected time.Duration
	}{
		{"duration", time.Second * 30, 30 * time.Second},
		{"string_duration", "1m30s", 90 * time.Second},
		{"int64_ns", int64(1000000000), time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValueFromInterface(tt.input)
			if got := v.Duration(0); got != tt.expected {
				t.Errorf("Duration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValue_StringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []string
	}{
		{"string_slice", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"interface_slice", []any{"x", "y"}, []string{"x", "y"}},
		{"json_string", `["foo","bar"]`, []string{"foo", "bar"}},
		{"comma_separated", "one,two,three", []string{"one", "two", "three"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValueFromInterface(tt.input)
			got := v.StringSlice(nil)
			if len(got) != len(tt.expected) {
				t.Errorf("StringSlice() length = %v, want %v", len(got), len(tt.expected))
				return
			}
			for i, val := range got {
				if val != tt.expected[i] {
					t.Errorf("StringSlice()[%d] = %v, want %v", i, val, tt.expected[i])
				}
			}
		})
	}
}

func TestValue_StringMap(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected map[string]string
	}{
		{
			"string_map",
			map[string]string{"key": "value"},
			map[string]string{"key": "value"},
		},
		{
			"interface_map",
			map[string]any{"foo": "bar", "count": 42},
			map[string]string{"foo": "bar", "count": "42"},
		},
		{
			"json_string",
			`{"name":"test","version":"1.0"}`,
			map[string]string{"name": "test", "version": "1.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValueFromInterface(tt.input)
			got := v.StringMap(nil)
			if len(got) != len(tt.expected) {
				t.Errorf("StringMap() length = %v, want %v", len(got), len(tt.expected))
				return
			}
			for k, val := range tt.expected {
				if got[k] != val {
					t.Errorf("StringMap()[%s] = %v, want %v", k, got[k], val)
				}
			}
		})
	}
}

// TestValue_EdgeCases 测试边界情况
func TestValue_EdgeCases(t *testing.T) {
	t.Run("nil_value", func(t *testing.T) {
		v := NewValueFromInterface(nil)
		if v.String() != "" {
			t.Errorf("nil value String() should return empty string")
		}
		if v.Int(42) != 42 {
			t.Errorf("nil value Int() should return default")
		}
		if v.Bool(true) != true {
			t.Errorf("nil value Bool() should return default")
		}
		if v.Float64(3.14) != 3.14 {
			t.Errorf("nil value Float64() should return default")
		}
	})

	t.Run("overflow_int", func(t *testing.T) {
		v := NewValueFromInterface(int64(9223372036854775807))
		result := v.Int(0)
		if result == 0 {
			t.Errorf("large int64 should convert to int")
		}
	})

	t.Run("invalid_string_conversions", func(t *testing.T) {
		tests := []struct {
			name    string
			input   string
			intDef  int
			floatDef float64
			boolDef bool
		}{
			{"letters_to_int", "abc", 99, 0, false},
			{"letters_to_float", "xyz", 0, 99.9, false},
			{"letters_to_bool", "notabool", 0, 0, true},
			{"empty_string", "", 42, 3.14, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				v := NewValue(tt.input)
				if got := v.Int(tt.intDef); got != tt.intDef {
					t.Errorf("Int() = %v, want default %v", got, tt.intDef)
				}
				if got := v.Float64(tt.floatDef); got != tt.floatDef {
					t.Errorf("Float64() = %v, want default %v", got, tt.floatDef)
				}
			})
		}
	})

	t.Run("duration_invalid", func(t *testing.T) {
		v := NewValue("invalid_duration")
		defaultDuration := 10 * time.Second
		if got := v.Duration(defaultDuration); got != defaultDuration {
			t.Errorf("invalid duration should return default")
		}
	})

	t.Run("string_slice_empty", func(t *testing.T) {
		tests := []struct {
			name        string
			input       any
			expectEmpty bool
		}{
			{"empty_json_array", "[]", true},
			{"empty_string", "", true},
			{"whitespace_only", "   ", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				v := NewValueFromInterface(tt.input)
				got := v.StringSlice(nil)
				if tt.expectEmpty && len(got) != 0 {
					t.Errorf("StringSlice() should return empty for %s, got %v", tt.name, got)
				}
			})
		}
	})

	t.Run("string_slice_comma_fallback", func(t *testing.T) {
		// 无效 JSON 会回退到逗号分隔解析
		v := NewValue("[broken")
		got := v.StringSlice(nil)
		// 应该解析为单个元素 "[broken"
		if len(got) != 1 || got[0] != "[broken" {
			t.Errorf("Invalid JSON should fallback to comma-separated, got %v", got)
		}
	})

	t.Run("string_map_invalid_json", func(t *testing.T) {
		v := NewValue("{broken json}")
		defaultMap := map[string]string{"fallback": "value"}
		got := v.StringMap(defaultMap)
		if len(got) != len(defaultMap) {
			t.Errorf("invalid JSON should return default map")
		}
	})

	t.Run("raw_access", func(t *testing.T) {
		original := map[string]int{"key": 42}
		v := NewValueFromInterface(original)
		raw := v.Raw()
		if raw == nil {
			t.Errorf("Raw() should return the original value")
		}
	})
}

// TestValue_Float32Conversion 测试 float32 转换
func TestValue_Float32Conversion(t *testing.T) {
	v := NewValueFromInterface(float32(3.14))
	if got := v.Float64(0); got != float64(float32(3.14)) {
		t.Errorf("Float32 conversion failed: got %v", got)
	}
}

// TestValue_ConcurrentAccess 测试并发访问安全（Value 是不可变的）
func TestValue_ConcurrentAccess(t *testing.T) {
	v := NewValue("concurrent_test")

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = v.String()
				_ = v.Int(0)
				_ = v.Bool(false)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
