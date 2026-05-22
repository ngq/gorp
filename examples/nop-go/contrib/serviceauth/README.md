# contrib/serviceauth

该目录当前不是“迁移中目录”，而是已经落下真实后端实现的能力域。

## 当前状态口径

- `已完成`：`token`、`mtls`

## 说明

- `token` 与 `mtls` 已具备真实服务认证接入，不应再与占位目录混写。
- `framework/provider/serviceauth/noop` 属于 kernel 默认回退能力，不属于这里的未完成 contrib。

## P0 约束

- 不再保留“Phase 外移说明”口径。
- 文档中必须明确区分已完成的 serviceauth 后端与其他能力域中的占位目录。
