# JSON 示例

本示例演示通过 WASM 加载 Config SDK，并使用 JSON 配置文件。

## 构建 WASM

```bash
task build:wasm
```

## 运行示例

```bash
task example:run EXAMPLE=json
```

## 可选环境变量

```bash
# 自定义 WASM 路径
export CONFIG_WASM_PATH=/path/to/config.wasm

# 自定义配置文件路径
export CONFIG_FILE=/path/to/config.json

# 覆盖配置（前缀 APP_）
export APP_SERVER_PORT=9090
```
