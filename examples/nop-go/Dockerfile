# nop-go 微服务通用 Dockerfile
# 多阶段构建，支持所有微服务
# 使用阿里云镜像加速

# 阶段1: 构建
FROM docker.m.daocloud.io/library/golang:1.26-alpine AS builder

# 设置 Go 代理（加速依赖下载）
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build/examples/nop-go

# 复制 go.work 文件
COPY examples/nop-go/go.work examples/nop-go/go.work.sum ./

# 复制共享模块
COPY examples/nop-go/shared/go.mod ./shared/

# 复制服务 go.mod
COPY examples/nop-go/services/*/go.mod ./services/

# 复制所有源代码（从仓库根作为构建上下文）
COPY . /build

# 构建参数：服务名称
ARG SERVICE_NAME

# 下载依赖
RUN cd services/${SERVICE_NAME} && go mod download || true

# 构建服务
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./services/${SERVICE_NAME}/cmd

# 确保 docs 目录始终存在，避免运行阶段 COPY 失败
RUN mkdir -p /build/examples/nop-go/services/${SERVICE_NAME}/docs

# 阶段2: 运行
FROM docker.m.daocloud.io/library/alpine:3.19

# 需要在运行阶段重新声明构建参数，供 COPY 使用
ARG SERVICE_NAME

# 设置镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata wget

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN adduser -D -g '' appuser

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/server .
COPY --from=builder /build/examples/nop-go/services/${SERVICE_NAME}/config ./config

# 复制 swagger 文档（如果存在）
COPY --from=builder /build/examples/nop-go/services/${SERVICE_NAME}/docs ./docs

# 设置权限
RUN chown -R appuser:appuser /app

USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8000}/healthz || exit 1

# 暴露端口
EXPOSE ${PORT:-8000}

# 启动服务
CMD ["./server"]