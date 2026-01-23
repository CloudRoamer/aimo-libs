# aimo-libs

AIMO 项目通用库集合，提供可复用的基础组件和工具。

## 项目结构

```
aimo-libs/
├── config/              # 统一配置管理 SDK
├── build/               # 构建容器定义
├── Taskfile.yml         # Task 构建任务定义
└── README.md
```

## 模块列表

### config - 统一配置管理 SDK

通用的统一配置管理 SDK，支持多配置源、优先级合并和热更新。

详细文档请查看：[config/README.md](config/README.md)

## 构建说明

本项目使用 [Task](https://taskfile.dev/) 作为任务运行器，提供统一的构建接口。

### 方式一：本地构建（推荐）

**前置要求**：
- Go 1.25.5 或更高版本
- Task（任务运行器）
- golangci-lint（代码检查工具）

**安装 Task**：

```bash
# macOS
brew install go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

# Windows
choco install go-task

# 或使用 Go 安装
go install github.com/go-task/task/v3/cmd/task@latest
```

**安装 golangci-lint**：

```bash
# macOS
brew install golangci-lint

# Linux/macOS (通用)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Windows
choco install golangci-lint
```

**可用任务**：

```bash
# 查看所有可用任务
task

# 构建所有模块
task build

# 运行测试
task test

# 运行测试（跳过外部依赖）
task test:short

# 生成测试覆盖率报告
task test:coverage

# 代码检查
task lint

# 代码格式化
task fmt

# 检查代码格式
task fmt:check

# 清理构建产物
task clean

# 整理依赖
task mod:tidy

# CI 流程（格式检查 + 代码检查 + 测试）
task ci

# 完整 CI 流程（包含覆盖率）
task ci:full
```

### 方式二：使用构建容器（无需本地环境）

如果不想在本地安装 Go 工具链，可以使用构建容器。

**构建镜像**：

```bash
docker build -f build/Dockerfile -t aimo-libs-builder .
```

**运行任务**：

```bash
# 查看可用任务
docker run --rm -v $(pwd):/workspace aimo-libs-builder task

# 运行测试
docker run --rm -v $(pwd):/workspace aimo-libs-builder task test

# 代码检查
docker run --rm -v $(pwd):/workspace aimo-libs-builder task lint

# CI 流程
docker run --rm -v $(pwd):/workspace aimo-libs-builder task ci
```

**使用 docker compose**（推荐）：

创建 `docker-compose.yml` 文件（可选）：

```yaml
version: '3'

services:
  builder:
    build:
      context: .
      dockerfile: build/Dockerfile
    volumes:
      - .:/workspace
    working_dir: /workspace
```

运行：

```bash
# 构建
docker compose run --rm builder task build

# 测试
docker compose run --rm builder task test

# CI 流程
docker compose run --rm builder task ci
```

## 开发工作流

### 1. 克隆项目

```bash
git clone https://github.com/CloudRoamer/aimo-libs.git
cd aimo-libs
```

### 2. 安装依赖

```bash
task mod:download
```

### 3. 开发

```bash
# 格式化代码
task fmt

# 运行测试
task test

# 代码检查
task lint
```

### 4. 提交前检查

```bash
# 运行 CI 流程
task ci

# 或运行完整检查（包含覆盖率）
task ci:full
```

## 许可证

商业授权 - 芜湖图忆科技有限公司
