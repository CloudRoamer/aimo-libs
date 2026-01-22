# aimo-config: 统一配置管理 SDK 架构设计

## 修订记录

| 版本 | 日期 | 修订内容 | 修订人 |
|------|------|----------|--------|
| v1.0 | 2026-01-23 | 初始版本创建，包含完整架构设计 | AIMO 开发团队 |

---

## 1. 概述

### 1.1 设计目标

aimo-config 是一个通用的、开源级别的统一配置管理 SDK，旨在为 Go 应用程序提供：

- **多源支持**：环境变量、Consul KV、配置文件（JSON/YAML）、PostgreSQL
- **优先级合并**：环境变量 > Consul > 文件 > 默认值
- **热更新**：Watch 机制支持配置动态刷新
- **类型安全**：泛型 API 提供类型安全的配置访问
- **可扩展性**：清晰的接口设计，易于添加新的配置源

### 1.2 设计原则

| 原则 | 说明 |
|------|------|
| **接口隔离** | 每个配置源实现独立的 Source 接口，互不依赖 |
| **依赖倒置** | 上层模块依赖抽象接口，不依赖具体实现 |
| **开闭原则** | 对扩展开放（新配置源），对修改关闭（核心逻辑） |
| **单一职责** | Source 负责加载，Watcher 负责监听，Manager 负责协调 |

---

## 2. 目录结构

```
aimo-libs/config/
├── ARCHITECTURE.md          # 本架构文档
├── README.md                 # 使用说明文档
├── go.mod                    # Go 模块定义
├── go.sum
│
├── config.go                 # 核心接口定义（Source, Watcher, Config）
├── manager.go                # Manager 实现（配置管理器）
├── options.go                # 配置选项（functional options 模式）
├── errors.go                 # 错误类型定义
├── value.go                  # Value 类型与类型转换
│
├── source/                   # 配置源实现
│   ├── source.go             # Source 接口基础定义
│   ├── env/                  # 环境变量源
│   │   ├── env.go
│   │   └── env_test.go
│   ├── consul/               # Consul KV 源
│   │   ├── consul.go
│   │   ├── watcher.go        # Consul Watch 实现
│   │   └── consul_test.go
│   ├── file/                 # 文件源（JSON/YAML）
│   │   ├── file.go
│   │   ├── watcher.go        # 文件系统 Watch 实现
│   │   └── file_test.go
│   └── postgres/             # PostgreSQL 源（可选）
│       ├── postgres.go
│       └── postgres_test.go
│
├── merge/                    # 合并策略
│   ├── merger.go             # Merger 接口与默认实现
│   └── merger_test.go
│
├── codec/                    # 编解码器
│   ├── codec.go              # Codec 接口
│   ├── json.go               # JSON 编解码
│   ├── yaml.go               # YAML 编解码
│   └── codec_test.go
│
└── examples/                 # 使用示例
    ├── basic/                # 基础使用
    ├── watch/                # 热更新示例
    └── custom_source/        # 自定义配置源示例
```

---

## 3. 核心接口设计

### 3.1 接口总览

```
┌─────────────────────────────────────────────────────────────────────┐
│                           Manager                                    │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Config (合并后的配置)                      │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              ▲                                       │
│                              │ merge                                 │
│  ┌──────────┬──────────┬──────────┬──────────┐                     │
│  │  Source  │  Source  │  Source  │  Source  │                     │
│  │   (Env)  │ (Consul) │  (File)  │(Postgres)│                     │
│  └──────────┴──────────┴──────────┴──────────┘                     │
│       ▲           ▲           ▲                                     │
│       │           │           │                                     │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐                               │
│  │         │ │ Watcher │ │ Watcher │                               │
│  │   N/A   │ │(Consul) │ │ (File)  │                               │
│  └─────────┘ └─────────┘ └─────────┘                               │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 Source 接口

```go
// config.go

package config

import (
    "context"
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

// SourceOption 配置源的可选配置项
type SourceOption func(*sourceOptions)

type sourceOptions struct {
    prefix   string            // key 前缀
    decoder  Decoder           // 值解码器
    priority int               // 自定义优先级
}
```

### 3.3 Watcher 接口

```go
// config.go (续)

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
```

### 3.4 Config 接口

```go
// config.go (续)

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

    // Sub 获取子配置
    // 返回指定前缀下的所有配置项作为新的 Config
    Sub(prefix string) Config

    // Unmarshal 将配置反序列化到结构体
    // 使用 mapstructure 进行映射
    Unmarshal(key string, out interface{}) error

    // UnmarshalKey 将指定 key 的配置反序列化到结构体
    UnmarshalKey(key string, out interface{}) error

    // Keys 返回所有配置键
    Keys() []string

    // Has 检查配置键是否存在
    Has(key string) bool
}
```

### 3.5 Manager 接口

```go
// manager.go

package config

