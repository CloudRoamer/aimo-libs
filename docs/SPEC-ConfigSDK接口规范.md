# Config SDK 接口规范

## 修订记录

| 版本 | 日期 | 修订内容 | 修订人 |
|------|------|----------|--------|
| v1.0 | 2026-01-23 | 初始版本创建，整理 Config SDK 对外接口 | AIMO开发团队 |

---

## 1. 概述

本规范描述 aimo-libs Config SDK 的公共接口、配置模型与内置配置源的使用方式，面向调用方与扩展配置源的实现方。

### 1.1 适用范围

- 配置访问接口与类型转换规则
- 配置源与监听器接口
- 配置管理器行为与回调协议
- 内置配置源与优先级约定

### 1.2 设计目标

- 提供统一的配置读取 API
- 支持多配置源合并与优先级覆盖
- 支持热更新与变更事件回调
- 保持接口最小化与易扩展

## 2. 版本信息

| 项目 | 值 |
|------|------|
| Module | `github.com/CloudRoamer/aimo-libs/config` |
| Go 版本 | 1.25.5 |

## 3. 接口概览

### 3.1 核心接口

| 接口 | 说明 |
|------|------|
| `Source` | 配置源接口，负责加载配置并提供监听器 |
| `Watcher` | 配置变更监听器接口 |
| `Manager` | 配置管理器，负责多源合并与热更新 |
| `Config` | 只读配置访问接口 |
| `Merger` | 配置合并策略接口 |

### 3.2 事件模型

`EventType` 枚举：

| 枚举值 | 含义 |
|------|------|
| `EventTypeCreate` | 新增配置 |
| `EventTypeUpdate` | 更新配置 |
| `EventTypeDelete` | 删除配置 |
| `EventTypeReload` | 全量重载 |
| `EventTypeError` | 监听错误 |

### 3.3 配置键约定

- 配置键使用点分隔符表示层级，如 `database.host` 。
- 环境变量默认会将 `_` 转为 `.` 并转换为小写。
- 合并规则：优先级高的配置源覆盖优先级低的配置源。

## 4. 接口定义

### 4.1 Source

```go
type Source interface {
    Name() string
    Priority() int
    Load(ctx context.Context) (map[string]Value, error)
    Watch() Watcher
}
```

行为说明：

- `Load` 返回扁平化的 key-value 映射。
- `Watch` 返回 `Watcher`，若不支持监听则返回 `nil`。

### 4.2 Watcher

```go
type Watcher interface {
    Start(ctx context.Context) (<-chan Event, error)
    Stop() error
}
```

行为说明：

- `Start` 返回事件通道，事件包含 `Type`、`Source`、`Keys`、`Timestamp`、`Error`。
- `Stop` 负责释放监听资源。

### 4.3 Manager

```go
func NewManager(opts ...ManagerOption) *Manager

func (m *Manager) AddSource(sources ...Source) *Manager
func (m *Manager) Load(ctx context.Context) error
func (m *Manager) Watch() error
func (m *Manager) OnChange(callback ChangeCallback)
func (m *Manager) Config() Config
func (m *Manager) Close() error
```

行为说明：

- `AddSource` 会按优先级排序，优先级高者在合并时覆盖。
- `Watch` 会为所有可监听的配置源启动事件处理协程。
- `OnChange` 注册变更回调，回调签名为：

```go
type ChangeCallback func(event Event, oldConfig, newConfig Config)
```

### 4.4 Config

```go
type Config interface {
    Get(key string) (Value, bool)
    GetString(key string, defaultVal string) string
    GetInt(key string, defaultVal int) int
    GetInt64(key string, defaultVal int64) int64
    GetFloat64(key string, defaultVal float64) float64
    GetBool(key string, defaultVal bool) bool
    GetDuration(key string, defaultVal time.Duration) time.Duration
    GetStringSlice(key string, defaultVal []string) []string
    GetStringMap(key string, defaultVal map[string]string) map[string]string
    Keys() []string
    Has(key string) bool
}
```

### 4.5 Value

```go
type Value struct {
    raw any
}

func NewValue(s string) Value
func NewValueFromInterface(v any) Value
func (v Value) Raw() any
func (v Value) String() string
func (v Value) Int(defaultVal int) int
func (v Value) Int64(defaultVal int64) int64
func (v Value) Float64(defaultVal float64) float64
func (v Value) Bool(defaultVal bool) bool
func (v Value) Duration(defaultVal time.Duration) time.Duration
func (v Value) StringSlice(defaultVal []string) []string
func (v Value) StringMap(defaultVal map[string]string) map[string]string
```

### 4.6 Merger 与选项

```go
type Merger interface {
    Merge(maps ...map[string]Value) map[string]Value
}

type ManagerOption func(*Manager)

func WithMerger(merger Merger) ManagerOption
```

## 5. 内置配置源

### 5.1 环境变量源 `env`

```go
source := env.New(
    env.WithPrefix("APP_"),
    env.WithSeparator("_"),
    env.WithPriority(100),
)
```

常用选项：

| 选项 | 说明 |
|------|------|
| `WithPrefix` | 设置环境变量前缀 |
| `WithSeparator` | 设置分隔符，默认 `_` |
| `WithPriority` | 设置优先级 |
| `WithKeyMapping` | 自定义 key 映射函数 |

### 5.2 文件源 `file`

```go
source, err := file.New("./config.yaml", file.WithPriority(60))
```

常用选项：

| 选项 | 说明 |
|------|------|
| `WithCodec` | 指定编解码器（JSON/YAML） |
| `WithPriority` | 设置优先级 |

### 5.3 Consul 源 `consul`

```go
source, err := consul.New("http://127.0.0.1:8500",
    consul.WithPrefix("config/dev/app"),
    consul.WithPriority(80),
)
```

常用选项：

| 选项 | 说明 |
|------|------|
| `WithPrefix` | 设置 KV 前缀 |
| `WithPriority` | 设置优先级 |
| `WithConfig` | 使用自定义 Consul 配置 |
| `WithSeparator` | 设置路径分隔符 |

### 5.4 PostgreSQL 源 `postgres`

```go
source, err := postgres.New(dsn,
    postgres.WithTable("app_config"),
    postgres.WithColumns("key", "value"),
    postgres.WithPriority(70),
)
```

常用选项：

| 选项 | 说明 |
|------|------|
| `WithTable` | 配置表名 |
| `WithColumns` | 设置 key/value 列名 |
| `WithPriority` | 设置优先级 |

## 6. 使用示例

```go
mgr := config.NewManager()

fileSource, err := file.New("./config.yaml")
if err != nil {
    return err
}

envSource := env.New(env.WithPrefix("APP_"))

mgr.AddSource(fileSource)
mgr.AddSource(envSource)

if err := mgr.Load(context.Background()); err != nil {
    return err
}

cfg := mgr.Config()
port := cfg.GetInt("server.port", 8080)
```

## 7. 错误处理

- `Load` 返回的错误应使用 `%w` 包装保留原始错误。
- `Watcher` 在监听异常时应发送 `EventTypeError` 事件。
- `Manager` 在事件处理失败时会回调错误事件，不会停止监听。

## 8. 并发与生命周期

- `Manager` 对外提供线程安全的配置访问与事件处理。
- `Close` 会停止所有监听器并等待事件处理协程退出。
- 建议在应用退出时显式调用 `Close` 释放资源。
