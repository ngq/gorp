# multi-flat-wire 模板目录说明

`multi-flat-wire` 适合已经明确要走多服务主线，并希望把 Wire 收口在 `cmd/` 层的团队。

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
  order/
  product/
proto/
deploy/
pkg/
docs/
```

## 各层职责

### 1. `services/<service>/cmd/`

职责：

- 服务启动入口
- Wire 装配声明
- 依赖图生成结果收口

这里默认只关心两件事：

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
- 与数据库或存储层的具体读写对接

### 4. `services/<service>/internal/service/`

职责：

- 应用服务编排
- 调用 biz / data / capability 之间的协作逻辑
- 对接 application 或必要 capability 的业务使用入口

这里可以出现“发消息”“调 RPC”“查缓存”的业务动作编排，但不应直接放底层 SDK 初始化细节。

### 5. `services/<service>/internal/server/http/`

职责：

- Gin route 注册
- handler
- request / response DTO
- 项目级 middleware

这里只做 HTTP 传输层接线，不承载业务核心逻辑。

### 6. `services/<service>/internal/server/grpc/`

职责：

- Proto-first gRPC register / adapter
- 对 framework 托管的 `grpc.Server` 做最小接线

这里只做 gRPC 传输层接线，不承载业务核心逻辑。

### 7. `proto/`

职责：

- Proto-first 契约
- gRPC 对外接口定义
- 跨服务通信协议边界

## 第三方能力放哪里

原则：

- 不默认创建 `third_party/`
- 不默认创建 `infrastructure/`
- 不按组件名堆目录，如 `redis/`、`consul/`、`dtm/`

当前模板固定的边界是：

- 业务动作放 `internal/service/`
- HTTP 接线放 `internal/server/http/`
- gRPC 接线放 `internal/server/grpc/`

其余基础设施接入细节，优先通过 framework capability / application 解决。

## `docs/` 目录的作用

- 解释目录边界
- 解释生成步骤
- 指向部署入口

## 一句话总结

`multi-flat-wire` 固定四件事：服务目录、`cmd/` 层 Wire 收口、application-first 启动主线、`internal/server/http|grpc` 传输分层。