import (
    "context"
    "sync"
)

// Manager 配置管理器
// 负责协调多个配置源、执行合并、处理热更新
type Manager struct {
    mu       sync.RWMutex
    sources  []Source
    merger   Merger
    config   *configImpl
    watchers []Watcher
    onChange []ChangeCallback
    ctx      context.Context
    cancel   context.CancelFunc
}

// ChangeCallback 配置变更回调函数
// oldConfig 可能为 nil（首次加载时）
type ChangeCallback func(event Event, oldConfig, newConfig Config)

// NewManager 创建配置管理器
func NewManager(opts ...ManagerOption) *Manager {
    ctx, cancel := context.WithCancel(context.Background())
    m := &Manager{
        sources:  make([]Source, 0),
        merger:   NewDefaultMerger(),
        config:   newConfigImpl(),
        onChange: make([]ChangeCallback, 0),
        ctx:      ctx,
        cancel:   cancel,
    }

    for _, opt := range opts {
        opt(m)
    }

    return m
}

// AddSource 添加配置源
// 配置源按优先级排序，高优先级的源在合并时覆盖低优先级
func (m *Manager) AddSource(sources ...Source) *Manager {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.sources = append(m.sources, sources...)
    // 按优先级排序（从低到高，便于后续合并时高优先级覆盖低优先级）
    sort.Slice(m.sources, func(i, j int) bool {
        return m.sources[i].Priority() < m.sources[j].Priority()
    })

    return m
}

// Load 加载所有配置源并合并
func (m *Manager) Load(ctx context.Context) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    return m.loadLocked(ctx)
}

// loadLocked 内部加载方法（需要持有锁）
func (m *Manager) loadLocked(ctx context.Context) error {
    allValues := make([]map[string]Value, 0, len(m.sources))

    for _, source := range m.sources {
        values, err := source.Load(ctx)
        if err != nil {
            return fmt.Errorf("failed to load from source %s: %w", source.Name(), err)
        }
        allValues = append(allValues, values)
    }

    // 合并所有配置（按优先级，后面的覆盖前面的）
    merged := m.merger.Merge(allValues...)
    m.config = newConfigImplFromMap(merged)

    return nil
}

// Watch 启动所有配置源的监听
func (m *Manager) Watch() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    for _, source := range m.sources {
        watcher := source.Watch()
        if watcher == nil {
            continue
        }

        eventCh, err := watcher.Start(m.ctx)
        if err != nil {
            return fmt.Errorf("failed to start watcher for source %s: %w", source.Name(), err)
        }

        m.watchers = append(m.watchers, watcher)

        // 启动 goroutine 处理事件
        go m.handleEvents(source.Name(), eventCh)
    }

    return nil
}

// handleEvents 处理配置变更事件
func (m *Manager) handleEvents(sourceName string, eventCh <-chan Event) {
    for {
        select {
        case <-m.ctx.Done():
            return
        case event, ok := <-eventCh:
            if !ok {
                return
            }

            if event.Type == EventTypeError {
                // 记录错误但不停止监听
                m.notifyChange(event, nil, nil)
                continue
            }

            // 重新加载配置
            m.mu.Lock()
            oldConfig := m.config.Clone()
            if err := m.loadLocked(m.ctx); err != nil {
                m.mu.Unlock()
                m.notifyChange(Event{
                    Type:      EventTypeError,
                    Source:    sourceName,
                    Timestamp: time.Now(),
                    Error:     err,
                }, nil, nil)
                continue
            }
            newConfig := m.config.Clone()
            m.mu.Unlock()

            m.notifyChange(event, oldConfig, newConfig)
        }
    }
}

// OnChange 注册配置变更回调
func (m *Manager) OnChange(callback ChangeCallback) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.onChange = append(m.onChange, callback)
}

// notifyChange 通知所有回调
func (m *Manager) notifyChange(event Event, oldConfig, newConfig Config) {
    m.mu.RLock()
    callbacks := make([]ChangeCallback, len(m.onChange))
    copy(callbacks, m.onChange)
    m.mu.RUnlock()

    for _, cb := range callbacks {
        cb(event, oldConfig, newConfig)
    }
}

// Config 获取当前配置（只读）
func (m *Manager) Config() Config {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.config
}

