# monolith

默认单服务 Gin Web API starter。

说明：
- 这里描述的是项目模板主线与单服务入口。
- 这不代表相关 `contrib` 后端已经进入“已完成”状态。

## 启动

```bash
go mod tidy
go run ./cmd/app
```

## 主要目录

```text
cmd/app/main.go          # 项目入口，消费 gorp application
app/http/                # Gin 路由、handler、request/response、项目级 middleware
internal/biz/            # 业务规则与仓储接口
internal/data/           # 仓储实现与数据访问
internal/service/        # 业务编排
config/app.yaml          # 应用配置
docs/deploy.md           # 部署入口
deploy/                  # Docker / Kubernetes 交付资产
```

## 默认主线

- `cmd/app/main.go` 使用 `gorp.Run + gorp.WithMicroserviceMode()` 启动默认治理主线
- 模板默认走微服务治理模式；如果后续明确只需要轻量单体，再切回 `gorp.WithMonolithMode()`
- 默认治理能力采用“隐式提供 + 显式覆盖”：
  - 常见治理件由模式自动开启
  - `config/app.yaml` 只写需要关闭或替换的部分
  - 可通过 `governance.disable` 关闭默认治理件
  - 可通过 `governance.providers.*` 替换默认 provider backend
- `app/http/routes.go` 只注册业务路由
- `/healthz` 与 `/metrics` 由 application 默认主线注册
- Demo CRUD 只是最小业务样板，可以直接替换
- 部署资产与镜像命名说明见 `docs/deploy.md`
- Kubernetes / 镜像默认使用项目级命名：`monolith`
- 模板可以直接替换，不代表 Kubernetes / configsource / registry contrib 已完成

### 治理模式

- `gorp.WithMicroserviceMode()` — 微服务模式（模板默认）
- `gorp.WithMonolithMode()` — 单体模式，只启用基础治理件（request_identity / logging / recovery / timeout / metrics）
- `gorp.WithGinFirstMode()` — Gin 优先模式，保持原生 Gin 开发风格同时可选接入部分微服务治理

### 诊断端点

启动后可通过以下端点查看治理生效状态：

- `GET /debug/governance` — JSON 格式完整治理诊断
- `GET /debug/governance?view=brief` — 文本摘要
- `GET /debug/governance?view=providers` — provider 决策详情
- `GET /debug/governance?view=features` — feature 启用状态
- `GET /doctor/governance` — 同上（doctor 路径别名）

## 开发顺序

1. 在 `internal/biz` 定义用例与仓储接口
2. 在 `internal/data` 实现仓储
3. 在 `internal/service` 组织业务编排
4. 在 `app/http/handler` 接入 Gin handler
5. 在 `app/http/routes.go` 挂路由

复杂治理能力通过配置和 framework capability 接入，不在模板里按组件名堆目录。
