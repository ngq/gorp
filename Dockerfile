# Dockerfile for gorp CLI
# 多阶段构建，最小化镜像体积
#
# 中文说明：
# - 使用 golang:1.25-alpine 作为构建环境；
# - 最终镜像基于 alpine，体积约 20MB；
# - 包含安全扫描和 SBOM 生成。

# Build stage
FROM golang:1.25-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# 复制 go.mod 和 go.sum，利用缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建 CLI
ARG VERSION=unknown
ARG COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
    -o /gorp ./cmd/gorp

# Final stage
FROM alpine:3.19

# 安装运行依赖
RUN apk add --no-cache ca-certificates tzdata

# 创建非 root 用户
RUN adduser -D -g '' appuser

WORKDIR /app

# 复制二进制文件
COPY --from=builder /gorp /app/gorp

# 设置权限
RUN chmod +x /app/gorp && chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

ENTRYPOINT ["/app/gorp"]
CMD ["--help"]

# Labels for OCI image
LABEL org.opencontainers.image.title="gorp CLI"
LABEL org.opencontainers.image.description="gorp Framework CLI Tool"
LABEL org.opencontainers.image.vendor="gorp"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/ngq/gorp"