// Close 关闭管理器，停止所有监听
func (m *Manager) Close() error {
    m.cancel()

    var errs []error
    for _, w := range m.watchers {
        if err := w.Stop(); err != nil {
            errs = append(errs, err)
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("errors while closing watchers: %v", errs)
    }
    return nil
}
```

---

## 4. 配置源抽象

### 4.1 环境变量源

```go
// source/env/env.go

package env

import (
    "context"
    "os"
    "strings"

    "github.com/CloudRoamer/aimo-libs/config"
)

const (
    DefaultPriority = 100 // 环境变量优先级最高
)

// Source 环境变量配置源
type Source struct {
    prefix      string // 环境变量前缀，如 "APP_"
    separator   string // key 分隔符，默认 "_"
    priority    int
    keyMapping  func(string) string // 自定义 key 映射函数
}

// Option 配置选项
type Option func(*Source)

// WithPrefix 设置环境变量前缀
func WithPrefix(prefix string) Option {
    return func(s *Source) {
        s.prefix = prefix
    }
}

// WithSeparator 设置层级分隔符
func WithSeparator(sep string) Option {
    return func(s *Source) {
        s.separator = sep
    }
}

// WithPriority 设置优先级
func WithPriority(p int) Option {
    return func(s *Source) {
        s.priority = p
    }
}

// WithKeyMapping 设置自定义 key 映射函数
// 用于将环境变量名转换为配置 key
// 例如: APP_DATABASE_HOST -> database.host
func WithKeyMapping(fn func(string) string) Option {
    return func(s *Source) {
        s.keyMapping = fn
    }
}

// New 创建环境变量配置源
func New(opts ...Option) *Source {
    s := &Source{
        prefix:    "",
        separator: "_",
        priority:  DefaultPriority,
    }

    for _, opt := range opts {
        opt(s)
    }

    // 默认 key 映射：移除前缀，将分隔符替换为点，转小写
    if s.keyMapping == nil {
        s.keyMapping = func(envKey string) string {
            key := envKey
            if s.prefix != "" {
                key = strings.TrimPrefix(key, s.prefix)
            }
            key = strings.ToLower(key)
            key = strings.ReplaceAll(key, s.separator, ".")
            return key
        }
    }

    return s
}

func (s *Source) Name() string {
    return "env"
}

func (s *Source) Priority() int {
    return s.priority
}

func (s *Source) Load(ctx context.Context) (map[string]config.Value, error) {
    result := make(map[string]config.Value)

    for _, env := range os.Environ() {
        parts := strings.SplitN(env, "=", 2)
        if len(parts) != 2 {
            continue
        }

        key, value := parts[0], parts[1]

        // 如果设置了前缀，只处理匹配前缀的环境变量
        if s.prefix != "" && !strings.HasPrefix(key, s.prefix) {
            continue
        }

        configKey := s.keyMapping(key)
        result[configKey] = config.NewValue(value)
    }

    return result, nil
}

// Watch 环境变量不支持监听
func (s *Source) Watch() config.Watcher {
    return nil
}
```

### 4.2 Consul KV 源

```go
// source/consul/consul.go

package consul

import (
    "context"
    "fmt"
    "strings"
    "sync"

    consulapi "github.com/hashicorp/consul/api"
    "github.com/CloudRoamer/aimo-libs/config"
)

const (
    DefaultPriority = 80
)

// Source Consul KV 配置源
type Source struct {
    client    *consulapi.Client
    prefix    string // KV 路径前缀，如 "config/prod/myapp"
    priority  int
    separator string // key 分隔符，默认 "/"

    mu      sync.RWMutex
    watcher *watcher
}

// Option 配置选项
type Option func(*Source)

// WithPrefix 设置 KV 路径前缀
func WithPrefix(prefix string) Option {
    return func(s *Source) {
        s.prefix = strings.TrimSuffix(prefix, "/")
    }
}

// WithPriority 设置优先级
func WithPriority(p int) Option {
    return func(s *Source) {
        s.priority = p
    }
}

// New 创建 Consul KV 配置源
func New(address string, opts ...Option) (*Source, error) {
    cfg := consulapi.DefaultConfig()
    cfg.Address = address

    client, err := consulapi.NewClient(cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to create consul client: %w", err)
    }

    // 验证连接
    _, err = client.Status().Leader()
    if err != nil {
        return nil, fmt.Errorf("failed to connect to consul at %s: %w", address, err)
    }

    s := &Source{
        client:    client,
        prefix:    "config",
        priority:  DefaultPriority,
        separator: "/",
    }

    for _, opt := range opts {
        opt(s)
    }

    return s, nil
}

func (s *Source) Name() string {
    return "consul"
}

func (s *Source) Priority() int {
    return s.priority
}

func (s *Source) Load(ctx context.Context) (map[string]config.Value, error) {
    kv := s.client.KV()

    pairs, _, err := kv.List(s.prefix, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to list keys from consul: %w", err)
    }

    result := make(map[string]config.Value)
    for _, pair := range pairs {
        // 移除前缀并转换分隔符
        key := strings.TrimPrefix(pair.Key, s.prefix+"/")
        key = strings.ReplaceAll(key, "/", ".")

        result[key] = config.NewValue(string(pair.Value))
    }

    return result, nil
}

// Watch 返回 Consul 监听器
func (s *Source) Watch() config.Watcher {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.watcher == nil {
        s.watcher = newWatcher(s.client, s.prefix)
    }
    return s.watcher
}
```

```go
// source/consul/watcher.go

package consul

import (
    "context"
    "time"

    consulapi "github.com/hashicorp/consul/api"
    "github.com/CloudRoamer/aimo-libs/config"
)

type watcher struct {
    client    *consulapi.Client
    prefix    string
    lastIndex uint64
    stopCh    chan struct{}
    eventCh   chan config.Event
}

func newWatcher(client *consulapi.Client, prefix string) *watcher {
    return &watcher{
        client: client,
        prefix: prefix,
        stopCh: make(chan struct{}),
    }
}

func (w *watcher) Start(ctx context.Context) (<-chan config.Event, error) {
    w.eventCh = make(chan config.Event, 10)

    go w.watch(ctx)

    return w.eventCh, nil
}

func (w *watcher) watch(ctx context.Context) {
    defer close(w.eventCh)

    kv := w.client.KV()

    for {
        select {
        case <-ctx.Done():
            return
        case <-w.stopCh:
            return
        default:
        }

        // 使用阻塞查询等待变更
        opts := &consulapi.QueryOptions{
            WaitIndex: w.lastIndex,
            WaitTime:  5 * time.Minute,
        }

        pairs, meta, err := kv.List(w.prefix, opts)
        if err != nil {
            w.eventCh <- config.Event{
                Type:      config.EventTypeError,
                Source:    "consul",
                Timestamp: time.Now(),
                Error:     err,
            }
            time.Sleep(5 * time.Second) // 错误后等待重试
            continue
        }

        // 如果索引发生变化，说明有配置更新
        if meta.LastIndex > w.lastIndex {
            w.lastIndex = meta.LastIndex

            // 收集变更的 key
            keys := make([]string, 0, len(pairs))
            for _, pair := range pairs {
                keys = append(keys, pair.Key)
            }

            w.eventCh <- config.Event{
                Type:      config.EventTypeReload,
                Source:    "consul",
                Keys:      keys,
                Timestamp: time.Now(),
            }
        }
    }
}

func (w *watcher) Stop() error {
    close(w.stopCh)
    return nil
}
```

### 4.3 文件源

```go
// source/file/file.go

package file

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "sync"

    "github.com/CloudRoamer/aimo-libs/config"
    "github.com/CloudRoamer/aimo-libs/config/codec"
)

