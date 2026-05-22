# contrib/registry

## Native Escape Hatch

- `registry` 不要求所有 provider 机械式统一暴露 native 能力，但所有真实第三方注册发现适配对象都应预留统一下探约定。
- `eureka / polaris / servicecomb / zookeeper / nacos / consul / etcd / kubernetes` 这类真实外部适配对象，进入默认交付后都应补统一 native escape hatch。
- `noop / fake / in-memory` 这类默认回退或测试替身，不要求补下探。
- 下探能力的目标不是把注册中心 SDK 全量平铺进公共接口，而是保证：
  1. 默认 `registry` 抽象保持轻便。
  2. 用户在需要实例治理、厂商规则、专有会话语义时，可以拿到底层 native client / SDK。
  3. 厂商专有能力不直接泄漏进公共 contract。
- 推荐统一约定优先级：
  1. `As(target any) bool`
  2. `Underlying() any`
- 不建议每个 provider 各自扩散不同的命名。

## Official SDK First

- 从当前阶段开始，`registry` 的默认方向也是：优先用官方 SDK / 官方 Go client 承接注册、发现、watch、心跳、会话与治理语义。
- 第一梯队：`polaris`
- 第二梯队：`kubernetes`
- 阻塞项：`eureka`、`servicecomb` 当前不承诺“完全基于官方 SDK”，仍按生产可用适配器路线继续收口。

该目录承接 gorp 的服务注册与发现后端实现，但不同子目录的完成度并不一致，当前必须按真实状态区分口径。

## 当前状态口径

- `已完成`：`consul`、`etcd`、`nacos`、`kubernetes`、`polaris`
- `部分可用`：`eureka`、`servicecomb`、`zookeeper`

## 说明

- `已完成` 表示已具备真实注册发现接入或最小可用闭环。
- `部分可用` 表示已完成当前阶段最小闭环，具备真实注册/发现主流程与行为测试，但仍未到完整产品化。
- `占位未完成` 表示目录与 provider 接线已存在，但关键主流程仍是 TODO、空实现、空结果或仅本地状态。
- `framework/provider/discovery/noop` 属于 kernel 默认回退能力，不属于这里的“占位未完成 contrib”。

## 当前阶段收口边界

- `kubernetes`：默认实现已切到 `client-go` `Endpoints Get/Watch`，并明确不再自造 `Register / Deregister` 语义；当前保留初始快照、缓存命中、not found、source error、关闭退出、关闭后拒绝 watch 与 native 下探能力，已具备真实发现闭环。
- `eureka / servicecomb / zookeeper`：已完成 P2 阶段最小闭环，其中这三项都已进入第二层继续补强；`zookeeper` 已补齐 `As / Underlying` 下探，但三者仍维持“部分可用”口径。
- `polaris`：默认实现已进一步切到官方 `polaris-go` `ProviderAPI + ConsumerAPI`，由官方 SDK 承接默认注册、发现与 watch；同时保留 fake client 注入、重复注册保护、关闭后拒绝 watch、watch 失败退避重试、稳定实例排序、重复快照抑制、native `As / Underlying` 与行为测试，已具备真实注册发现闭环。

## 当前 P2 进度

- `eureka`：已完成 P2 第二层补强，在第一版最小注册/发现闭环基础上，补齐 `Register / Deregister / Discover / Watch`、最小心跳续租框架、失败退避、实例丢失后重注册、重复注册保护、注销停止续租、watch 首帧快照、重复快照抑制、watch 临时错误重试、关闭后旧 watcher 退出与 fake client 行为测试。
- `zookeeper`：已完成 P2 第二层继续补强，在第一版最小注册/发现闭环基础上，补齐 watcher 初始快照、节点消失通知、关闭后拒绝 watch、连接/会话类错误退避重连、稳定实例排序、重复快照抑制与 fake backend 行为测试。
- `zookeeper`：当前已补齐统一 `As / Underlying` 下探，可直接拿到底层 `*zk.Conn`。
- `polaris`：已在 P2 第二层继续补强基础上，继续把默认实现切到官方 `polaris-go` `ProviderAPI + ConsumerAPI`，由官方 SDK 承接默认注册、发现与 watch；同时保留 fake client 注入、重复注册保护、关闭后拒绝 watch、watch 失败退避重试、稳定实例排序、重复快照抑制、native `As / Underlying` 与行为测试。
- `kubernetes`：已在第二梯队完成默认实现切换，由 `client-go` 承接 Endpoints 查询与 watch；同时保留 fake client 注入、`Register / Deregister` not-supported、native `As / Underlying` 与行为测试。
- `servicecomb`：已完成 P2 第二层继续补强，在第一版最小注册/发现闭环基础上，补齐最小心跳续租框架、失败退避、实例丢失后重注册、重复注册保护、注销停止续租、最小 watch 快照监听、重复快照抑制、服务消失空快照与 fake client 行为测试。
- `servicecomb`：当前已补齐统一 `As / Underlying` 下探，并补上最小 `Watch` 快照监听闭环；但仍未进入官方 SDK 默认实现路线。
- 当前 `eureka / zookeeper / polaris / servicecomb` 已从“占位未完成”提升为“部分可用”。

## 当前生产前置

- `eureka` 当前适合已有 Spring Cloud / Eureka 生态、且先只需要最小注册/发现闭环的场景；当前已补齐最小心跳续租、失败退避、实例丢失后重注册、watch 首帧快照、重复快照抑制与 watch 重拉，但更完整心跳、续租与实例治理仍未产品化。
- `zookeeper` 当前适合已有 Zookeeper 集群、且接受临时节点注册语义的场景；当前已补齐首帧快照、节点删除通知、重复快照抑制、关闭退出与连接/会话类错误重试，但更完整 session 生命周期与官方重连治理仍未补齐。
- `polaris` 当前默认实现已不再停留在 `in-memory` 主路径，而是复用官方 `polaris-go`；首次 watch 可拿到快照，注册/注销后可收到更新，watch 失败后会按退避间隔重试，并对实例顺序抖动做稳定排序和重复快照抑制。完整路由规则、心跳治理与更深产品化语义仍待继续补齐。
- `consul / etcd / nacos / kubernetes / zookeeper` 当前都已补齐统一 `As / Underlying` 下探，用户可以默认走 gorp `registry` 契约，也可以按需取到底层官方 client / SDK。
- `servicecomb` 当前已补最小心跳续租、实例治理骨架与轮询 watch 闭环：注册后可按间隔续租，注销与关闭会停止续租，心跳失败会退避，实例丢失时会尝试重注册，watch 遇到临时错误会继续轮询，并对重复快照做抑制、在服务消失时发送空快照；但更完整治理语义与 SDK 能力仍未补齐。
- `eureka / servicecomb` 当前虽然都已具备统一 `As / Underlying` 下探，但这只解决了“高级用户可下沉”的问题，并不代表已经切到官方 SDK 默认实现路线。
- `eureka` 当前已补最小心跳续租、失败退避与 watcher 语义：注册后可按心跳间隔续租，注销与关闭会停止续租，心跳 not found 时会尝试重注册，首次 watch 可拿到快照，重复快照会被抑制，watch 临时错误后会按间隔重拉；但实例摘除与更完整健康治理仍未补齐。

## P0 约束

- 不因为目录存在就按“能力已完成”对外宣传。
- 不再使用“Phase 外移中”“目录骨架先放着”这类迁移期表述。
- 进入 `P1` 之前，先把状态标记、README 口径与专项文档口径全部收敛一致。
