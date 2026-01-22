package config

import (
	"testing"
	"time"
)

func TestValue_String(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
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
		input    interface{}
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
		input    interface{}
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
		input    interface{}
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
		input    interface{}
		expected []string
	}{
		{"string_slice", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"interface_slice", []interface{}{"x", "y"}, []string{"x", "y"}},
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
		input    interface{}
		expected map[string]string
	}{
		{
			"string_map",
			map[string]string{"key": "value"},
			map[string]string{"key": "value"},
		},
		{
			"interface_map",
			map[string]interface{}{"foo": "bar", "count": 42},
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
