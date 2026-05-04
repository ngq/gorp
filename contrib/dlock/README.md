# contrib/dlock

该目录承接 gorp 的分布式锁后端实现，当前主线状态清晰，P0 重点是守住已完成边界。

## 当前状态口径

- `已完成`：`redis`

## 说明

- `redis` 是当前唯一明确完成的真实分布式锁后端。
- `framework/provider/dlock/noop` 属于 kernel 默认回退能力，不属于这里的未完成 contrib。

## P0 约束

- 不把 `noop` 与真实锁后端混写。
- 不把其他候选锁后端提前写成已支持能力。
