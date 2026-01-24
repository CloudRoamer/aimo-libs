# aimo-config

通用的统一配置管理 SDK，支持多配置源、优先级合并和热更新。

当前仅支持以 WASM 形式使用，示例位于 `config/examples/json` 与 `config/examples/yaml` 。

## 特性

- **多源支持**：环境变量、配置文件（JSON/YAML）、Consul KV、PostgreSQL
- **优先级合并**：环境变量 > Consul > PostgreSQL > 文件 > 默认值
- **类型安全**：提供类型安全的配置访问 API
- **热更新**：Watch 机制支持配置动态刷新
- **易扩展**：清晰的接口设计，易于添加新的配置源

## 使用方式

通过本仓库构建 WASM 模块后在应用中加载使用。

## 快速开始

### WASM 构建

```bash
task build:wasm
```

### 运行示例

```bash
task example:run EXAMPLE=yaml
```

### 配置文件示例

**config.yaml**

```yaml
database:
  host: localhost
  port: 5432
  user: postgres

server:
  port: 8080
  timeout: 30s
```

### 环境变量覆盖

```bash
# 环境变量会覆盖文件配置
export APP_DATABASE_HOST=prod.db.com
export APP_DATABASE_PORT=3306

# 运行程序
go run main.go
# 输出: Database: prod.db.com:3306
```

---

## 配置源

### 环境变量

```go
import "github.com/CloudRoamer/aimo-libs/config/source/env"

// 基础用法
envSource := env.New()

// 设置前缀
envSource := env.New(
    env.WithPrefix("MYAPP_"),  // 只读取 MYAPP_ 开头的变量
)

// 自定义映射
envSource := env.New(
    env.WithPrefix("APP_"),
    env.WithKeyMapping(func(key string) string {
        // 自定义环境变量到配置键的映射逻辑
        return strings.ToLower(key)
    }),
)
```

**环境变量命名规则**：

- `APP_DATABASE_HOST` -> `database.host`
- `APP_SERVER_PORT` -> `server.port`

**优先级**：100（最高）

### 文件配置

```go
import "github.com/CloudRoamer/aimo-libs/config/source/file"

// JSON 文件
jsonSource, _ := file.New("config.json")

// YAML 文件
yamlSource, _ := file.New("config.yaml")

// 自定义优先级
fileSource, _ := file.New("config.yaml",
    file.WithPriority(80),  // 默认 60
)
```

**支持的格式**：

- JSON（`.json`）
- YAML（`.yaml`, `.yml`）

**优先级**：60

### Consul KV

```go
import "github.com/CloudRoamer/aimo-libs/config/source/consul"

// 基础用法
consulSource, err := consul.New("localhost:8500")
if err != nil {
    log.Fatal(err)
}

// 设置 key 前缀
consulSource, _ := consul.New("localhost:8500",
    consul.WithPrefix("config/prod/myapp"),
)

// 设置 ACL Token
consulSource, _ := consul.New("localhost:8500",
    consul.WithPrefix("config/prod/myapp"),
    consul.WithToken("your-acl-token"),
)
```

**Consul 中的 key 会自动转换**：

- `config/prod/myapp/database/host` -> `database.host`

**优先级**：80

### PostgreSQL

```go
import "github.com/CloudRoamer/aimo-libs/config/source/postgres"

// 基础用法
pgSource, err := postgres.New("postgres://user:pass@localhost/dbname?sslmode=disable")
if err != nil {
    log.Fatal(err)
}

// 自定义表名和列名
pgSource, _ := postgres.New(dsn,
    postgres.WithTable("app_config"),
    postgres.WithColumns("config_key", "config_value"),
)
```

**默认表结构**：

```sql
CREATE TABLE app_config (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL
);
```

**优先级**：70

---

## 配置优先级

配置源按优先级合并，数值越大优先级越高：

| 配置源 | 优先级 | 说明 |
|--------|--------|------|
| 环境变量 | 100 | 最高优先级，用于运行时覆盖 |
| Consul KV | 80 | 集中配置管理 |
| PostgreSQL | 70 | 持久化配置存储 |
| 文件 | 60 | 本地配置文件 |
| 默认值 | 0 | 代码内置默认值 |

**合并规则**：

```yaml
# config.yaml
database:
  host: localhost
  port: 5432
```

```bash
# 环境变量
export APP_DATABASE_HOST=prod.db.com
```

```go
cfg := mgr.Config()
// database.host = prod.db.com (环境变量覆盖)
// database.port = 5432 (文件配置保留)
```

---

## Watch 热更新

### 基础用法

```go
// 注册配置变更回调
mgr.OnChange(func(event config.Event, oldCfg, newCfg config.Config) {
    log.Printf("配置变更: 来源=%s, 类型=%s", event.Source, event.Type)

    if event.Error != nil {
        log.Printf("错误: %v", event.Error)
        return
    }

    // 检查具体配置变化
    oldHost := oldCfg.GetString("database.host", "")
    newHost := newCfg.GetString("database.host", "")
    if oldHost != newHost {
        log.Printf("数据库主机变更: %s -> %s", oldHost, newHost)
        // 触发数据库重连等操作
    }
})

// 启动配置监听
if err := mgr.Watch(); err != nil {
    log.Fatal(err)
}
```

