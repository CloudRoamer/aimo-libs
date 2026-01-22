package config

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Value 封装配置值，提供类型转换方法
type Value struct {
	raw any // 原始值
}

// NewValue 从字符串创建 Value
func NewValue(s string) Value {
	return Value{raw: s}
}

// NewValueFromInterface 从任意类型创建 Value
func NewValueFromInterface(v any) Value {
	return Value{raw: v}
}

// Raw 返回原始值
func (v Value) Raw() any {
	return v.raw
}

// String 返回字符串表示
func (v Value) String() string {
	if v.raw == nil {
		return ""
	}
	switch val := v.raw.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Int 转换为整数
func (v Value) Int(defaultVal int) int {
	if v.raw == nil {
		return defaultVal
	}

	switch val := v.raw.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		i, err := strconv.Atoi(val)
		if err != nil {
			return defaultVal
		}
		return i
	default:
		return defaultVal
	}
}

// Int64 转换为 int64
func (v Value) Int64(defaultVal int64) int64 {
	if v.raw == nil {
		return defaultVal
	}

	switch val := v.raw.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return defaultVal
		}
		return i
	default:
		return defaultVal
	}
}

// Float64 转换为浮点数
func (v Value) Float64(defaultVal float64) float64 {
	if v.raw == nil {
		return defaultVal
	}

	switch val := v.raw.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return defaultVal
		}
		return f
	default:
		return defaultVal
	}
}

// Bool 转换为布尔值
func (v Value) Bool(defaultVal bool) bool {
	if v.raw == nil {
		return defaultVal
	}

	switch val := v.raw.(type) {
	case bool:
		return val
	case string:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return defaultVal
		}
		return b
	default:
		return defaultVal
	}
}

// Duration 转换为时间间隔
func (v Value) Duration(defaultVal time.Duration) time.Duration {
	if v.raw == nil {
		return defaultVal
	}

	switch val := v.raw.(type) {
	case time.Duration:
		return val
	case int64:
		return time.Duration(val)
	case string:
		d, err := time.ParseDuration(val)
		if err != nil {
			return defaultVal
		}
		return d
	default:
		return defaultVal
	}
}

// StringSlice 转换为字符串切片
func (v Value) StringSlice(defaultVal []string) []string {
	if v.raw == nil {
		return defaultVal
	}

	switch val := v.raw.(type) {
	case []string:
		return val
	case []any:
		result := make([]string, 0, len(val))
		for _, item := range val {
			result = append(result, fmt.Sprintf("%v", item))
		}
		return result
	case string:
		// 尝试 JSON 解析
		var slice []string
		if err := json.Unmarshal([]byte(val), &slice); err == nil {
			return slice
		}
		// 尝试逗号分隔
		parts := strings.Split(val, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
		return defaultVal
	default:
		return defaultVal
	}
}

// StringMap 转换为字符串映射
func (v Value) StringMap(defaultVal map[string]string) map[string]string {
	if v.raw == nil {
		return defaultVal
	}

	switch val := v.raw.(type) {
	case map[string]string:
		return val
	case map[string]any:
		result := make(map[string]string, len(val))
		for k, v := range val {
			result[k] = fmt.Sprintf("%v", v)
		}
		return result
	case string:
		var m map[string]string
		if err := json.Unmarshal([]byte(val), &m); err == nil {
			return m
		}
		return defaultVal
	default:
		return defaultVal
	}
}
