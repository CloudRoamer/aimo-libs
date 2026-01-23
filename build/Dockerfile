# aimo-libs 构建容器
# 用于提供一致的构建环境，无需本地安装 Go 工具链

FROM harbor.paralife-app.com/aimo/golang:1.25.5-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache \
    git \
    make \
    curl \
    bash \
    golangci-lint

# 安装 Task runner
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

# 设置工作目录
WORKDIR /workspace

# 设置环境变量
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GOPROXY=https://goproxy.cn,direct

# 默认命令
CMD ["task", "--list"]