### 支持 Watch 的配置源

| 配置源 | Watch 支持 | 说明 |
|--------|------------|------|
| 文件 | 支持 | 基于 fsnotify 监听文件变更 |
| Consul | 支持 | 基于 blocking query 长轮询 |
| 环境变量 | 不支持 | 环境变量在进程启动后无法改变 |
| PostgreSQL | 不支持 | 可通过 LISTEN/NOTIFY 扩展 |

### 完整示例

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
    "github.com/CloudRoamer/aimo-libs/config/source/file"
)

func main() {
    mgr := config.NewManager()

    // 添加配置源
    fileSource, _ := file.New("config.yaml")
    consulSource, _ := consul.New("localhost:8500",
        consul.WithPrefix("config/prod/myapp"),
    )
    envSource := env.New(env.WithPrefix("APP_"))

    mgr.AddSource(fileSource, consulSource, envSource)

    // 注册变更回调
    mgr.OnChange(func(event config.Event, oldCfg, newCfg config.Config) {
        log.Printf("配置变更: %s", event.Type)
    })

    // 加载配置
    ctx := context.Background()
    if err := mgr.Load(ctx); err != nil {
        log.Fatal(err)
    }

    // 启动监听
    if err := mgr.Watch(); err != nil {
        log.Fatal(err)
    }

    log.Println("配置监听已启动...")

    // 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    mgr.Close()
}
```

---

## 类型转换

Config 接口提供丰富的类型转换方法：

```go
cfg := mgr.Config()

// 字符串
host := cfg.GetString("database.host", "localhost")

// 整数
port := cfg.GetInt("database.port", 5432)
port64 := cfg.GetInt64("database.max_connections", 100)

// 浮点数
ratio := cfg.GetFloat64("cache.hit_ratio", 0.95)

// 布尔值
debug := cfg.GetBool("debug", false)

// 时间间隔
timeout := cfg.GetDuration("server.timeout", 30*time.Second)

// 字符串切片
hosts := cfg.GetStringSlice("redis.hosts", []string{"localhost"})

// 字符串映射
labels := cfg.GetStringMap("app.labels", nil)
```

**类型转换规则**：

- 支持自动类型转换（如字符串 "123" -> int 123）
- 转换失败时返回默认值
- 支持 JSON 格式解析（如切片、映射）

---

## 配置检查

```go
// 检查键是否存在
if cfg.Has("database.host") {
    // 键存在
}

// 获取原始值
if val, ok := cfg.Get("database.host"); ok {
    fmt.Println(val.Raw())     // 原始值
    fmt.Println(val.String())  // 字符串表示
}

// 列出所有键
keys := cfg.Keys()
for _, key := range keys {
    fmt.Println(key)
}
```

---

## API 参考

### Manager

| 方法 | 说明 |
|------|------|
| `NewManager()` | 创建配置管理器 |
| `AddSource(sources ...Source)` | 添加配置源 |
| `Load(ctx context.Context)` | 加载所有配置源 |
| `Watch()` | 启动配置监听 |
| `OnChange(callback)` | 注册配置变更回调 |
| `Config()` | 获取当前配置 |
| `Close()` | 关闭管理器 |

### Config

| 方法 | 说明 |
|------|------|
| `Get(key)` | 获取原始配置值 |
| `GetString(key, default)` | 获取字符串值 |
| `GetInt(key, default)` | 获取整数值 |
| `GetInt64(key, default)` | 获取 int64 值 |
| `GetFloat64(key, default)` | 获取浮点数值 |
| `GetBool(key, default)` | 获取布尔值 |
| `GetDuration(key, default)` | 获取时间间隔 |
| `GetStringSlice(key, default)` | 获取字符串切片 |
| `GetStringMap(key, default)` | 获取字符串映射 |
| `Keys()` | 返回所有配置键 |
| `Has(key)` | 检查键是否存在 |

### Source 接口

```go
type Source interface {
    Name() string
    Priority() int
    Load(ctx context.Context) (map[string]Value, error)
    Watch() Watcher
}
```

### Watcher 接口

```go
type Watcher interface {
    Start(ctx context.Context) (<-chan Event, error)
    Stop() error
}
```

---

## 示例代码

WASM 示例请参考 `config/examples/json` 与 `config/examples/yaml` 。

---

## 架构设计

详细的架构设计文档请参考 [ARCHITECTURE.md](./ARCHITECTURE.md)

---

## 测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定包测试
go test -v ./source/env
go test -v ./source/file
go test -v ./source/consul
go test -v ./source/postgres

# 跳过需要外部依赖的测试
go test -short ./...
```

---

## 许可证

商业授权 - 芜湖图忆科技有限公司