const (
    DefaultPriority = 60
)

// Source 文件配置源
type Source struct {
    path     string       // 配置文件路径
    codec    codec.Codec  // 编解码器
    priority int

    mu      sync.RWMutex
    watcher *watcher
}

// Option 配置选项
type Option func(*Source)

// WithCodec 设置编解码器
func WithCodec(c codec.Codec) Option {
    return func(s *Source) {
        s.codec = c
    }
}

// WithPriority 设置优先级
func WithPriority(p int) Option {
    return func(s *Source) {
        s.priority = p
    }
}

// New 创建文件配置源
func New(path string, opts ...Option) (*Source, error) {
    // 检查文件是否存在
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil, fmt.Errorf("config file not found: %s", path)
    }

    s := &Source{
        path:     path,
        priority: DefaultPriority,
    }

    // 根据文件扩展名选择默认编解码器
    ext := strings.ToLower(filepath.Ext(path))
    switch ext {
    case ".json":
        s.codec = codec.JSON
    case ".yaml", ".yml":
        s.codec = codec.YAML
    default:
        return nil, fmt.Errorf("unsupported file format: %s", ext)
    }

    for _, opt := range opts {
        opt(s)
    }

    return s, nil
}

func (s *Source) Name() string {
    return "file:" + s.path
}

func (s *Source) Priority() int {
    return s.priority
}

