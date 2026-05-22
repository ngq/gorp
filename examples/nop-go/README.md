# nop-go

多服务进阶模板（独立治理版）

> 适用场景：
> - 已经明确需要每个服务独立 `go.mod`
> - 需要保留 `shared/` 作为稳定共享边界
> - 希望默认主线保持 application-first，而不是让业务先理解 framework 内部装配

## 先看这三件事

- 创建项目：`gorp new --template multi-independent`
- 生成 Wire：`make generate`
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
    go.mod
  order/
  product/
shared/
  README.md
  config/
  db/
  logger/
deploy/
docs/
go.work
Makefile
```

目录说明看：`docs/structure.md`

## 默认开发主线

- 每个服务统一从 `services/<service>/cmd/main.go` 启动
- Wire 只保留在 `cmd/` 层收口依赖
- 默认优先通过 `gorp.Run(...) + gorp.WithMicroserviceMode()` 启动微服务治理主线
- HTTP 接线放 `internal/server/http/`
- gRPC 扩展位放 `internal/server/grpc/`
- `shared/` 只放稳定共享 helper，不再承担服务默认装配

## 快速开始

### 1. 初始化 workspace

```bash
go work sync
```

### 2. 生成 Wire

```bash
make generate
```

说明：
- `wire.go` 只留在各服务 `cmd/` 层
- 默认模板直接以各服务 `cmd/main.go` 作为启动主线

### 3. 启动服务

```bash
make run-user
make run-order
make run-product
```

说明：
- 各服务统一走 `services/<service>/cmd/main.go`
- HTTP 传输层统一放 `internal/server/http/`

### 4. 运行测试

```bash
make test
```

说明：
- 默认附带最小 HTTP smoke 测试
- 业务逻辑继续放在 `internal/biz/` 与 `internal/service/`

## 继续往下时再看

### 5. 本地联调

```bash
make deploy-local
```

说明：
- 使用 `deploy/compose/docker-compose.yaml`
- 默认通过 workspace 直接运行三个服务
- 容器名使用 `nop-go-*`，避免多个生成项目在同一台机器上撞名

### 6. Harbor 推镜像

```bash
make harbor-push HARBOR_REGISTRY=harbor.example.com HARBOR_NAMESPACE=nop-go IMAGE_TAG=v1.0.0
```

说明：
- 本地与 Harbor 镜像名使用项目级前缀：`nop-go-user-service`、`nop-go-order-service`、`nop-go-product-service`
- Kubernetes overlays 也使用同一套镜像名映射，避免构建、推送、部署三处手工改名

### 7. 部署资产

看：`docs/deploy.md`

## 治理能力

默认走微服务治理模式（`gorp.WithMicroserviceMode()`），自动开启以下治理件：

- HTTP 侧：request_identity / logging / recovery / timeout / metrics
- 微服务增强：metadata / tracing / selector / serviceauth / circuitbreaker

### 关闭或替换治理件

在各服务 `config/app.yaml` 中通过 `governance` 段覆盖：

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

- 业务日志优先走 `gorp/log`
- `shared/logger/` 只保留稳定共享门面，不承担 starter 教学职责
