# JSON 示例

本示例演示通过 Go Plugin 加载 Config SDK，并使用 JSON 配置文件。

## 构建 Plugin

```bash
# macOS
task build:plugin:darwin

# Linux
task build:plugin:linux
```

## 运行示例

```bash
task example:run EXAMPLE=json
```

## 可选环境变量

```bash
# 自定义 Plugin 路径
export CONFIG_PLUGIN_PATH=/path/to/config.so

# 自定义配置文件路径
export CONFIG_FILE=/path/to/config.json

# 覆盖配置（前缀 APP_）
export APP_SERVER_PORT=9090
```
