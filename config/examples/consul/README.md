# Consul 示例

本示例演示通过 Consul 配置源读取配置。该示例为本地 Go 方式运行，用于展示 Consul 配置源行为。

## 前置条件

- 本地可访问 Consul：`localhost:8500`
- 已写入示例配置数据

## 写入示例数据

```bash
CONSUL_PREFIX=config/examples/consul

consul kv put ${CONSUL_PREFIX}/app/name example-app
consul kv put ${CONSUL_PREFIX}/app/version 1.0.0
consul kv put ${CONSUL_PREFIX}/server/host 0.0.0.0
consul kv put ${CONSUL_PREFIX}/server/port 8080
consul kv put ${CONSUL_PREFIX}/server/timeout 30s
consul kv put ${CONSUL_PREFIX}/debug false
```

## 运行示例

```bash
CONSUL_ADDR=localhost:8500 \
CONSUL_PREFIX=config/examples/consul \
go run main.go
```

## 可选环境变量

```bash
# 覆盖配置（前缀 APP_）
export APP_SERVER_PORT=9090
```