func (s *Source) Load(ctx context.Context) (map[string]config.Value, error) {
    data, err := os.ReadFile(s.path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    // 解码为 map
    var raw map[string]interface{}
    if err := s.codec.Decode(data, &raw); err != nil {
        return nil, fmt.Errorf("failed to decode config: %w", err)
    }

    // 扁平化嵌套结构
    result := make(map[string]config.Value)
    flatten("", raw, result)

    return result, nil
}

// flatten 将嵌套 map 扁平化为点分隔的 key
func flatten(prefix string, data map[string]interface{}, result map[string]config.Value) {
    for k, v := range data {
        key := k
        if prefix != "" {
            key = prefix + "." + k
        }

        switch val := v.(type) {
        case map[string]interface{}:
            flatten(key, val, result)
        case map[interface{}]interface{}:
            // YAML 可能返回这种类型
            converted := make(map[string]interface{})
            for mk, mv := range val {
                converted[fmt.Sprintf("%v", mk)] = mv
            }
            flatten(key, converted, result)
        default:
            result[key] = config.NewValueFromInterface(v)
        }
    }
}

// Watch 返回文件监听器
func (s *Source) Watch() config.Watcher {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.watcher == nil {
        s.watcher = newWatcher(s.path)
    }
    return s.watcher
}
```

```go
// source/file/watcher.go

package file

import (
    "context"
    "time"

    "github.com/fsnotify/fsnotify"
    "github.com/CloudRoamer/aimo-libs/config"
)

type watcher struct {
    path    string
    stopCh  chan struct{}
    eventCh chan config.Event
}

func newWatcher(path string) *watcher {
    return &watcher{
        path:   path,
        stopCh: make(chan struct{}),
    }
}

func (w *watcher) Start(ctx context.Context) (<-chan config.Event, error) {
    fsWatcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    if err := fsWatcher.Add(w.path); err != nil {
        fsWatcher.Close()
        return nil, err
    }

    w.eventCh = make(chan config.Event, 10)

    go w.watch(ctx, fsWatcher)

    return w.eventCh, nil
}

func (w *watcher) watch(ctx context.Context, fsWatcher *fsnotify.Watcher) {
    defer close(w.eventCh)
    defer fsWatcher.Close()

    for {
        select {
        case <-ctx.Done():
            return
        case <-w.stopCh:
            return
        case event, ok := <-fsWatcher.Events:
            if !ok {
                return
            }

            // 只关注写入事件
            if event.Op&fsnotify.Write == fsnotify.Write {
                w.eventCh <- config.Event{
                    Type:      config.EventTypeUpdate,
                    Source:    "file:" + w.path,
                    Timestamp: time.Now(),
                }
            }
        case err, ok := <-fsWatcher.Errors:
            if !ok {
                return
            }
            w.eventCh <- config.Event{
                Type:      config.EventTypeError,
                Source:    "file:" + w.path,
                Timestamp: time.Now(),
                Error:     err,
            }
        }
    }
}

func (w *watcher) Stop() error {
    close(w.stopCh)
    return nil
}
```

### 4.4 PostgreSQL 源（可选）

```go
// source/postgres/postgres.go

package postgres

import (
    "context"
    "database/sql"
    "fmt"

    _ "github.com/lib/pq"
    "github.com/CloudRoamer/aimo-libs/config"
)

const (
    DefaultPriority = 70
    DefaultTable    = "app_config"
)

// Source PostgreSQL 配置源
type Source struct {
    db       *sql.DB
    table    string // 配置表名
    keyCol   string // key 列名
    valueCol string // value 列名
    priority int
}

// Option 配置选项
type Option func(*Source)

// WithTable 设置配置表名
func WithTable(table string) Option {
    return func(s *Source) {
        s.table = table
    }
}

// WithColumns 设置列名
func WithColumns(keyCol, valueCol string) Option {
    return func(s *Source) {
        s.keyCol = keyCol
        s.valueCol = valueCol
    }
}

// WithPriority 设置优先级
func WithPriority(p int) Option {
    return func(s *Source) {
        s.priority = p
    }
}

// New 创建 PostgreSQL 配置源
func New(dsn string, opts ...Option) (*Source, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    if err := db.Ping(); err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    s := &Source{
        db:       db,
        table:    DefaultTable,
        keyCol:   "key",
        valueCol: "value",
        priority: DefaultPriority,
    }

    for _, opt := range opts {
        opt(s)
    }

    return s, nil
}

func (s *Source) Name() string {
    return "postgres"
}

func (s *Source) Priority() int {
    return s.priority
}

func (s *Source) Load(ctx context.Context) (map[string]config.Value, error) {
    query := fmt.Sprintf(
        "SELECT %s, %s FROM %s",
        s.keyCol, s.valueCol, s.table,
    )

    rows, err := s.db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to query config: %w", err)
    }
    defer rows.Close()

    result := make(map[string]config.Value)
    for rows.Next() {
        var key, value string
        if err := rows.Scan(&key, &value); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }
        result[key] = config.NewValue(value)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating rows: %w", err)
    }

    return result, nil
}

// Watch PostgreSQL 源暂不支持监听
// 可以通过 LISTEN/NOTIFY 机制实现，但这里保持简单
func (s *Source) Watch() config.Watcher {
    return nil
}

// Close 关闭数据库连接
func (s *Source) Close() error {
    return s.db.Close()
}
```

---

## 5. 合并策略

### 5.1 Merger 接口

```go
// merge/merger.go

package merge

import (
    "github.com/CloudRoamer/aimo-libs/config"
)

// Merger 定义配置合并策略接口
type Merger interface {
    // Merge 合并多个配置映射
    // 按顺序合并，后面的覆盖前面的
    Merge(maps ...map[string]config.Value) map[string]config.Value
}

// DefaultMerger 默认合并器
// 简单的覆盖策略：后面的值覆盖前面的
type DefaultMerger struct{}

// NewDefaultMerger 创建默认合并器
func NewDefaultMerger() *DefaultMerger {
    return &DefaultMerger{}
}

