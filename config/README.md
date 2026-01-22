# aimo-config

通用的统一配置管理 SDK，支持多配置源、优先级合并和热更新。

## 特性

- **多源支持**: 环境变量、配置文件（JSON/YAML）、Consul KV（即将支持）
- **优先级合并**: 环境变量 > Consul > 文件 > 默认值
- **类型安全**: 提供类型安全的配置访问 API
- **热更新**: 支持配置动态刷新（即将支持）
- **易扩展**: 清晰的接口设计，易于添加新的配置源

## 安装

```bash
go get github.com/CloudRoamer/aimo-libs/config
```

## 快速开始

### 基础使用

```go
package main

import (
    "context"
    "log"

    "github.com/CloudRoamer/aimo-libs/config"
    "github.com/CloudRoamer/aimo-libs/config/source/env"
    "github.com/CloudRoamer/aimo-libs/config/source/file"
)

func main() {
    // 创建配置管理器
    mgr := config.NewManager()

    // 添加配置源
    fileSource, _ := file.New("config.yaml")
    envSource := env.New(env.WithPrefix("APP_"))

    mgr.AddSource(fileSource, envSource)

    // 加载配置
    ctx := context.Background()
    if err := mgr.Load(ctx); err != nil {
        log.Fatal(err)
    }

    // 获取配置
    cfg := mgr.Config()
    dbHost := cfg.GetString("database.host", "localhost")
    dbPort := cfg.GetInt("database.port", 5432)

    log.Printf("Database: %s:%d", dbHost, dbPort)

    defer mgr.Close()
}
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
- `APP_DATABASE_HOST` → `database.host`
- `APP_SERVER_PORT` → `server.port`

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
- JSON (`.json`)
- YAML (`.yaml`, `.yml`)

## 配置优先级

配置源按优先级合并，数值越大优先级越高：

| 配置源 | 优先级 | 说明 |
|--------|--------|------|
| 环境变量 | 100 | 最高优先级，用于运行时覆盖 |
| Consul KV | 80 | 集中配置管理（即将支持）|
| 文件 | 60 | 本地配置文件 |
| 默认值 | 0 | 代码内置默认值 |

**示例**：
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
- 支持自动类型转换（如字符串 "123" → int 123）
- 转换失败时返回默认值
- 支持 JSON 格式解析（如切片、映射）

## 配置检查

```go
// 检查键是否存在
if cfg.Has("database.host") {
    // 键存在
}

// 获取原始值
if val, ok := cfg.Get("database.host"); ok {
    fmt.Println(val.Raw())  // 原始值
    fmt.Println(val.String())  // 字符串表示
}

// 列出所有键
keys := cfg.Keys()
for _, key := range keys {
    fmt.Println(key)
}
```

## 架构设计

详细的架构设计文档请参考 [ARCHITECTURE.md](./ARCHITECTURE.md)

### 核心接口

```go
// Source 配置源接口
type Source interface {
    Name() string
    Priority() int
    Load(ctx context.Context) (map[string]Value, error)
    Watch() Watcher  // 可选：配置监听
}

// Config 配置访问接口
type Config interface {
    Get(key string) (Value, bool)
    GetString(key string, defaultVal string) string
    GetInt(key string, defaultVal int) int
    // ... 更多类型方法
}
```

## 运行示例

```bash
# 进入示例目录
cd examples/basic

# 运行示例
go run main.go

# 使用环境变量覆盖
APP_DATABASE_HOST=prod.db.com go run main.go
```

## 测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定包测试
go test -v ./source/env
go test -v ./source/file
```

## 下一步计划

- [ ] Consul KV 配置源
- [ ] PostgreSQL 配置源
- [ ] 热更新机制（Watch）
- [ ] 配置加密支持
- [ ] 更多示例

## 许可证

商业授权 - 芜湖图忆科技有限公司

## 参与贡献

欢迎提交 Issue 和 Pull Request。
