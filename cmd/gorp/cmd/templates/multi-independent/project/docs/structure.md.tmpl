# multi-independent 模板目录说明

`multi-independent` 适合已经明确要按服务独立治理、独立 `go.mod`、后续可拆仓演进的团队。

## 当前模板目录

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
  config/
  db/
  logger/
deploy/
docs/
go.work
```

## 各层职责

### 1. `services/<service>/cmd/`

职责：
- 服务启动入口
- Wire 装配声明
- 依赖图生成结果收口

默认只关心两件事：
- 这个服务如何启动
- 依赖如何在 `cmd/` 层收口

### 2. `services/<service>/internal/biz/`

职责：
- 领域规则
- usecase
- 业务核心行为

### 3. `services/<service>/internal/data/`

职责：
- 仓储实现
- 数据持久化
- 与数据库的具体读写对接

### 4. `services/<service>/internal/service/`

职责：
- 应用服务编排
- 调用 biz / data / capability 之间的协作逻辑
- 对接 application 或必要 capability 的业务使用入口

### 5. `services/<service>/internal/server/http/`

职责：
- Gin route 注册
- handler
- request / response DTO
- 项目级 HTTP middleware

这里只做 HTTP 传输层接线。

### 6. `services/<service>/internal/server/grpc/`

职责：
- 预留 Proto-first gRPC register / adapter 扩展位
- 默认模板不预置跨服务 gRPC 契约

### 7. `shared/`

职责：
- 稳定共享配置 helper
- 稳定共享数据库 helper
- 稳定共享日志门面 helper

判断标准：
- 至少被两个服务稳定复用
- 不携带服务私有语义
- 不承担默认 starter 装配主线

明确禁止继续堆旧的共享基础设施、共享 proto、共享脚本目录。

## 第三方能力放哪里

原则：
- 不默认创建 `third_party/`
- 不默认创建 `infrastructure/`
- 不按组件名堆目录，如 `redis/`、`consul/`、`dtm/`

当前模板固定边界是：
- 业务动作放 `internal/service/`
- HTTP 接线放 `internal/server/http/`
- gRPC 接线放 `internal/server/grpc/`

其余基础设施接入细节，优先通过 framework capability / application 解决。

## `docs/` 的作用

- 解释目录边界
- 解释生成步骤
- 指向部署入口

## 一句话总结

`multi-independent` 固定五件事：服务独立 `go.mod`、`shared/` 稳定边界、`cmd/` 层 Wire 收口、application-first 启动主线、`internal/server/http|grpc` 传输分层。