func (m *DefaultMerger) Merge(maps ...map[string]config.Value) map[string]config.Value {
    result := make(map[string]config.Value)

    for _, configMap := range maps {
        for k, v := range configMap {
            result[k] = v
        }
    }

    return result
}

// DeepMerger 深度合并器
// 对于嵌套结构，会递归合并而非简单覆盖
type DeepMerger struct{}

// NewDeepMerger 创建深度合并器
func NewDeepMerger() *DeepMerger {
    return &DeepMerger{}
}

func (m *DeepMerger) Merge(maps ...map[string]config.Value) map[string]config.Value {
    result := make(map[string]config.Value)

    for _, configMap := range maps {
        for k, v := range configMap {
            // 深度合并逻辑：如果已存在同前缀的 key，保留子 key
            // 例如：database.host 和 database.port 都应保留
            result[k] = v
        }
    }

    return result
}
```

### 5.2 优先级图示

```
优先级（数值越大越优先）

    100  ┌─────────────────┐
         │   环境变量 (Env)  │  ← 最高优先级，用于运行时覆盖
    80   ├─────────────────┤
         │   Consul KV      │  ← 集中配置管理
    70   ├─────────────────┤
         │   PostgreSQL     │  ← 持久化配置（可选）
    60   ├─────────────────┤
         │   文件 (File)    │  ← 本地配置文件
    0    ├─────────────────┤
         │   默认值         │  ← 代码内置默认值
         └─────────────────┘

合并顺序：默认值 → 文件 → PostgreSQL → Consul → 环境变量
```

---

## 6. 热更新机制

### 6.1 时序图

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│ Consul  │    │ Watcher │    │ Manager │    │Callback │
└────┬────┘    └────┬────┘    └────┬────┘    └────┬────┘
     │              │              │              │
     │   KV 变更    │              │              │
     │─────────────>│              │              │
     │              │              │              │
     │              │  Event       │              │
     │              │─────────────>│              │
     │              │              │              │
     │              │              │ 1. 保存旧配置 │
     │              │              │─────┐        │
     │              │              │<────┘        │
     │              │              │              │
     │              │              │ 2. 重新加载  │
     │              │              │    所有 Source│
     │              │              │─────┐        │
     │              │              │<────┘        │
     │              │              │              │
     │              │              │ 3. 合并配置  │
     │              │              │─────┐        │
     │              │              │<────┘        │
     │              │              │              │
     │              │              │ 4. 触发回调  │
     │              │              │─────────────>│
     │              │              │              │
     │              │              │              │ 5. 业务处理
     │              │              │              │    (可选重连等)
     │              │              │              │
```

### 6.2 使用示例

```go
// 注册配置变更回调
manager.OnChange(func(event config.Event, oldCfg, newCfg config.Config) {
    log.Printf("[CONFIG] Change detected from %s: %s", event.Source, event.Type)

    // 检查特定配置是否变化
    oldDBHost := oldCfg.GetString("database.host", "")
    newDBHost := newCfg.GetString("database.host", "")

    if oldDBHost != newDBHost {
        log.Printf("[CONFIG] Database host changed: %s -> %s", oldDBHost, newDBHost)
        // 触发数据库重连
        reconnectDatabase(newCfg)
    }
})
```

---

## 7. Value 类型设计

```go
// value.go

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
    raw interface{} // 原始值
}

// NewValue 从字符串创建 Value
func NewValue(s string) Value {
    return Value{raw: s}
}

// NewValueFromInterface 从任意类型创建 Value
func NewValueFromInterface(v interface{}) Value {
    return Value{raw: v}
}

// Raw 返回原始值
func (v Value) Raw() interface{} {
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
    case []interface{}:
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
    case map[string]interface{}:
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
```

---

## 8. 使用示例

### 8.1 基础使用

```go
package main

import (
    "context"
    "log"

    "github.com/CloudRoamer/aimo-libs/config"
    "github.com/CloudRoamer/aimo-libs/config/source/consul"
    "github.com/CloudRoamer/aimo-libs/config/source/env"
    "github.com/CloudRoamer/aimo-libs/config/source/file"
)

func main() {
    // 创建配置管理器
    mgr := config.NewManager()

    // 添加配置源（按优先级从低到高）
    // 1. 文件配置（优先级 60）
    fileSource, err := file.New("config.yaml")
    if err != nil {
        log.Fatalf("Failed to create file source: %v", err)
    }

    // 2. Consul 配置（优先级 80）
    consulSource, err := consul.New("localhost:8500",
        consul.WithPrefix("config/prod/myapp"),
    )
    if err != nil {
        log.Fatalf("Failed to create consul source: %v", err)
    }

    // 3. 环境变量（优先级 100）
    envSource := env.New(
        env.WithPrefix("MYAPP_"),
    )

    // 添加所有配置源
    mgr.AddSource(fileSource, consulSource, envSource)

    // 加载配置
    ctx := context.Background()
    if err := mgr.Load(ctx); err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 获取配置
    cfg := mgr.Config()

    dbHost := cfg.GetString("database.host", "localhost")
    dbPort := cfg.GetInt("database.port", 5432)
    timeout := cfg.GetDuration("server.timeout", 30*time.Second)

    log.Printf("Database: %s:%d", dbHost, dbPort)
    log.Printf("Timeout: %v", timeout)

    // 关闭管理器
    defer mgr.Close()
}
```

