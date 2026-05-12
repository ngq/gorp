# contrib/configsource

## Native Escape Hatch

- `configsource` 不要求每个 provider 都机械式暴露 native 能力。只有那些用户确实可能需要下探 vendor 能力的真实第三方 client/SDK 实现，才需要提供统一的 escape hatch。
- 像 `apollo / polaris / nacos / consul / etcd / kubernetes` 这类真实外部适配器，进入默认交付后，都应具备统一的 native escape hatch。
- 像 `local / noop / fake` 这类内核默认能力或测试桩，不需要强制提供 escape hatch 方法。
- escape hatch 的目标，不是把完整 SDK surface 映射进 application，而是保证：
  1. 默认统一接口保持轻量。
  2. 高阶用户在需要时可以拿到底层 native client/SDK。
  3. vendor-specific 能力不会泄漏进公开 contract。
- 推荐统一约定优先级：
  1. `As(target any) bool`
  2. `Underlying() any`
- 不建议每个 provider 各自发明完全不同的公开 escape 方法名。

## Official SDK First

- 从当前阶段开始，`configsource` 的默认方向是：优先使用官方 SDK / 官方 Go client 承载复杂语义，而不是长期维护自建等价 HTTP client。
- 第一梯队：`nacos`, `apollo`, `polaris`
- 第二梯队：`kubernetes`
- 阻塞说明：`apollo / polaris / kubernetes` 虽然已经切到官方 SDK / 官方 Go client，但在更完整的 publish / push / governance 语义补齐前，仍只保留“partially available”口径。

这个目录承载 gorp 的配置源后端实现，但不同子目录的完成度并不一致，必须按真实状态区分。

## Current Status

- `Done`: `consul`, `etcd`, `apollo`, `nacos`, `kubernetes`, `polaris`
- `Partially available`: none

## Explanation

- `Done` 表示已经具备真实配置源访问能力或最小可用闭环。
- `Partially available` 表示已经完成当前阶段最小闭环，具备真实 `Load / Watch` 主流程与行为测试，但还没有达到完整产品化。
- `Placeholder not done` 表示目录和 provider wiring 虽然存在，但像 `Load / Watch` 这类关键主流程仍是 TODO、空实现，或者实际上没有完成。
- `framework/provider/configsource/local` 和 `framework/provider/configsource/noop` 属于内核默认能力，不属于“placeholder not done contrib”。

## Current Stage Boundary

- `apollo`: 默认实现已经切换到官方 `agollo/v4`，并保留 fake client 注入和 native escape hatch；当前阶段已经完成 initial load、change watch、duplicate revision suppression、error classification 与 retry/stop-retry 边界，已具备真实闭环。
- `polaris`: 默认实现已经切换到官方 `polaris-go` `ConfigAPI`，并保留 fake client 注入和 native escape hatch；当前阶段已经完成 initial load、change watch、duplicate revision suppression、error classification、retry/stop-retry 边界以及 poll fallback，已具备真实闭环。
- `kubernetes`: 默认实现已经切换到 `client-go` 的 `ConfigMaps Get/Watch`，并保留 fake client 注入、not found、source error、set-not-supported、close-refuse-watch 以及 native escape hatch，已具备真实闭环。
- `nacos`: 默认实现已经切换到 `nacos-sdk-go/v2` 承载 `Load / Watch / Set` 主流程，并保留 fake client 注入、initial callback、not found、publish failure、close-refuse-watch 以及 native escape hatch，已具备真实闭环。

## Current P2 Progress

- `apollo`: 在 P2 第二层加固基础上，默认实现进一步切换到官方 `agollo/v4`，由官方 SDK 承载默认加载与变更回调；同时保留 fake client 注入、initial load、watch retry、duplicate revision suppression、`AuthFailed/ConfigNotFound/SourceUnavailable` 错误分类、poll fallback、native `As/Underlying` 以及行为测试。
- `polaris`: 在 P2 第二层加固基础上，默认实现进一步切换到官方 `polaris-go` `ConfigAPI`，由官方 SDK 承载默认加载与配置变更 watch；同时保留 fake client 注入、initial load、watch retry、duplicate revision suppression、`AuthFailed/ConfigNotFound/SourceUnavailable` 错误分类、poll fallback、native `As/Underlying` 以及行为测试。
- `kubernetes`: 已完成第二梯队默认实现切换，使用 `client-go` 承载 ConfigMap 读取与 watch；同时保留 fake client 注入、close-refuse-watch、native `As/Underlying` 以及行为测试。

## Current Production Prerequisites

- `nacos` 默认实现不再使用自建 HTTP client，而是复用官方 `nacos-sdk-go/v2`；公开 contract 保持不变；需要 vendor 能力的用户可通过 `As/Underlying` 下探到底层 native client。
- `apollo` 默认实现不再使用自建 HTTP pull，而是复用官方 `agollo/v4`；公开 contract 保持不变；需要 Apollo native 能力的用户可通过 `As/Underlying` 下探到底层 client。
- `apollo` 当前 Watch 已明确区分“可重试错误”和“必须停止错误”：`SourceUnavailable` 会按 backoff 重试，`AuthFailed/ConfigNotFound` 不再盲目重试。
- `consul / etcd / kubernetes` 当前也都提供统一 `As/Underlying` escape hatch；用户可以继续通过 gorp contract 走默认主路径，也可以在需要时拿到底层官方 Go client。
- `apollo / polaris` 当前都明确不支持 `Set` 的产品化回写；如果需要写路径，必须单独补齐完整 publish 语义与失败策略。
- `apollo / polaris` 当前都具备 fake client 行为测试，且默认实现已切换到官方 SDK；但仍未完成完整 publish 语义、server-side long-polling 细节暴露、回写能力以及更深治理语义。
- `apollo / polaris` 当前 Watch 已达到“即使没有显式先 `Load`，仍会先拉一次 initial snapshot”，并且在 source 不可达时按 backoff interval 重试，同 revision 不会重复分发。
- `polaris` 另外具备基于 `PollInterval` 的 fallback refresh：即使 SDK watch 不主动推送，也能周期性拉取配置，并基于 revision 做增量分发。

## P0 Constraints

- 不要因为目录存在，就对外宣称“配置中心能力已完成”。
- 不要把 `configsource/nacos` 和 `registry/nacos` 混淆成“Nacos 整体能力已完成”。
- 不要把 Kubernetes 部署模板误写成 `configsource/kubernetes` 已完成。
