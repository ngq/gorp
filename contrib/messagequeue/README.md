# contrib/messagequeue

该目录承接 gorp 的消息队列后端实现，当前真实完成态集中在 Redis 路线，P0 重点是守住边界而不是扩写候选矩阵。

## 当前状态口径

- `已完成`：`redis`

## 说明

- `redis` 是当前唯一明确完成的真实消息队列后端。
- `framework/provider/messagequeue/noop` 属于 kernel 默认回退能力，不属于这里的未完成 contrib。
- 当前不应把 `rabbitmq`、`kafka` 等候选方向提前写成“已支持 MQ 后端”。

## P0 约束

- 文档只宣传当前真实完成的 `redis` 路线。
- 候选方向保留为后续评估项，不进入当前已完成口径。
