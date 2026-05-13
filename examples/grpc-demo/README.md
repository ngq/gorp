# grpc-demo

多服务起步模板（Wire 装配版）

> 起步：
> - 已明确要走多服务主线时，直接使用 `gorp new multi-wire`
> - 只想先做单服务业务时，优先使用 `gorp new`

## 先看这三件事

- 创建项目：`gorp new multi-wire`
- 生成代码：`make generate`
- 启动服务：`make run-user` / `make run-order` / `make run-product`

## 目录速览

```text
services/
  user/
    cmd/
      main.go
      wire.go
      wire_gen.go
    internal/
      biz/
      data/
      service/
      server/
        http/
          routes.go
          handler/
          request/
          response/
          middleware/
        grpc/
    config/
  order/
  product/
proto/
deploy/
docs/
Makefile
```

目录：

- `docs/structure.md`

## 默认开发主线

- 每个服务默认从 `services/<service>/cmd/main.go` 启动
- Wire 只留在 `cmd/` 层收口依赖装配
- HTTP 接线在 `internal/server/http`，gRPC 接线在 `internal/server/grpc`
- 业务编排放在 `internal/service`
- gRPC 默认主线是 Proto-first：服务端围绕 `GRPCServerRegistrar + pb.RegisterXxxServer(...)`，客户端围绕 `GRPCConnFactory + pb.NewXxxClient(conn)`

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 统一生成代码

```bash
make generate
```

生成：

- 统一执行 Wire 与 Proto 生成
- `wire.go` 只收口 `cmd/` 层装配
- 服务入口由 `services/<service>/cmd/main.go` 承接

### 3. 启动服务

```bash
make run-user
make run-order
make run-product
```

说明：

- 各服务统一由 `services/<service>/cmd/main.go` 启动
- 默认优先通过 `gorp.Run(...) + gorp.WithMicroserviceMode()` 启动微服务治理主线
- HTTP 接线在 `internal/server/http`，gRPC 接线在 `internal/server/grpc`

### 4. 运行测试

```bash
make test
```

说明：

- 模板默认带最小测试基线
- `internal/biz` 下提供 usecase 级样板测试
- `internal/server/http` 下提供 HTTP smoke 测试
- `user` / `order` 额外保留最小 gRPC / metadata / dlock 样板测试

## 继续往下时再看

### 5. 本地联调

```bash
make deploy-local
```

说明：

- 通过 `deploy/compose/docker-compose.yaml` 拉起 Redis 与三组服务样板
- compose 会把项目根目录挂载到 `/workspace` 并执行 `go run ./services/<service>/cmd`
- 容器名使用 `grpc-demo-*`，避免多个生成项目在同一台机器上撞名
- 先构建镜像时再执行 `make deploy-local-build`
- 完成后执行 `make deploy-local-down`

### 6. Harbor 推送镜像

```bash
make harbor-push HARBOR_REGISTRY=harbor.example.com HARBOR_NAMESPACE=grpc-demo IMAGE_TAG=v1.0.0
```

说明：

- 本地与 Harbor 镜像名使用项目级前缀：`grpc-demo-user-service`、`grpc-demo-order-service`、`grpc-demo-product-service`
- Kubernetes overlays 也使用同一套镜像名映射，避免构建、推送、部署三处手工改名

### 7. 部署资产

看：`docs/deploy.md`

### 8. 验证 Proto-first gRPC 双服务链路

先确认一条服务间调用链路能跑通，再继续扩展更多治理能力。

- `order-service` 作为 gRPC client / caller
- `user-service` 作为 gRPC server / callee

关键配置：

- `service.name`
- `rpc.mode=grpc`
- `rpc.grpc.address` / `rpc.grpc.target`
- `metadata.enabled + metadata.propagate_prefix`
- `tracing.enabled + tracing.backend`
- `service_auth.mode=token + service_auth.token.secret`

主线入口：

- `GRPCServerRegistrar + pb.RegisterXxxServer(...)`
- `GRPCConnFactory + pb.NewXxxClient(conn)`

### 9. 生成 proto 代码

修改 `proto/*.proto` 后，直接使用 `protoc` + Go 插件生成：

```bash
PATH="$(go env GOPATH)/bin:$PATH" protoc -I . \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/user/v1/user.proto
```

说明：

- 默认优先按服务间 Proto-first 接线验证
- 这一步按需执行

## 治理能力

默认走微服务治理模式（`gorp.WithMicroserviceMode()`），自动开启以下治理件：

- HTTP 侧：request_identity / logging / recovery / timeout / metrics
- 微服务增强：metadata / tracing / selector / serviceauth / circuitbreaker

### 关闭或替换治理件

在 `config/app.yaml` 中通过 `governance` 段覆盖：

```yaml
governance:
  disable:
    - tracing        # 显式关闭默认治理件
  providers:
    serviceauth: mtls  # 替换默认 provider backend
```

也可通过代码选项覆盖：

- `gorp.WithGovernanceDisabled("tracing", "selector")` — 代码侧显式关闭
- `gorp.WithGovernanceProvider("serviceauth", "mtls")` — 代码侧替换 backend

优先级：代码显式覆盖 > 配置显式覆盖 > 模式默认值 > provider 兜底

### 治理模式

- `gorp.WithMicroserviceMode()` — 微服务模式（默认）
- `gorp.WithMonolithMode()` — 单体模式，只启用基础治理件
- `gorp.WithGinFirstMode()` — Gin 优先模式，保持原生 Gin 开发风格同时可选接入部分微服务治理

### 诊断端点

启动后可通过以下端点查看治理生效状态：

- `GET /debug/governance` — JSON 格式完整治理诊断
- `GET /debug/governance?view=brief` — 文本摘要
- `GET /debug/governance?view=providers` — provider 决策详情
- `GET /debug/governance?view=features` — feature 启用状态
- `GET /doctor/governance` — 同上（doctor 路径别名）

## 日志建议

- 优先使用 `gorp/log` 作为业务日志入口
- 请求级字段按实际业务需要再补
