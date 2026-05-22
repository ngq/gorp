# 组件归位参考

这里只回答一件事：
- 新能力先放哪一层

## 归位判断

先按这个顺序理解：

1. 业务规则放 `services/<service>/internal/biz/`
2. 数据访问放 `services/<service>/internal/data/`
3. 应用编排放 `services/<service>/internal/service/`
4. HTTP / gRPC 传输注册放 `services/<service>/internal/server/http/` 与 `services/<service>/internal/server/grpc/`
5. 启动与装配放 `services/<service>/cmd/`
6. 稳定共享 helper 放 `shared/config/`、`shared/db/`、`shared/logger/`

## 常见组件怎么归位

| 场景 | 默认归位 |
| --- | --- |
| 缓存 / Redis 使用动作 | `internal/service/` |
| 消息发布 / 消费编排 | `internal/service/` |
| 外部 HTTP / 平台调用编排 | `internal/service/` |
| 仓储实现 / 数据库读写 | `internal/data/` |
| 领域规则 / 用例 | `internal/biz/` |
| HTTP 路由与 handler | `internal/server/http/` |
| gRPC register / adapter | `internal/server/grpc/` |
| 共享配置 helper / 日志门面 / DB helper | `shared/` |

## 什么时候先不要新建目录

下面这些情况，先不要急着新建 `third_party/`、`infrastructure/`、`platform/`：

- 只是刚接一个中间件
- 只有 1~2 个服务会用到
- 当前主要问题还是“放哪”，而不是“现有结构已经失控”

## 一个最小判断法

先问自己一句：

> 这是业务什么时候调用，还是底层能力怎么接入？

- 业务动作：先放 `internal/service/`
- 底层接入：优先继续走 framework 能力入口

## 当前建议

- 先按现有结构开发
- 需要时回看 `docs/structure.md`
- 部署动作回看 `docs/deploy.md`
