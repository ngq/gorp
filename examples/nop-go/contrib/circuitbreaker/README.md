# contrib/circuitbreaker

该目录承接 gorp 的熔断与限流后端实现，当前必须按真实状态口径描述，而不是按迁移阶段描述。

## 当前状态口径

- `部分可用`：`sentinel`

## 说明

- `sentinel` 已具备 P1 最小治理闭环，当前可以覆盖规则加载、`Entry / block`、状态记录与等待语义主路径。
- `sentinel` 当前已补齐统一 `As / Underlying` 下探，可按需拿到官方 Sentinel 全局 `SlotChain`，继续下沉到原生规则与治理能力。
- 规则加载来源、完整治理规则边界与更深产品化仍未完成。
- 因此这里当前只能按 `部分可用` 宣传，不能按“完整熔断 / 限流治理后端”宣传。

## 当前 P1 收尾边界

- 当前 `P1` 已补 `State / RecordSuccess / RecordFailure / WaitTimeout / Reservation` 与关键行为测试。
- 当前 `P1` 已补默认值回退、空规则返回与 runtime init 失败路径。
- 本轮只收“最小治理闭环”，不直接承诺完整治理平台或完整规则中心能力。

## P0 约束

- 不再保留任何“Phase 外移中”这类迁移期说法。
- P0 只收敛状态与口径，不扩展到 P1 的真实治理能力做实。
