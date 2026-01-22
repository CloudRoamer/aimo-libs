package config

import (
	"context"
	"time"
)

// Source 定义配置源接口
// 每个配置源（环境变量、Consul、文件等）都需要实现此接口
type Source interface {
	// Name 返回配置源的唯一标识名称
	// 用于日志记录和错误追踪
	Name() string

	// Priority 返回配置源的优先级
	// 数值越大优先级越高，合并时高优先级覆盖低优先级
	// 推荐值: Env=100, Consul=80, File=60, Default=0
	Priority() int

	// Load 加载配置数据
	// 返回扁平化的 key-value 映射，key 使用点分隔符表示层级
	// 例如: {"database.host": "localhost", "database.port": "5432"}
	Load(ctx context.Context) (map[string]Value, error)

	// Watch 返回配置变更监听器（可选）
	// 如果配置源不支持监听，返回 nil
	Watch() Watcher
}

// Watcher 定义配置变更监听接口
type Watcher interface {
	// Start 启动监听
	// 当配置发生变更时，通过 channel 发送事件
	Start(ctx context.Context) (<-chan Event, error)

	// Stop 停止监听
	Stop() error
}

// Event 定义配置变更事件
type Event struct {
	// Type 事件类型
	Type EventType

	// Source 触发事件的配置源名称
	Source string

	// Keys 发生变更的配置键列表（可选）
	// 如果为空，表示需要重新加载整个配置源
	Keys []string

	// Timestamp 事件发生时间
	Timestamp time.Time

	// Error 如果监听过程中发生错误
	Error error
}

// EventType 事件类型枚举
type EventType int

const (
	EventTypeUnknown EventType = iota
	EventTypeCreate            // 新增配置
	EventTypeUpdate            // 更新配置
	EventTypeDelete            // 删除配置
	EventTypeReload            // 全量重载
	EventTypeError             // 监听错误
)

func (e EventType) String() string {
	switch e {
	case EventTypeCreate:
		return "create"
	case EventTypeUpdate:
		return "update"
	case EventTypeDelete:
		return "delete"
	case EventTypeReload:
		return "reload"
	case EventTypeError:
		return "error"
	default:
		return "unknown"
	}
}

// Config 定义配置访问接口
// 提供类型安全的配置值获取方法
type Config interface {
	// Get 获取原始配置值
	// key 支持点分隔符表示嵌套路径，如 "database.host"
	Get(key string) (Value, bool)

	// GetString 获取字符串值，不存在时返回默认值
	GetString(key string, defaultVal string) string

	// GetInt 获取整数值，不存在或类型转换失败时返回默认值
	GetInt(key string, defaultVal int) int

	// GetInt64 获取 int64 值
	GetInt64(key string, defaultVal int64) int64

	// GetFloat64 获取浮点数值
	GetFloat64(key string, defaultVal float64) float64

	// GetBool 获取布尔值
	GetBool(key string, defaultVal bool) bool

	// GetDuration 获取时间间隔值
	// 支持格式: "10s", "5m", "1h30m" 等
	GetDuration(key string, defaultVal time.Duration) time.Duration

	// GetStringSlice 获取字符串切片
	// 支持 JSON 数组格式或逗号分隔格式
	GetStringSlice(key string, defaultVal []string) []string

	// GetStringMap 获取字符串映射
	GetStringMap(key string, defaultVal map[string]string) map[string]string

	// Keys 返回所有配置键
	Keys() []string

	// Has 检查配置键是否存在
	Has(key string) bool
}

// configImpl Config 接口的实现
type configImpl struct {
	data map[string]Value
}

// newConfigImpl 创建配置实例
func newConfigImpl() *configImpl {
	return &configImpl{
		data: make(map[string]Value),
	}
}

// newConfigImplFromMap 从 map 创建配置实例
func newConfigImplFromMap(data map[string]Value) *configImpl {
	return &configImpl{
		data: data,
	}
}

func (c *configImpl) Get(key string) (Value, bool) {
	val, ok := c.data[key]
	return val, ok
}

func (c *configImpl) GetString(key string, defaultVal string) string {
	if val, ok := c.data[key]; ok {
		return val.String()
	}
	return defaultVal
}

func (c *configImpl) GetInt(key string, defaultVal int) int {
	if val, ok := c.data[key]; ok {
		return val.Int(defaultVal)
	}
	return defaultVal
}

func (c *configImpl) GetInt64(key string, defaultVal int64) int64 {
	if val, ok := c.data[key]; ok {
		return val.Int64(defaultVal)
	}
	return defaultVal
}

func (c *configImpl) GetFloat64(key string, defaultVal float64) float64 {
	if val, ok := c.data[key]; ok {
		return val.Float64(defaultVal)
	}
	return defaultVal
}

func (c *configImpl) GetBool(key string, defaultVal bool) bool {
	if val, ok := c.data[key]; ok {
		return val.Bool(defaultVal)
	}
	return defaultVal
}

func (c *configImpl) GetDuration(key string, defaultVal time.Duration) time.Duration {
	if val, ok := c.data[key]; ok {
		return val.Duration(defaultVal)
	}
	return defaultVal
}

func (c *configImpl) GetStringSlice(key string, defaultVal []string) []string {
	if val, ok := c.data[key]; ok {
		return val.StringSlice(defaultVal)
	}
	return defaultVal
}

func (c *configImpl) GetStringMap(key string, defaultVal map[string]string) map[string]string {
	if val, ok := c.data[key]; ok {
		return val.StringMap(defaultVal)
	}
	return defaultVal
}

func (c *configImpl) Keys() []string {
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

func (c *configImpl) Has(key string) bool {
	_, ok := c.data[key]
	return ok
}
