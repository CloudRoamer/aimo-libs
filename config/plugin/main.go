//go:build linux || darwin

package main

import (
	"github.com/CloudRoamer/aimo-libs/config"
	"github.com/CloudRoamer/aimo-libs/config/codec"
	"github.com/CloudRoamer/aimo-libs/config/source/consul"
	"github.com/CloudRoamer/aimo-libs/config/source/env"
	"github.com/CloudRoamer/aimo-libs/config/source/file"
	"github.com/CloudRoamer/aimo-libs/config/source/postgres"
)

// 导出核心类型和函数
var (
	NewManager           = config.NewManager
	NewDefaultMerger     = config.NewDefaultMerger
	NewValue             = config.NewValue
	NewValueFromInterface = config.NewValueFromInterface
)

// 导出 Manager 选项
var (
	WithMerger = config.WithMerger
)

// 包装函数：返回 config.Source 接口而非具体类型
// 这样可以避免类型断言问题

// NewEnvSource 创建环境变量配置源
func NewEnvSource(opts ...env.Option) config.Source {
	return env.New(opts...)
}

// NewFileSource 创建文件配置源
func NewFileSource(path string, opts ...file.Option) (config.Source, error) {
	return file.New(path, opts...)
}

// NewConsulSource 创建 Consul 配置源
func NewConsulSource(address string, opts ...consul.Option) (config.Source, error) {
	return consul.New(address, opts...)
}

// NewPostgresSource 创建 PostgreSQL 配置源
func NewPostgresSource(dsn string, opts ...postgres.Option) (config.Source, error) {
	return postgres.New(dsn, opts...)
}

// Env Source 选项
var (
	EnvWithPrefix     = env.WithPrefix
	EnvWithSeparator  = env.WithSeparator
	EnvWithPriority   = env.WithPriority
	EnvWithKeyMapping = env.WithKeyMapping
)

// File Source 选项
var (
	FileWithPriority = file.WithPriority
	FileWithCodec    = file.WithCodec
)

// Consul Source 选项
var (
	ConsulWithPrefix    = consul.WithPrefix
	ConsulWithPriority  = consul.WithPriority
	ConsulWithSeparator = consul.WithSeparator
)

// Postgres Source 选项
var (
	PostgresWithTable    = postgres.WithTable
	PostgresWithColumns  = postgres.WithColumns
	PostgresWithPriority = postgres.WithPriority
)

// Codec 编解码器
var (
	CodecJSON = codec.JSON
	CodecYAML = codec.YAML
)

// EventType 枚举
const (
	EventTypeUnknown = config.EventTypeUnknown
	EventTypeCreate  = config.EventTypeCreate
	EventTypeUpdate  = config.EventTypeUpdate
	EventTypeDelete  = config.EventTypeDelete
	EventTypeReload  = config.EventTypeReload
	EventTypeError   = config.EventTypeError
)

// Source 优先级常量
const (
	EnvSourcePriority      = env.DefaultPriority
	ConsulSourcePriority   = consul.DefaultPriority
	PostgresSourcePriority = postgres.DefaultPriority
	FileSourcePriority     = file.DefaultPriority
)

func main() {}
