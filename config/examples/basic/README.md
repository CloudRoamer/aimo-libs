# 基础使用示例

本示例展示 aimo-config 的基础使用方法。

## 功能演示

- 从 YAML 配置文件加载配置
- 使用环境变量覆盖配置
- 类型安全的配置访问
- 列出所有配置键

## 运行方式

### 基本运行

```bash
cd examples/basic
go run main.go
```

### 使用环境变量覆盖

```bash
# 设置环境变量（带 APP_ 前缀）
export APP_DATABASE_HOST=192.168.1.100
export APP_DATABASE_PORT=3306

# 运行示例
go run main.go
```

## 配置文件

示例使用 `config.yaml` 作为配置文件：

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: secret

server:
  port: 8080
  timeout: 30s
  max_connections: 100
```

## 预期输出

### 仅文件配置

```
Database: localhost:5432
Timeout: 30s

All config keys:
  database.host = localhost
  database.password = secret
  database.port = 5432
  database.user = postgres
  server.max_connections = 100
  server.port = 8080
  server.timeout = 30s
```

### 使用环境变量覆盖后

```
Database: 192.168.1.100:3306
Timeout: 30s

All config keys:
  database.host = 192.168.1.100
  database.password = secret
  database.port = 3306
  database.user = postgres
  server.max_connections = 100
  server.port = 8080
  server.timeout = 30s
```

## 代码说明

```go
// 创建配置管理器
mgr := config.NewManager()

// 添加文件配置源（优先级 60）
fileSource, _ := file.New("config.yaml")

// 添加环境变量配置源（优先级 100，会覆盖文件配置）
envSource := env.New(env.WithPrefix("APP_"))

// 添加配置源
mgr.AddSource(fileSource, envSource)

// 加载配置
mgr.Load(ctx)

// 获取配置值
cfg := mgr.Config()
dbHost := cfg.GetString("database.host", "localhost")
dbPort := cfg.GetInt("database.port", 5432)
```

## 配置优先级

本示例中的配置优先级为：

1. **环境变量**（优先级 100）- 最高优先级
2. **文件配置**（优先级 60）- 作为基础配置

环境变量会覆盖文件配置中的同名配置项。
