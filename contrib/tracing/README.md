# contrib/tracing

该目录承接 gorp 的链路追踪后端实现，当前真实完成态已经明确，P0 重点是持续守住 middleware 与后端的边界口径。

## 当前状态口径

- `已完成`：`otel`

## 说明

- `otel` 是当前唯一明确完成的真实 tracing 后端。
- `otel` 当前已补齐统一 `As / Underlying` 下探，用户可以继续走 gorp tracing 契约，也可以按需拿到原生 `*sdktrace.TracerProvider`、`trace.Tracer` 与 `trace.Span`。
- `framework/provider/tracing/noop` 属于 kernel 默认回退能力。
- `framework/provider/tracing/middleware` 属于 kernel 中间件层，不属于这里的未完成 contrib。

## P0 约束

- 文档中必须明确区分 tracing middleware 与 tracing backend。
- 不把未来可能扩展的 tracing 后端提前写成当前已支持矩阵。