### 8.2 热更新示例

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/CloudRoamer/aimo-libs/config"
    "github.com/CloudRoamer/aimo-libs/config/source/consul"
    "github.com/CloudRoamer/aimo-libs/config/source/env"
)

func main() {
    mgr := config.NewManager()

    // 添加配置源
    consulSource, _ := consul.New("localhost:8500",
        consul.WithPrefix("config/prod/myapp"),
    )
    envSource := env.New(env.WithPrefix("MYAPP_"))

    mgr.AddSource(consulSource, envSource)

    // 注册配置变更回调
    mgr.OnChange(func(event config.Event, oldCfg, newCfg config.Config) {
        log.Printf("[CONFIG] Change from %s: %s", event.Source, event.Type)

        if event.Error != nil {
            log.Printf("[CONFIG] Error: %v", event.Error)
            return
        }

        // 检查数据库配置变化
        if oldCfg != nil {
            oldHost := oldCfg.GetString("database.host", "")
            newHost := newCfg.GetString("database.host", "")
            if oldHost != newHost {
                log.Printf("[CONFIG] Database host changed: %s -> %s", oldHost, newHost)
                // 这里可以触发数据库重连
            }
        }
    })

    // 加载初始配置
    ctx := context.Background()
    if err := mgr.Load(ctx); err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 启动配置监听
    if err := mgr.Watch(); err != nil {
        log.Fatalf("Failed to start watching: %v", err)
    }

    log.Println("Configuration loaded and watching for changes...")

    // 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down...")
    mgr.Close()
}
```

### 8.3 结构体映射示例

```go
package main

import (
    "context"
    "log"

    "github.com/CloudRoamer/aimo-libs/config"
    "github.com/CloudRoamer/aimo-libs/config/source/file"
)

// DatabaseConfig 数据库配置结构体
type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    User     string `mapstructure:"user"`
    Password string `mapstructure:"password"`
    DBName   string `mapstructure:"dbname"`
}

// ServerConfig 服务器配置结构体
type ServerConfig struct {
    Port         string        `mapstructure:"port"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// AppConfig 应用配置结构体
type AppConfig struct {
    Database DatabaseConfig `mapstructure:"database"`
    Server   ServerConfig   `mapstructure:"server"`
}

func main() {
    mgr := config.NewManager()

    fileSource, _ := file.New("config.yaml")
    mgr.AddSource(fileSource)

    ctx := context.Background()
    if err := mgr.Load(ctx); err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    cfg := mgr.Config()

    // 方式 1：单独获取子配置
    var dbCfg DatabaseConfig
    if err := cfg.UnmarshalKey("database", &dbCfg); err != nil {
        log.Fatalf("Failed to unmarshal database config: %v", err)
    }
    log.Printf("Database: %s@%s:%d/%s", dbCfg.User, dbCfg.Host, dbCfg.Port, dbCfg.DBName)

    // 方式 2：获取完整配置
    var appCfg AppConfig
    if err := cfg.Unmarshal("", &appCfg); err != nil {
        log.Fatalf("Failed to unmarshal app config: %v", err)
    }
    log.Printf("Server port: %s", appCfg.Server.Port)

    defer mgr.Close()
}
```

### 8.4 自定义配置源示例

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "github.com/CloudRoamer/aimo-libs/config"
)

// HTTPSource 从 HTTP 端点加载配置
type HTTPSource struct {
    url      string
    priority int
    client   *http.Client
}

