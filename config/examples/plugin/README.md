# Config Plugin 使用示例

本示例演示如何通过 Go Plugin 方式动态加载和使用 aimo-libs/config SDK。

## 什么是 Plugin 模式

Plugin 模式允许将 config SDK 编译为共享库（.so 文件），在运行时动态加载。这种方式的优势：

1. **解耦依赖**：应用程序无需在编译时链接 SDK
2. **灵活升级**：可以独立更新 SDK 版本而无需重新编译应用
3. **减少二进制大小**：多个应用可以共享同一个 SDK 库

## 构建 Plugin

### macOS

```bash
# 在项目根目录执行
task build:plugin:darwin
```

生成的 Plugin 文件位于：`dist/darwin/config.so`

### Linux

```bash
# 在项目根目录执行
task build:plugin:linux
```

生成的 Plugin 文件位于：`dist/linux/config.so`

**注意**：Plugin 必须在目标平台上编译，不支持交叉编译。

## 运行示例

### 准备环境

```bash
# 1. 进入示例目录
cd config/examples/plugin

# 2. 确保已构建 Plugin
ls ../../dist/config.so  # 应该存在

# 3. （可选）设置环境变量
export APP_DEBUG=true
export APP_SERVER_PORT=9090
```

### 运行程序

```bash
go run main.go
```

### 自定义 Plugin 路径

```bash
export CONFIG_PLUGIN_PATH=/path/to/config.so
go run main.go
```

### 自定义配置文件

```bash
export CONFIG_FILE=/path/to/config.yaml
go run main.go
```

## 示例说明

示例程序展示了以下功能：

### 1. 加载 Plugin

```go
p, err := plugin.Open("../../dist/config.so")
if err != nil {
    log.Fatalf("Failed to open plugin: %v", err)
}
```

### 2. 查找导出的符号

```go
newManagerSym, err := p.Lookup("NewManager")
if err != nil {
    log.Fatalf("Failed to lookup NewManager: %v", err)
}
```

### 3. 创建配置管理器

```go
newManager := newManagerSym.(func(...interface{}) *config.Manager)
manager := newManager()
```

### 4. 添加配置源

支持多种配置源：

- **文件配置源**（YAML/JSON）
- **环境变量配置源**
- **Consul 配置源**
- **PostgreSQL 配置源**

### 5. 读取配置值

```go
cfg := manager.Config()
appName := cfg.GetString("app.name", "unknown")
port := cfg.GetInt("server.port", 8080)
debug := cfg.GetBool("debug", false)
```

### 6. 监听配置变更

```go
manager.OnChange(func(event config.Event, oldConfig, newConfig config.Config) {
    fmt.Printf("Config changed: type=%s, source=%s\n", event.Type, event.Source)
})
manager.Watch()
```

## 配置优先级

当多个配置源提供相同的配置键时，优先级规则如下：

1. **环境变量**（优先级：100）- 最高
2. **Consul**（优先级：80）
3. **PostgreSQL**（优先级：70）
4. **文件**（优先级：60）- 最低

高优先级的配置会覆盖低优先级的配置。

## 环境变量映射规则

环境变量配置源默认映射规则：

| 环境变量 | 配置键 |
|---------|--------|
| `APP_DEBUG` | `debug` |
| `APP_SERVER_PORT` | `server.port` |
| `APP_DATABASE_HOST` | `database.host` |

映射规则：
- 移除前缀（如 `APP_`）
- 转换为小写
- 将下划线 `_` 替换为点 `.`

## 测试配置热更新

### 文件配置源

1. 运行示例程序
2. 修改 `config.yaml` 文件
3. 观察控制台输出的变更事件

```bash
# 修改配置
echo "debug: true" >> config.yaml
```

### 环境变量配置源

环境变量配置源不支持热更新（需要重启程序）。

## 平台兼容性

| 平台 | 支持状态 |
|------|----------|
| Linux | ✅ 支持 |
| macOS | ✅ 支持 |
| Windows | ❌ 不支持（Go Plugin 限制） |

## 常见问题

### Q: 为什么 macOS Plugin 必须在 macOS 上构建？

A: Go Plugin 使用 CGO 和动态链接，依赖目标平台的 C 库和符号。不同平台的二进制格式不兼容。

### Q: Plugin 加载失败怎么办？

A: 检查以下几点：
1. Plugin 是否在当前平台上编译
2. Go 版本是否一致（编译和运行时）
3. Plugin 路径是否正确
4. 符号名称是否匹配

### Q: 可以在 Windows 上使用吗？

A: 不行。Go Plugin 仅支持 Linux 和 macOS。Windows 用户请直接导入 SDK 包。

### Q: 多个应用可以共享同一个 Plugin 吗？

A: 可以。这是 Plugin 模式的主要优势之一。只需将 `config.so` 放在共享位置，多个应用通过 `CONFIG_PLUGIN_PATH` 环境变量指向它。

## 进阶用法

### 自定义合并器

```go
// 从 Plugin 获取 WithMerger 选项
withMergerSym, _ := p.Lookup("WithMerger")
withMerger := withMergerSym.(func(interface{}) interface{})

// 实现自定义 Merger（需要实现 config.Merger 接口）
customMerger := &MyCustomMerger{}
manager := newManager(withMerger(customMerger))
```

### 自定义 key 映射

```go
// 从 Plugin 获取 EnvWithKeyMapping 选项
envWithKeyMappingSym, _ := p.Lookup("EnvWithKeyMapping")
envWithKeyMapping := envWithKeyMappingSym.(func(func(string) string) interface{})

// 自定义映射函数
customMapping := func(envKey string) string {
    // 自定义逻辑
    return strings.ToLower(envKey)
}

envSource := newEnvSource(
    envWithPrefix("APP_"),
    envWithKeyMapping(customMapping),
)
```

## 参考文档

- [Config SDK 设计文档](../../docs/DESIGN-ConfigSDK核心架构.md)
- [Go Plugin 官方文档](https://pkg.go.dev/plugin)
- [示例代码](./main.go)
