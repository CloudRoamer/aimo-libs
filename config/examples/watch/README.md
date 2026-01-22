# 配置热更新示例

本示例展示如何使用 aimo-config 的配置热更新功能。

## 功能特性

- 支持多配置源：文件、Consul、环境变量
- 配置优先级：环境变量 > Consul > 文件
- 配置热更新：文件变更自动重载
- 变更回调：监听配置变化并执行自定义逻辑

## 运行方式

### 基本运行（仅文件配置源）

```bash
cd examples/watch
go run main.go
```

### 使用 Consul 配置源

```bash
# 设置 Consul 地址
export CONSUL_ADDR=localhost:8500

# 运行示例
go run main.go
```

### 使用环境变量覆盖

```bash
# 通过环境变量覆盖配置
export APP_DATABASE_HOST=192.168.1.100
export APP_DATABASE_PORT=3306

go run main.go
```

## 测试配置变更

### 测试文件配置热更新

1. 启动程序后，编辑 `config.yaml` 文件
2. 修改 `database.host` 或 `database.port` 的值
3. 保存文件
4. 观察程序输出的变更日志

### 测试 Consul 配置热更新

1. 确保 Consul 服务运行
2. 在 Consul KV 中设置配置：
   ```bash
   consul kv put config/dev/myapp/database/host localhost
   consul kv put config/dev/myapp/database/port 5432
   ```
3. 启动程序
4. 修改 Consul KV 中的值：
   ```bash
   consul kv put config/dev/myapp/database/host 192.168.1.100
   ```
5. 观察程序输出的变更日志

## 预期输出

```
已添加文件配置源
已添加环境变量配置源
初始配置加载成功

========== 当前配置 ==========
database.host = localhost
database.port = 5432
database.user = postgres
database.password = secret
server.port = 8080
server.timeout = 30s
=============================

配置监听已启动，等待变更...
应用运行中... 当前数据库: localhost:5432

配置变更事件: 来源=file:config.yaml, 类型=update
数据库主机变更: localhost -> 192.168.1.100
变更的配置项: [database.host database.port database.user database.password server.port server.timeout]
应用运行中... 当前数据库: 192.168.1.100:5432
```

## 注意事项

1. Consul 配置源是可选的，如果未设置 `CONSUL_ADDR` 环境变量，程序将仅使用文件和环境变量配置源
2. 文件配置源会监听文件变更，但需要确保文件系统支持 inotify/kqueue
3. 环境变量配置源不支持热更新（环境变量在进程启动后无法改变）
4. 配置变更回调在单独的 goroutine 中执行，注意并发安全
