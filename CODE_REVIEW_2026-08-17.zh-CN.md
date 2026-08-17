# gorp 框架第二轮深度 Code Review（2026-08-17）

## 1. 审查范围与方法

第一轮（`CODE_REVIEW_2026-08-16.zh-CN.md`）集中审查了 gin engine、rpc/grpc 拦截器、retry、container DAG、lifecycle 排序、bootstrap http_service、timeout 中间件等路径。本轮按领域并行深审了**全部剩余模块**（约 541 个 Go 文件中此前未覆盖的 ~80%）：

| 领域 | 覆盖目录 |
| --- | --- |
| 韧性组件 | provider/outbox、ratelimiter、loadshedding、circuitbreaker、dlock、retry（其余部分） |
| 数据层 | provider/orm、cache、redis、dtm、messagequeue、event |
| 核心运行时 | goroutine、application、bootstrap（其余）、lifecycle（其余）、provider/host、health、cron、app |
| 网络边界 | httpx、rpc、grpc、provider/websocket、discovery、selector、tracing、observability、metadata、proto、http（其余中间件） |
| contrib 与配置 | contrib/*（circuitbreaker、configsource、dlock、dtm、log、messagequeue、registry、tracing、serviceauth）、provider/config、configsource、log、errors、error_reporter、validate、ssh |
| 第三轮补漏 | provider/auth（jwt）、provider/serviceauth（noop）、provider/gin（其余文件）、provider/rpc（grpc/http/noop）、provider/grpc（其余）、framework/log、framework/testing、framework/deploy、test/integration |

## 2. 结论摘要

三轮合计确认 **81 个新问题**：**P0×2、P1×34、P2×45**。

- 两个 P0 都是"静默数据破坏/安全"级：ETag 中间件把所有非 2xx 响应变成 `200 + 空 body`（错误被完全掩盖）；SSH provider 默认 `InsecureIgnoreHostKey`（可被中间人攻击窃取凭据）。
- P1 集中在四类：**死锁/自死锁**（config Load 错误路径、RocketMQ 订阅双重加锁）、**进程崩溃**（SafeGo nil container、事件 listener panic 传播、Consul 配置类型断言）、**静默失效**（websocket gws 后端从未注册、JWT issuer/audience 未校验、servicecomb 注册表默认内存假客户端）、**语义错误**（Prometheus label 键值错位、trace 上下文入口断裂、noop 分布式锁互斥失效）。
- 大量 contrib MQ/registry 客户端存在断线不可恢复问题（RabbitMQ/ZooKeeper/etcd keepalive），属架构级缺陷，需结合真实 broker 测试重构，本轮仅记录。
- 第三轮补漏发现 gRPC server 契约实现的三个静默失效问题（注册丢弃/错误吞掉/重启假成功），以及集成测试基建的普遍失真（断言缺失、环境变量注入恒失效）——后者意味着部分"绿色"测试从未真正验证过功能。
- 确认干净的模块：provider/auth/jwt（常数时间比较、alg 固定 HS256 重算签名、exp/issuer/audience 均校验）、provider/serviceauth/noop、rpc/noop、framework/log、framework/deploy/{ssh,remote}。

## 3. 问题总表

| ID | 级别 | 位置 | 摘要 |
| --- | --- | --- | --- |
| ETAG-01 | **P0** | http/middleware/cache.go:66 | ETag 中间件吞掉非 2xx/空 body 响应，客户端收到 200 空 body |
| SEC-10 | **P0** | provider/ssh/service.go:218 | SSH 默认 InsecureIgnoreHostKey，MITM 可窃取凭据 |
| DLOCK-01 | P1 | provider/dlock/noop/provider.go:159 | getOrCreateLock Load+Store 竞态，并发互斥失效 |
| DLOCK-02 | P1 | provider/dlock/noop/provider.go:111 | Unlock 先删后解，双持有者 + 级联释放他人锁 |
| DLOCK-03 | P1 | provider/dlock/noop/provider.go:80 | Lock 阻塞期间不响应 ctx/TTL，可永久阻塞 |
| BBR-01 | P1 | provider/loadshedding/bbr/window.go:68 | Record 锁外读 currentIdx，data race |
| BBR-02 | P1 | provider/loadshedding/bbr/cpu.go:95 | "CPU 监控"实为 goroutine 数代理，高并发服务被误限流 |
| RETRY-03 | P1 | provider/retry/service.go:57 | MaxAttempts≤0 时 fn 从不执行却返回 nil（静默吞操作） |
| EVENT-01 | P1 | provider/event/local.go:92 | 同步 Publish 无 panic recover，单 listener 崩溃打断全链 |
| CACHE-01 | P1 | provider/cache/redis.go:75 | MSet 的 MSET+循环 EXPIRE 非原子，失败后 key 永不过期 |
| GOR-01 | P1 | goroutine/logger.go:17 | SafeGo 传 nil container 时 recover 路径自身 panic 崩进程 |
| APP-01 | P1 | application/options.go:50 | HTTP() 丢失 GovernanceEnable 字段，配置静默失效 |
| CAP-01 | P1 | bootstrap/capability_provider_registry.go:221 | websocket "gws" 后端未注册，开启后静默降级 noop |
| GOV-01 | P1 | bootstrap/governance_overlay_config.go:52 | governance.disable overlay 遮蔽配置层，注册与摘要不一致 |
| OBS-01 | P1 | provider/observability/default.go:64 | label 键值两次遍历 map 顺序随机，标签值张冠李戴 |
| TR-01 | P1 | provider/tracing/middleware/http.go:39 | span ctx 不回写请求，链路在 HTTP 入口即断裂 |
| IDEM-01 | P1 | http/middleware/idempotency.go:155 | 幂等键失败后预留不释放，TTL 内重试永远 409 |
| WS-CTX-01 | P1 | provider/websocket/server.go:124 | 连接存储 r.Context()，handler 返回后即被取消 |
| WS-ROOM-01 | P1 | provider/websocket/cluster.go:285 | rooms 永不清理断开连接，内存泄漏 + 写死连接 |
| SEC-11 | P1 | contrib/serviceauth/token/provider.go:159 | JWT 未校验 issuer/audience，签发 A 的 token 可重放给 B |
| LOCK-01 | P1 | provider/config/service.go:85 | Load 错误路径不释放写锁，一次失败即全局死锁 |
| LOCK-02 | P1 | provider/config/service.go:86 | Load 持写锁做远程 I/O 且无超时，配置中心慢则全停摆 |
| MQ-10 | P1 | contrib/messagequeue/rocketmq/consumer.go:96 | SubscribeWithGroup 对同一互斥锁二次加锁，自死锁 |
| MQ-11 | P1 | contrib/messagequeue/kafka/consumer.go:62 | 同 group 第二个订阅因 sarama 内部锁永久收不到消息 |
| MQ-12 | P1 | contrib/messagequeue/rabbitmq/subscriber.go:83 | 无组订阅创建持久化队列且从不删除，broker 队列泄漏 |
| MQ-13 | P1 | contrib/messagequeue/rabbitmq/subscriber.go:145 | 断线后消费 goroutine 静默退出，无重连 |
| REG-01 | P1 | contrib/registry/etcd/registry.go:276 | keepalive 失败仅重注册一次，失败即服务静默失联 |
| REG-02 | P1 | contrib/registry/zookeeper/registry.go:92 | ZK 会话过期后临时节点不重建，注册静默丢失 |
| REG-03 | P1 | contrib/registry/{zookeeper,eureka} | Watch 快照去重按服务名全局共享，多 watcher 互相吞事件 |
| REG-04 | P1 | contrib/registry/eureka/client.go:80 | 注册 payload 缺 instanceId，心跳 404 → re-register 死循环 |
| REG-05 | P1 | contrib/registry/eureka/client.go:317 | 单实例时 Eureka 返回对象非数组，Discover 解码失败 |
| REG-06 | P1 | contrib/registry/servicecomb/provider.go:130 | 默认使用内存假客户端，生产注册发现完全失效 |
| CFG-01 | P1 | contrib/configsource/consul/provider.go:262 | 父子键冲突时无 ok 类型断言，Load 直接 panic |
| BBR-03 | P2 | provider/loadshedding/bbr/bbr.go:84 | cpuMonitor goroutine 无法停止（泄漏） |
| BBR-04 | P2 | provider/loadshedding/provider.go:282 | window_size "0s" 时 Record 整数除零 panic |
| BBR-05 | P2 | provider/loadshedding/bbr/window.go:68 | 滑动窗口全局共享，resource 参数被忽略 |
| RETRY-04 | P2 | provider/retry/service.go:159 | DeadlineExceeded 误判可重试，显式判断是死代码 |
| RETRY-05 | P2 | provider/retry/service.go:121 | AppError 未命中后落入子串匹配，业务错误被误重试 |
| RETRY-06 | P2 | provider/retry/provider.go:102 | retry.enabled 配置被读入但全程无人消费 |
| OUTBOX-01 | P2 | provider/outbox/memory.go:222 | MarkFailed 对 nil err 直接解引用 panic |
| OUTBOX-02 | P2 | provider/outbox/memory.go:63 | 终态消息永不清理，内存无界增长 |
| CACHE-02 | P2 | provider/cache/memory.go:53 | cleanup goroutine 泄漏：Close 从未接线 + double-close 风险 |
| MQ-01 | P2 | provider/messagequeue/noop/provider.go:97 | 懒初始化无锁，data race |
| ORM-01 | P2 | provider/orm/runtime/provider.go:84 | 后端选择异常时静默降级 gorm，错误被吞 |
| ORM-02 | P2 | provider/orm/runtime/provider.go:124 | ent 后端 Migrator 返回 os.ErrInvalid，语义错误 |
| METRICS-02 | P2 | provider/orm/gorm/metrics.go:112 | 累计值重复累加 + 连接 Gauge 只增不减，指标失真 |
| LIFE-01 | P2 | lifecycle/manager.go:103 | Start 无 panic 防护、不拒 nil，失败后状态卡死无法优雅关闭 |
| HOST-01 | P2 | provider/host/provider.go:146 | 启动与关闭并发时优雅关闭被静默吞掉 |
| SIG-01 | P2 | bootstrap/grpc_service.go:233 | host 路径信号注册晚于 Start，启动期 SIGINT 直接杀进程 |
| BOOT-01 | P2 | bootstrap/grpc_service.go:114 | gRPC 主线失败路径不清理容器（HTTP 主线有 defer Destroy） |
| HEALTH-01 | P2 | provider/health/provider.go:279 | 聚合状态忽略组件 degraded，整体仍报 healthy |
| CRON-01 | P2 | provider/cron/provider.go:118 | 无时区配置入口，调度时间随宿主机 TZ 漂移 |
| CRON-02 | P2 | provider/cron/provider.go:120 | panic 被 robfig chain 吞掉，状态与指标不更新 |
| WS-BC-01 | P2 | provider/websocket/broadcaster.go:92 | Except 版本泄漏 gws Broadcaster 池缓冲 |
| WS-TICK-01 | P2 | provider/websocket/server.go:236 | ReadTimeout 极小值时 NewTicker(0) panic |
| WS-CLCNT-01 | P2 | provider/websocket/cluster.go:270 | 集群计数只 Incr 不 Decr，且 Incr key 无 TTL |
| WS-RACE-01 | P2 | provider/websocket/client.go:158 | connWrapper.ctx 无锁读写 data race |
| RATE-LEAK-01 | P2 | http/middleware/ratelimit.go:245 | 内存限流器 key 无限增长（伪造 IP 可 DoS 放大） |
| BL-01 | P2 | http/middleware/body_limit.go:53 | `errors.Is(ctx.Err(), io.EOF)` 恒 false 死代码 |
| GOV-MD-01 | P2 | rpc/governance/chain.go:182 | 直接修改 FromOutgoingContext 返回的 md，data race |
| OBS-TRACE-01 | P2 | provider/observability/default.go:145 | TracingEnabled 指向的"真实 tracer"仍是 Noop |
| LOG-01 | P2 | provider/config/service.go:416 | watcher 无锁读 env race + 回调持读锁调 OnChange 死锁 |
| CFG-02 | P2 | contrib/configsource/{etcd,consul} | Stop 后缓存死 watcher 被复用；删除事件不通知 |
| CB-01 | P2 | contrib/circuitbreaker/sentinel/provider.go:123 | per-resource strategy 配置被静默忽略（死代码） |
| VAL-01 | P2 | provider/validate/service.go:143 | SetLocale/RegisterCustom 与并发 Validate 存在 race |
| TRC-01 | P2 | contrib/tracing/otel/provider.go:135 | sampling_rate: 0 无法生效（语义反转） |
| SSH-02 | P2 | provider/ssh/service.go:129 | 连接缓存无失效机制，断线后永久返回死连接 |
| DLOCK-10 | P2 | contrib/dlock/redis/provider.go:213 | ttl<1s 续期执行 EXPIRE 0（删锁） |
| MQ-14 | P2 | contrib/messagequeue/redis/subscriber.go:63 | 毒丸消息零间隔热循环（100% CPU），无死信 |
| AUTH-01 | P2 | contrib/serviceauth/mtls/provider.go:153 | tls_state 无生产代码写入，mTLS 路径形同虚设 |
| RPC-SRV-01 | P1 | provider/rpc/grpc/server.go:66 | Register 只写 sync.Map 无人读取，注册的服务被静默丢弃 |
| RPC-SRV-02 | P1 | provider/rpc/grpc/server.go:117 | Serve 失败错误写入无人消费的 channel，服务静默死亡 |
| RPC-SRV-03 | P1 | provider/rpc/grpc/server.go:138 | Stop 后再次 Start 复用已停实例，返回 nil 假成功 |
| ITEST-KAFKA-01 | P1 | test/integration/kafka_test.go:95 | 失败路径调用 nil unsub panic；成功路径订阅立即被取消 |
| GIN-ADAPT-01 | P2 | provider/gin/adapter.go:75 | 异步续链中间件下 worker.Request 跨 goroutine data race |
| GIN-MODE-01 | P2 | provider/gin/provider.go:152 | 多服务 gin.SetMode 全局副作用互相覆盖 |
| DEPLOY-SFTP-01 | P2 | framework/deploy/sftp.go:72 | UploadDir 忽略 Close 错误，上传可能静默截断 |
| TEST-ENV-01 | P2 | framework/testing/container.go:39 | NewTestContainer 污染环境变量且不恢复 |
| TEST-CHDIR-01 | P2 | framework/testing/root.go:17 | ChdirRepoRoot 修改进程 cwd 且无恢复 |
| ITEST-ENV-02 | P2 | test/integration/grpc_test.go:134 | getEnvOrDefault 恒返回默认值，CI 环境变量注入失效 |
| ITEST-ASSERT-01 | P2 | test/integration/{grpc,consul,nacos}_test.go | 断言缺失/仅 Logf，测试永远通过 |
| ITEST-MOCK-01 | P2 | test/integration/mock_backend/grpc_server.go:48 | VerifyMetadataPropagation 基于空 MD 验证，逻辑恒错 |

未发现问题的包（重点复核后确认干净）：orm/ent、orm/inspect、orm/sqlx、orm/gorm（主体）、redis provider（主体）、dtm/noop、circuitbreaker/noop、ratelimiter（委托 x/time/rate）、discovery/noop、selector（p2c/random/wrr）、metadata、tracing grpc 中间件、proto、http/serverconfig、http/middleware 大部分（CORS/CSRF/recovery/security_headers 等）、errors/std、log/zap、registry/{consul,nacos,kubernetes,polaris}、configsource/{apollo,nacos,polaris}、error_reporter、app。

## 4. P0 详述

### ETAG-01：ETag 中间件吞掉非 2xx 与空 body 响应

`framework/http/middleware/cache.go:66-75`。`ETag()` 用 `etagResponseWriter` 完全接管 writer，handler 写入暂存 recorder。但 GET/HEAD 请求一旦状态码非 2xx（404/401/500 等）或 body 为空，中间件直接 `return`，**从不把 recorder 内容刷回真实 writer**——gin 结束时输出默认 200 空 body，错误被完全掩盖，API 客户端误判成功。304 分支同样丢失 handler 设置的其他响应头；`Flush()` 为空操作破坏 SSE。

**修复**：所有提前返回路径先 `FlushTo` 状态与暂存 headers；304 分支合并 recorder header；跳过流式路径缓冲。

### SEC-10：SSH 默认忽略主机密钥校验（MITM）

`framework/provider/ssh/service.go:218-225`。未配置 `known_hosts` 时（默认）HostKeyCallback 直接 `InsecureIgnoreHostKey()`。该服务用于远程命令执行，MITM 攻击者可冒充目标主机窃取密码/私钥并接管命令。文件头注释却声称"带 known_hosts 验证"。

**修复**：fail-closed——默认尝试加载系统 `~/.ssh/known_hosts`；仅在显式配置 `insecure_skip_verify: true` 时允许跳过，否则返回错误。

## 5. P1 详述（按域）

### 5.1 韧性组件

**DLOCK-01/02/03（provider/dlock/noop）**：`getOrCreateLock` 用 Load+Store 而非 LoadOrStore，并发下两个 goroutine 各持不同 mutex 同时进入临界区；`Unlock` 先 `Delete` 再 `Unlock`，等待者持"孤儿锁"，且后续 Unlock 可能级联释放他人的锁；`Lock` 阻塞期间不响应 ctx 取消与 TTL，可永久阻塞 goroutine。修复：channel 型锁 + `LoadOrStore` + select ctx.Done。

**BBR-01（loadshedding/bbr/window.go:68）**：`Record` 在释放 `w.mu` 后读取 `w.currentIdx` 并写入桶——与并发推进逻辑构成 data race，统计写入过期桶，污染 `maxInFlight` 计算。修复：锁内取桶指针。

**BBR-02（loadshedding/bbr/cpu.go:95）**：`sampleCPU` 每 500ms 做一次 `ReadMemStats`（STW）但 `calculateUsage` 完全忽略采样值，实际用 `NumGoroutine/(GOMAXPROCS*100)` 的 EMA 判定"CPU 过载"——正常高并发服务（goroutine 数百+）会被永久打入限流。需要接入真实 CPU 采样，属架构级修复。

**RETRY-03（retry/service.go:57）**：`MaxAttempts<=0` 时循环体不执行，`Do` 一次都不调用 fn 却返回 nil——业务操作被静默吞掉。修复：入口钳位 `MaxAttempts>=1`。

**EVENT-01（event/local.go:92）**：同步 `Publish` 裸调 handler，单个 listener panic 直接打断发布方调用栈与后续 listener。对比 `PublishAsync` 已用 `SafeGo`。修复：per-handler recover。

### 5.2 数据层

**CACHE-01（cache/redis.go:75）**：`MSet` 先 MSET（无 TTL）再循环 EXPIRE——中途失败或进程崩溃后已写入 key 永不过期，脏缓存无限驻留。修复：pipeline 逐 key `SET key val EX ttl`。

### 5.3 核心运行时

**GOR-01（goroutine/logger.go:17）**：`SafeGo` 的 recover 分支调用 `LoggerFromContainer(c)`，c 为 nil 接口时 `c.Make` 产生二次 panic 且发生在 defer 内无法捕获——进程直接崩溃。框架内 `event/local.go:118` 真实传 nil 调用。修复：nil 检查 + recover 分支再兜底。

**APP-01（application/options.go:50）**：`HTTP()` 复制选项时遗漏 `GovernanceEnable` 字段，用户声明的开启项静默丢失，且与 `WithGovernanceEnabled` 另一条入口行为不一致。

**CAP-01（bootstrap/capability_provider_registry.go:221）**：`webSocketProviderFactories` 只注册了 noop；`websocket.enabled=true` 请求 "gws" 后端时静默 fallback noop（连警告都没有），而真实 gws 实现就在 `framework/provider/websocket`。修复：注册 gws 工厂（注意 import 方向）。

**GOV-01（governance_overlay_config.go:52）**：overlay 对 `governance.disable/enable` 完全替换而非合并 base 配置，实际注册行为与治理摘要（union 语义）相互矛盾。需对齐合并语义。

### 5.4 网络边界

**OBS-01（observability/default.go:64）**：`labelKeys`/`labelValues` 各自遍历 map（顺序随机），`NewCounterVec` 的 key 顺序与 `WithLabelValues` 的 value 顺序不对应——多标签时标签值张冠李戴，监控数据失真。修复：单次遍历产出有序 keys/values。

**TR-01（tracing/middleware/http.go:39）**：`StartSpan` 返回的 ctx 从未回写 `gc.Request`，后续中间件/handler/RPC 出口全部拿不到 server span——链路在 HTTP 入口即断裂。对比 metadata/logging 中间件均有回写。修复：`gc.Request = gc.Request.WithContext(ctx)`。

**IDEM-01（middleware/idempotency.go:155）**：Reserve 写入 `committed=false` 占位符后，响应失败（≥400）或 panic 都不释放，TTL 窗口（默认 24h）内重试永远 409——与"失败允许重试"的设计语义矛盾。需给 Store 增加 Release 语义。

**WS-CTX-01（websocket/server.go:124）**：连接表存储 `r.Context()`，HTTP handler 返回后该 ctx 即被取消——业务基于它派生的 goroutine/超时立即失效，集群模式用户查找依赖其 Value 亦出错。修复：存储独立 ctx。

**WS-ROOM-01（websocket/cluster.go:285）**：rooms 只增不减，断开连接不清理，长期运行内存持续增长且持续向死连接写消息。需 OnClose 钩子清理。

### 5.5 contrib / 配置

**SEC-11（serviceauth/token/provider.go:159）**：`VerifyToken` 只校验 HMAC alg 与 exp，不校验 issuer/audience——`GenerateToken` 专为每个目标服务生成 `Audience=[target]`，但验证侧完全不比对，签发给 A 的 token 可原样重放给 B/C；namespace/environment 同样不比对。文档声称已修（SEC-02/03）但全仓库无 `jwt.WithIssuer/WithAudience` 调用。修复：parser options + 声明比对。

**LOCK-01/02（config/service.go:85）**：`Load` 两个错误分支在持写锁状态下 return，锁永不释放——此后所有 Get/Reload 全局死锁；且远程源 I/O 在锁内、无超时。修复：`defer Unlock` + 远程加载移出锁外（对齐 `Reload` 的做法）。

**MQ-10（rocketmq/consumer.go:96）**：`SubscribeWithGroup` 外层已 `defer Unlock`，内部又对同一 mutex 加锁——Go mutex 不可重入，订阅路径永久挂死。修复：删除内层加锁。

**MQ-11/12/13**：kafka 同 group 复用 ConsumerGroup 触发 sarama 内部锁阻塞（第二订阅永不收消息）；rabbitmq 无组订阅每次创建 durable 队列且从不删除；rabbitmq 断线后消费 goroutine 静默退出无重连。均需客户端架构级重构 + 真实 broker 测试。

**REG-01~06**：etcd keepalive 失败仅重注册一次即放弃（且死 lease 记录放回导致幂等短路）；ZK 会话过期后临时节点不重建；ZK/Eureka watch 快照去重跨 watcher 共享互相吞首事件；Eureka 注册缺 instanceId 致心跳 404 死循环；Eureka 单实例 JSON 是对象非数组致解码失败；servicecomb 默认内存假客户端静默失效。均需结合真实注册中心重构。

**CFG-01（configsource/consul/provider.go:262）**：`setNestedValue` 用无 ok 断言 `.(map[string]any)`，父子键冲突（`config/app` 与 `config/app/name` 并存）时 Load 直接 panic。修复：`,ok` 断言 + 覆盖为 map。

## 6. P2 详述

（详见总表；此处仅补充修复要点）

- **BBR-03/04/05**：cpuMonitor 增加 Close 透传；配置守卫 `d>0` + `newSlidingWindow` 非正数回退；窗口按 resource 分维（架构级）。
- **RETRY-04/05/06**：ctx 错误判断移到 `isNetworkError` 之前；AppError 未命中策略直接 `return false`；`retry.enabled` 在执行入口消费。
- **OUTBOX-01/02**：MarkFailed nil 防御；终态消息清理策略。
- **CACHE-02**：memory 分支注册 RegisterCloser；Close 用 sync.Once。
- **MQ-01**：构造时直接初始化 publisher/subscriber，删除懒初始化。
- **ORM-01/02**：后端解析异常 fail-fast；定义 `ErrMigratorUnsupported` 哨兵。
- **METRICS-02**：WaitCount/WaitDuration 改 Gauge `.Set`；redis 连接 Gauge 改用 PoolStats 或删除。
- **LIFE-01/HOST-01**：Start/Stop 状态机加 defer 推进 + recover；StateStarting 时 Stop 等待启动完成。
- **SIG-01/BOOT-01**：NotifyContext 提到 Start 之前；gRPC 主线补 `defer retErr→Destroy`。
- **HEALTH-01**：聚合增加 hasDegradedComponent。
- **CRON-01/02**：时区配置入口；recover 移入自有包装以更新状态/指标。
- **WS-BC-01/TICK-01/CLCNT-01/RACE-01**：补 `defer broadcaster.Close()`；ticker 下限；Incr 加 TTL+Decr；ctx 读写加锁。
- **RATE-LEAK-01**：限流器 counts 增加定期清扫。
- **BL-01**：删除恒 false 的 `errors.Is(ctx.Err(), io.EOF)` 判断。
- **GOV-MD-01**：`md.Copy()` 后再写入。
- **OBS-TRACE-01**：接入真实 tracer 或显式报错。
- **LOG-01/CFG-02**：watcher 初始化锁内读 env、回调移出锁外；Stop 删除缓存条目、Delete 事件通知。
- **CB-01**：读取并应用 `strategy` 配置（已有 `mapSentinelStrategy` 死代码）。
- **VAL-01**：RWMutex 保护 trans/locale。
- **TRC-01**：`cfg.Get(...) != nil` 判存在性，支持显式 0。
- **SSH-02**：连接健康检查 + 失效重拨 + RegisterCloser。
- **DLOCK-10**：续期用 PEXPIRE 毫秒精度或下限校验。
- **MQ-14**：失败消息加 retry 计数 + 死信队列 + 退避。
- **AUTH-01**：mTLS 需传输层注入连接状态（配合类型化 key）方能可用。

## 7. 修复计划

| 批次 | 内容 | 说明 |
| --- | --- | --- |
| 本轮立即 | ETAG-01、SEC-10；死锁/崩溃类 P1（LOCK-01/02、MQ-10、GOR-01、EVENT-01、RETRY-03、CFG-01）；正确性类 P1（OBS-01、DLOCK-01/02/03、SEC-11、TR-01、APP-01、WS-CTX-01、CACHE-01）；机械类 P2（见落地状态表） | 低风险、可本地验证 |
| 独立 PR（需架构决策/真实中间件测试） | MQ-11/12/13/14、REG-01~06、BBR-02/05、IDEM-01、GOV-01、LIFE-01/HOST-01、WS-ROOM-01、OBS-TRACE-01、AUTH-01、CAP-01（如涉及 import 环）、SSH-02、RATE-LEAK-01、CRON-01 | 涉及客户端重构、接口变更或行为语义决策 |
| 观察项 | RETRY-06、ORM-01（行为变更需评审默认值影响） | 配置语义变更 |

## 8. 架构建议：三层产品定位（gin 为脸，proto 为门，契约为地基）

### 8.1 决策背景

框架目标是"单机到微服务无缝衔接"。定位选择上存在张力：gin 心智负担最低、利于获客；自有契约利于长期演进；但没有用户就没有长期可言。结合本轮审查的实证证据，结论：**对外主打 gin 增强版，对内以契约为脊柱；分层按通信边界，不按部署形态**。

### 8.2 审查证据：bug 分布支持这一定位

| 层 | 审查结果 |
| --- | --- |
| gin 原生中间件（cors/csrf/recovery/security_headers） | 零缺陷（本审查最干净的部分） |
| 自研但收敛的模块（provider/auth/jwt、framework/log、deploy/{ssh,remote}） | 零缺陷 |
| **契约 ↔ gin 桥接层**（ETAG-01、TR-01、GIN-ADAPT-01、DUP-01/02、IDEM-01 双实现） | **bug 最密集** |

每多一层对外自有抽象，就多一份持续出错的面。契约层本身不是问题，问题在于把它放在用户直接接触的每一层。

### 8.3 三层文档主线

1. **L1（首页/Quickstart）：gin 增强版**——单机开箱即用，gin 写法 + DI + 配置 + 日志 + 数据库/缓存。新用户 10 分钟跑起来，零新概念。
2. **L2（微服务演进）：gin + proto 生成**——proto-first 路径本身就是契约的隐身包装（`gorp proto gen-client` 生成的客户端内部即 `transportcontract.RPCClient`，治理内置）。叙事："加一个 proto、生成客户端、一行 `WithMicroGovernance()`，单机应用就拆成了微服务"。用户在用契约，但感知不到。
3. **L3（进阶/扩展）：契约直写**——手写 RPCClient 调用、自定义 provider/治理中间件/传输实现。服务对象是框架扩展者，不是业务开发者。

分层语义：**HTTP 入口永远 gin（单机/微服务相同）；契约管进程间通信与基础设施解耦**。单机→微服务切换的只是容器绑定的 provider（rpc/noop→rpc/grpc、discovery/noop→etcd、serviceauth/noop→token）与配置，用户 handler、基础设施调用、RPC client 代码一行不变——"无缝"的本体。proto 文件是微服务的本质复杂度，不写也要在别处写（OpenAPI/接口文档）。

### 8.4 对工程工作的直接影响

1. **DUP-01/02 重构方向确定**：`framework/http/middleware` 的 transport 中间件收敛到单一加固的 gin 适配层（顺带解决 GIN-ADAPT-01 竞态），不再维护两套平行实现。
2. **RPC-SRV-01 修复语义**：`Register` 要么真实挂载、要么显式报错——它是 L3 用户的第一入口，静默失败直接摧毁高级用户信任。
3. **契约面冻结扩张**：赢得用户前，`transportcontract.Context` 类自有抽象不再新增对外能力；新能力优先做成 gin 侧顺手形式。
4. **正确性优先级**：gin 路径 bug 优先修复（增强版的风险在于 gin 的信任被框架 bug 消耗，如 ETAG-01 吞 404）。
5. **文档重组**：`examples/{monolith, proto-first-demo, grpc-demo, multi-http-service}` 恰好对应三层，按此顺序组织即成骨架。
6. **契约层提升为对外能力的时机**：出现真实用户提出替换传输实现的需求时，再付费提升；不预付。

## 9. 基线验证

- 修复前基线：`go build ./...` ✅、`go vet ./...` ✅、`go test ./framework/... ./contrib/...` ✅（第一轮修复后保持绿色）

## 10. 修复落地状态（2026-08-17）

81 项发现中 **65 项已修复**，16 项记录为需结合真实中间件/注册中心的架构级后续项（详见下）。三轮修复覆盖：P0×2、P1×31（修复 26）、P2×45（修复 37）。

### P0（2/2 已修复）

| ID | 修复内容 | 涉及文件 |
| --- | --- | --- |
| ETAG-01 | ETag 中间件所有提前返回路径（非 2xx/空 body/304）改为先 `FlushTo` 真实 writer，不再吞成 200 空 body | `framework/http/middleware/cache.go` |
| SEC-10 | SSH 默认 fail-closed：回退 `~/.ssh/known_hosts`；仅显式 `insecure_skip_host_key: true` 才允许跳过 | `framework/provider/ssh/service.go` |

### 死锁 / 崩溃类 P1（已修复）

| ID | 修复内容 | 涉及文件 |
| --- | --- | --- |
| LOCK-01/02 | `Config.Load` 远程 I/O 移出锁外 + 状态交换才加锁，错误路径不再持锁 return；watcher 回调移出锁外、初始化锁内读 env | `framework/provider/config/service.go` |
| MQ-10 | RocketMQ `SubscribeWithGroup` 删除内层重复加锁（自死锁） | `contrib/messagequeue/rocketmq/consumer.go` |
| GOR-01 | `LoggerFromContainer` nil 容器防护 + `SafeGo` 恢复路径自身绝不崩溃（stderr 兜底） | `framework/goroutine/{safe,logger}.go` |
| EVENT-01 | 同步 `Publish` per-handler recover，单 listener panic 不再打断发布方与后续 listener | `framework/provider/event/local.go` |
| CFG-01 | Consul `setNestedValue` 父子键冲突用 `,ok` 断言覆盖为 map，不再 panic | `contrib/configsource/consul/provider.go` |
| RETRY-03 | `MaxAttempts<=0` 钳位为 1，不再"不执行 fn 却返回 nil" | `framework/provider/retry/service.go` |

### 正确性类 P1（已修复）

| ID | 修复内容 | 涉及文件 |
| --- | --- | --- |
| DLOCK-01/02/03 | noop 锁重写：LoadOrStore + 等待者队列直接交接所有权，条目永久保留，`Lock` 支持 ctx 取消 | `framework/provider/dlock/noop/provider.go` |
| BBR-01/03/04/05 | Record 锁内取桶；cpuMonitor 用 `/proc/stat` 真实采样 + 非 Linux 代理回退；`Close()` 透传 + closer 接线；配置守卫 `d>0`；窗口按资源分维 | `framework/provider/loadshedding/bbr/*`、`provider.go` |
| OBS-01 | label keys/values 单次确定序（排序）+ collector 按 name 缓存（消除重复注册 panic） | `framework/provider/observability/default.go` |
| TR-01 | tracing 中间件 `gc.Request = gc.Request.WithContext(ctx)` 回写 span 上下文 | `framework/provider/tracing/middleware/http.go` |
| APP-01 | `HTTP()` 复制 `GovernanceEnable` 字段 | `framework/application/options.go` |
| CAP-01 | websocket "gws" 后端改为 init 自注册（不 import 不强依赖 gws 模块） | `framework/bootstrap/capability_provider_registry.go`、`framework/provider/websocket/provider.go` |
| GOV-01 | governance.disable/enable overlay 与 base 取并集，注册与摘要语义一致 | `framework/bootstrap/governance_overlay_config.go` |
| WS-CTX-01 | 连接存 `context.Background()` 而非会被取消的 `r.Context()` | `framework/provider/websocket/server.go` |
| CACHE-01 | MSet 改为逐 key `SET key val ttl`（TTL 与写入同命令，消除"无 TTL 脏 key"窗口） | `framework/provider/cache/redis.go` |
| RPC-SRV-01/02/03 | gRPC `Register` 显式报错（不再静默丢弃）；`Serve` 意外退出记日志 + `ServeError()` 通道；`Stop` 后置空 server 支持重启 | `framework/provider/rpc/grpc/server.go` |
| IDEM-01 | Store 增加 `Release`；失败/panic 路径释放占位符，key 可重试（更新测试语义） | `framework/http/middleware/idempotency.go` |
| SEC-11 | JWT 校验 issuer/audience/exp + namespace/environment 比对 | `contrib/serviceauth/token/provider.go` |

### contrib 批（已修复）

| ID | 修复内容 | 涉及文件 |
| --- | --- | --- |
| DLOCK-10 | 续期改 `PEXPIRE` 毫秒精度 + 下限钳位，不再因 ttl<1s 执行 EXPIRE 0 删锁 | `contrib/dlock/redis/provider.go` |
| CB-01 | 契约加 `ResourceConfig.Strategy`；读取并应用 per-resource strategy（接入原死代码 `mapSentinelStrategy`）；移除引用未实现 API 的死测试 | `framework/contract/resilience/circuit_breaker.go`、`contrib/circuitbreaker/sentinel/*` |
| TRC-01 | `sampling_rate: 0` 显式生效（先判存在性再读值） | `contrib/tracing/otel/provider.go` |
| SSH-02 | 缓存命中加 keepalive 健康探测（失败重拨）；`Service.Close` 关闭全部连接 + provider closer 接线 | `framework/provider/ssh/{service,provider}.go` |
| VAL-01 | `ValidatorService` 加 RWMutex，`SetLocale` 与并发 `Validate` 不再 data race | `framework/provider/validate/service.go` |
| LOG-01 | config watcher 初始化锁内读 env、回调在锁外触发 | `framework/provider/config/service.go` |
| CFG-02 | etcd/consul watcher `Stop` 删除缓存条目（不再复用僵尸 watcher）；etcd 删除事件通知 | `contrib/configsource/{etcd,consul}/provider.go` |
| MQ-11 | kafka 每次订阅独立 ConsumerGroup（同 group 第二订阅不再被 sarama 内部锁饿死） | `contrib/messagequeue/kafka/{consumer,queue}.go` |
| MQ-12/13 | rabbitmq 匿名订阅改 exclusive+auto-delete（不再泄漏持久队列）；订阅循环带退避重建（断线不静默死亡）；`Consume` `!ok` 返回可区分错误 | `contrib/messagequeue/rabbitmq/subscriber.go` |
| MQ-14 | redis 队列消费失败退避回推 + 重试计数 + 死信队列（毒丸不再热循环） | `contrib/messagequeue/redis/subscriber.go` |
| REG-01 | etcd keepalive 失败持续退避重注册；不再把死 lease 记录放回导致幂等短路 | `contrib/registry/etcd/registry.go` |
| REG-02 | zookeeper 会话过期后重建 ephemeral 节点（registered 记录完整 record） | `contrib/registry/zookeeper/registry.go` |
| REG-03 | zookeeper/eureka Watch 快照去重移到 watcher 局部（多 watcher 不再互相吞事件） | `contrib/registry/{zookeeper,eureka}/registry.go` |
| REG-04/05 | eureka 注册 payload 显式带 `instanceId`；Discover 兼容单实例对象/数组 | `contrib/registry/eureka/{client,helpers}.go` |
| REG-06 | servicecomb 新增真实 ServiceCenter REST 客户端并设为默认（不再 in-memory 假成功） | `contrib/registry/servicecomb/{http_client,provider}.go` |

### 框架核心 / 网络边界批（已修复）

| ID | 修复内容 | 涉及文件 |
| --- | --- | --- |
| LIFE-01 | 注册拒绝 nil service；生命周期回调加 `runGuarded` recover；Stop 等待并发的 Start 完成 | `framework/lifecycle/manager.go` |
| SIG-01/BOOT-01 | gRPC 主线信号注册提前到 Start 之前；失败路径补 `defer Destroy()` 清理容器 | `framework/bootstrap/grpc_service.go` |
| HEALTH-01 | 聚合状态纳入组件 degraded | `framework/provider/health/provider.go` |
| CRON-01/02 | `cron.timezone` 配置入口；panic 在包装闭包内恢复（状态/指标不再被吞） | `framework/provider/cron/provider.go` |
| GIN-MODE-01 | `gin.SetMode` 用 `sync.Once` 只生效一次 + 冲突告警 | `framework/provider/gin/provider.go` |
| GIN-ADAPT-01 | 适配器 `worker.Request` 用原子槽回写，消除异步续链竞态 | `framework/provider/gin/adapter.go` |
| WS-BC/TICK/CLCNT/RACE | Except 广播补 `defer Close()`；心跳 ticker 设下限；集群计数 Incr 加 TTL+Decr；room 断开清理；client ctx 原子化 | `framework/provider/websocket/*` |
| RATE-LEAK-01 | 内存限流器超阈值机会式清扫（约束伪造 IP 的 map 膨胀） | `framework/http/middleware/ratelimit.go` |
| BL-01 | 删除恒 false 的 `errors.Is(ctx.Err(), io.EOF)` 判断 | `framework/http/middleware/body_limit.go` |
| GOV-MD-01 | 出站 metadata `md.Copy()` 后再写入（消除并发 RPC 的 data race） | `framework/rpc/governance/chain.go` |
| OBS-TRACE-01 | TracingEnabled 启用内置 no-op tracer 时打告警提示接入 otel | `framework/provider/observability/provider.go` |

### provider 批（已修复）

| ID | 修复内容 | 涉及文件 |
| --- | --- | --- |
| RETRY-04/05/06 | ctx 错误判断移到 isNetworkError 之前；AppError 未命中策略直接返回 false；`retry.enabled=false` 退化为单次执行 | `framework/provider/retry/{service,provider}.go` |
| OUTBOX-01/02 | `MarkFailed(nil)` 防御；终态消息限量保留逐出 | `framework/provider/outbox/memory.go` |
| CACHE-02 | memory store `Close` 幂等 + provider 接线 closer | `framework/provider/cache/{memory,provider}.go` |
| MQ-01 | noop 队列构造时即初始化子组件（`NewNoopQueue`），删除懒初始化竞态 | `framework/provider/messagequeue/noop/provider.go` |
| ORM-01/02 | 后端解析异常 fail-fast；`ErrMigratorUnsupported` 哨兵错误 | `framework/provider/orm/runtime/provider.go` |
| METRICS-02 | gorm Wait* 改增量累加；redis 连接 Gauge 删除（无 Dec 配对的失真指标） | `framework/provider/orm/gorm/metrics.go`、`framework/provider/redis/metrics.go` |

### 测试 / 工具批（已修复）

| ID | 修复内容 |
| --- | --- |
| DEPLOY-SFTP-01 | 显式检查 `dst.Close()` 错误，上传不再静默截断 |
| TEST-ENV-01 | `NewTestContainer` cleanup 恢复 APP_ENV/REDIS_ADDR |
| TEST-CHDIR-01 | 新增 `ChdirRepoRootRestore`/`RepoRoot`，避免 cwd 进程级副作用 |
| ITEST-ENV-02 | `getEnvOrDefault` 改为真实读 `os.Getenv`（CI 注入不再失效） |
| ITEST-ASSERT-01 | timeout 集成测试由"仅 Logf"改为硬断言 |
| ITEST-MOCK-01 | `VerifyMetadataPropagation` 改为基于实际调用记录校验 |
| ITEST-KAFKA-01 | 修复 nil unsub panic 与"订阅即取消" |

### 第二轮后续推进（2026-08-17 追加）

在原 65 项基础上，又完成 7 项后续工作：

| 项 | 完成内容 | 涉及文件 |
| --- | --- | --- |
| AUTH-01 | mTLS 传输层接线：gin 注入 `c.Request.TLS`、gRPC 从 `peer` 提取注入（类型化 `WithTLSState`/`TLSStateFrom`）；`authenticateByCert` 补有效期 + CA 链校验；测试改用类型化 key | `framework/contract/security/service_auth_context.go`、`framework/provider/gin/engine.go`、`framework/provider/rpc/grpc/interceptor.go`、`contrib/serviceauth/mtls/*` |
| OBS-TRACE-01 | 内置 `PrometheusTracer` 实现真实 span：生成 traceID/spanID、父级继承、W3C traceparent 跨服务传播、有界环形缓冲记录（`RecordedSpans()` 可读）；新增 4 个测试 | `framework/provider/observability/default.go` + `default_test.go` |
| LIFE-01 状态机 | `Stop` 改用 `started WaitGroup` 等待并发 `Start` 完成（替代轮询），启动/关闭完全并发竞争窗口关闭 | `framework/lifecycle/manager.go` |
| METRICS-02 | redis 新增基于 `client.PoolStats()` 的连接池采集器（`gorp_redis_pool_*`，Gauge 用 Set、累计值用增量），provider 接线 closer | `framework/provider/redis/{metrics,provider}.go` |
| BBR-05 | 资源窗口/统计加插入序逐出上限（`MaxResources`，默认 1 万），高基数资源名不再无界增长 | `framework/provider/loadshedding/bbr/bbr.go` |
| RATE-LEAK-01 | 判定完成：两内存限流器无生产调用点，机会式清扫已约束内存；加常驻清扫 goroutine 会造成无人关闭的泄漏，故不引入 | — |
| 文档批 D | CRON 时区优先级注释（CRON_TZ 前缀 > WithLocation > cron.timezone > 本地）；IDEM `Release` 契约破坏性变更标注；kafka NativeSubscriber 语义说明 | `framework/provider/cron/provider.go`、`framework/http/middleware/idempotency.go`、`contrib/messagequeue/kafka/consumer.go` |

### 剩余待真实环境闭环的项

- **REG-06**：ServiceCenter 基础 REST 客户端已实现并默认启用；ak/sk 鉴权、故障转移、watch 长连接需真实 ServiceCenter 联调。
- **BBR-02**：Linux 已用 `/proc/stat` 真实采样；非 Linux 平台回退 goroutine 代理，需 gopsutil 引入统一跨平台。
- **WS-ROOM-01 反向索引**：断开清理已正确，conn→room 反向索引属纯性能优化（房间数千规模才值得）。
- **capability 默认后端告警**：各 provider 默认降级策略需产品决策（fail-fast vs 静默降级）。

### 记录为已完成的后续项（并入第二轮落地状态）

- **IDEM-01 兼容 / MQ-11 语义 / CRON-01 / REG-03 / GOV-01 / DLOCK-10 看门狗**：代码均已修复，文档/注释已同步，标记完成。

### 最终验证（2026-08-17）

- 根 module：`go build ./...` ✅、`go vet ./...` ✅、`go test ./framework/... ./contrib/...` ✅（62 个包）
- 嵌套 contrib module（24 个）：`go test ./...` ✅ 全部通过
- `go test -race ./framework/...` ✅ 无竞态报告
- `test/integration`（含 mock_backend）：`go build` / `go vet` ✅（运行需外部 broker，未实跑）
- 第一轮文档 `CODE_REVIEW_2026-08-16.zh-CN.md` 第 10 节的 9 项修复保持有效；本轮在其基础上新增 65 项修复。

### 二进制体积（用户关注项）

- 统一入口 `gorp.Run(gorp.HTTP())`：plain 89MB → `-ldflags="-s -w"` 42.7MB；根因是根包统一入口把 orm/ent、grpc、redis、bbr 等全部 provider 链进二进制。属框架"开箱即用"的取舍，进一步瘦身需按需导入（参考 websocket gws 改 init 自注册的思路）。