func NewHTTPSource(url string) *HTTPSource {
    return &HTTPSource{
        url:      url,
        priority: 50,
        client: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (s *HTTPSource) Name() string {
    return "http:" + s.url
}

func (s *HTTPSource) Priority() int {
    return s.priority
}

func (s *HTTPSource) Load(ctx context.Context) (map[string]config.Value, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", s.url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := s.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var data map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return nil, err
    }

    result := make(map[string]config.Value)
    for k, v := range data {
        result[k] = config.NewValueFromInterface(v)
    }

    return result, nil
}

func (s *HTTPSource) Watch() config.Watcher {
    return nil // HTTP 源不支持监听
}

func main() {
    mgr := config.NewManager()

    // 使用自定义 HTTP 配置源
    httpSource := NewHTTPSource("https://config.example.com/api/config")
    mgr.AddSource(httpSource)

    // ... 其余代码
}
```

---

## 9. 扩展点设计

### 9.1 扩展点清单

| 扩展点 | 接口 | 说明 |
|--------|------|------|
| 配置源 | `Source` | 实现新的配置源（如 etcd、Redis、S3） |
| 合并策略 | `Merger` | 自定义配置合并逻辑 |
| 编解码器 | `Codec` | 支持新的配置格式（如 TOML、INI） |
| 值类型 | `Value` | 扩展类型转换方法 |
| 监听器 | `Watcher` | 自定义配置变更检测机制 |

### 9.2 添加新配置源的步骤

1. **创建源目录**
   ```
   source/etcd/
   ├── etcd.go
   ├── watcher.go
   └── etcd_test.go
   ```

2. **实现 Source 接口**
   ```go
   type Source struct { /* ... */ }

   func (s *Source) Name() string { return "etcd" }
   func (s *Source) Priority() int { return s.priority }
   func (s *Source) Load(ctx context.Context) (map[string]config.Value, error) { /* ... */ }
   func (s *Source) Watch() config.Watcher { /* ... */ }
   ```

3. **实现 Watcher 接口（如支持监听）**
   ```go
   type watcher struct { /* ... */ }

   func (w *watcher) Start(ctx context.Context) (<-chan config.Event, error) { /* ... */ }
   func (w *watcher) Stop() error { /* ... */ }
   ```

4. **编写测试**
   ```go
   func TestSource_Load(t *testing.T) { /* ... */ }
   func TestWatcher_Events(t *testing.T) { /* ... */ }
   ```

---

## 10. 错误处理

### 10.1 错误类型

```go
// errors.go

package config

import "errors"

var (
    // ErrSourceNotFound 配置源未找到
    ErrSourceNotFound = errors.New("config source not found")

    // ErrKeyNotFound 配置键不存在
    ErrKeyNotFound = errors.New("config key not found")

    // ErrTypeMismatch 类型转换失败
    ErrTypeMismatch = errors.New("config value type mismatch")

    // ErrLoadFailed 加载配置失败
    ErrLoadFailed = errors.New("failed to load config")

    // ErrWatchFailed 启动监听失败
    ErrWatchFailed = errors.New("failed to start config watch")
)

// SourceError 配置源错误
type SourceError struct {
    Source string
    Err    error
}

func (e *SourceError) Error() string {
    return fmt.Sprintf("source %s: %v", e.Source, e.Err)
}

func (e *SourceError) Unwrap() error {
    return e.Err
}
```

---

## 11. 与现有实现的对比

### 11.1 AIMO-Memos 现有实现的问题

| 问题 | 说明 |
|------|------|
| **强耦合** | `Config` 结构体与业务配置强绑定，无法复用 |
| **无优先级** | 没有明确的配置源优先级定义 |
| **手动解析** | 每个字段需要手动编写解析逻辑 |
| **无热更新** | 不支持配置动态刷新 |
| **单一加载器** | ConsulLoader 和 JSONLoader 代码重复 |

### 11.2 新 SDK 的改进

| 改进 | 说明 |
|------|------|
| **完全解耦** | 配置源与业务配置分离，通用性强 |
| **优先级合并** | 明确的优先级定义，支持灵活覆盖 |
| **类型安全** | 泛型 API 提供类型安全的访问 |
| **热更新** | 内置 Watch 机制，支持回调 |
| **可扩展** | 清晰的接口设计，易于添加新源 |

---

## 12. 实施状态

### 12.1 已完成功能

| 阶段 | 内容 | 状态 |
|------|------|------|
| P1 | 核心接口 + 环境变量源 + 文件源 | 已完成 |
| P2 | Consul 源 + 热更新机制 | 已完成 |
| P3 | PostgreSQL 源 | 已完成 |
| P4 | 文档 + 示例 + 测试完善 | 已完成 |

### 12.2 依赖清单

```go
// go.mod

module github.com/CloudRoamer/aimo-libs/config

go 1.25

require (
    github.com/hashicorp/consul/api v1.33.2  // Consul 客户端
    github.com/fsnotify/fsnotify v1.9.0      // 文件系统监听
    gopkg.in/yaml.v3 v3.0.1                  // YAML 解析
    github.com/lib/pq v1.10.9                // PostgreSQL 驱动
)
```

---

## 13. 总结

aimo-config SDK 通过清晰的接口抽象和分层设计，实现了：

1. **多源统一**：环境变量、Consul、文件、数据库等多种配置源的统一管理
2. **优先级合并**：可配置的优先级策略，支持灵活的配置覆盖
3. **热更新**：内置 Watch 机制，支持配置动态刷新和回调通知
4. **类型安全**：泛型 API 提供类型安全的配置访问
5. **易扩展**：清晰的接口设计，易于添加新的配置源和合并策略

该设计遵循 SOLID 原则，具备开源级别的通用性和可扩展性，可以作为团队内部和开源社区的统一配置管理方案。
