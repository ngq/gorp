# gorp 框架深度 Code Review 报告

> 审查日期：2026-07-20  
> 审查范围：根模块、`framework/`、`contrib/`、CLI/模板及 `hade/documents/` 公开文档  
> 审查方式：源码审查、文档与实现交叉核对、并发/安全专项审查、测试与静态检查基线验证  
> 总体结论：**DONE_WITH_CONCERNS，不建议按当前状态发布**

## 1. 执行摘要

本次审查确认了多项会影响生产安全、运行时可靠性和框架核心契约的缺陷：

- gRPC 服务认证在依赖解析失败时静默降级为未认证服务。
- JWT 与服务间认证缺少密钥时使用公开固定密钥。
- JWT 的 issuer/audience 校验允许通过省略 claim 绕过。
- 根模块引用不存在的本地 replace，导致全量测试和 `go vet` 无法运行。
- HTTP timeout、Provider 依赖顺序、远程配置、生命周期回滚和正常关停行为与框架承诺不符。
- gRPC 连接池、Container、Retry、Outbox、Health Checker 等共享组件存在竞态、资源泄漏或错误语义。
- 部分公开文档中的 API、默认配置和运行时行为与实现不一致。

建议先完成 P0/P1 问题并建立对应回归测试，再继续发布或扩大生产使用范围。

## 2. 严重级别

| 级别 | 定义 |
| --- | --- |
| P0 | 可直接造成认证绕过、令牌伪造等安全事故，必须立即修复 |
| P1 | 阻断构建发布，或可造成核心能力失效、资源泄漏、服务不可用 |
| P2 | 特定条件下产生错误行为、数据竞争、文档误导或维护风险 |

## 3. 发现汇总

| ID | 级别 | 问题 | 主要位置 |
| --- | --- | --- | --- |
| SEC-01 | P0 | gRPC 服务认证依赖失败时 fail-open | `framework/provider/rpc/grpc/server.go:236` |
| SEC-02 | P0 | JWT 与 ServiceAuth 使用公开默认密钥 | `framework/provider/auth/jwt/provider.go:37`、`contrib/serviceauth/token/provider.go:22` |
| SEC-03 | P0 | JWT issuer/audience 可通过省略 claim 绕过 | `framework/provider/auth/jwt/jwt_service.go:127` |
| BUILD-01 | P1 | 根模块引用不存在的本地 replace | `go.mod:214` |
| HTTP-01 | P1 | timeout 到期后仍可能永久等待 handler | `framework/http/middleware/timeout.go:45` |
| CONFIG-01 | P1 | 远程 ConfigSource 未接入默认 Config | `framework/provider/config/provider.go:31` |
| DI-01 | P1 | `DependsOn` 不参与真实 Provider 装配顺序 | `framework/container/container.go:413` |
| LIFE-01 | P1 | 正常关停不销毁 Container | `framework/bootstrap/http_service.go:289` |
| LIFE-02 | P1 | `OnStarted` 失败时当前服务不会回滚 | `framework/lifecycle/manager.go:152` |
| RPC-01 | P1 | gRPC 连接池关闭/建连竞态导致泄漏 | `framework/provider/rpc/grpc/client.go:185` |
| RETRY-01 | P1 | 资源级 Retry 可重试条件被默认策略覆盖 | `framework/provider/retry/service.go:45` |
| OUTBOX-01 | P1 | MemoryOutbox 并发重复投递 | `framework/provider/outbox/memory.go:82` |
| RATE-01 | P1 | RateLimiter 零值配置导致所有资源被限流 | `framework/provider/ratelimiter/provider.go:50` |
| HEALTH-01 | P2 | Health Checker 共享 map 与超时语义不安全 | `framework/provider/health/provider.go:107` |
| CONTAINER-01 | P2 | Make/Destroy 交错可能遗漏资源回收 | `framework/container/container.go:207` |
| RPC-02 | P2 | gRPC Serve 错误被吞掉且状态仍为 running | `framework/provider/rpc/grpc/server.go:110` |
| DOC-01 | P2 | Config 文档使用不存在或签名错误的 API | `hade/documents/reference/config.md:581` |
| APP-01 | P2 | `RunContext` 运行期间忽略 context 取消 | `framework/application/run.go:47` |
| DOC-02 | P2 | HTTP 地址默认值、必填校验与文档冲突 | `framework/bootstrap/config_schema.go:23` |
| DOC-03 | P2 | `app.debug` 不会按文档开启 pprof | `framework/bootstrap/http_service.go:321` |

## 4. P0 安全问题

### SEC-01：gRPC 服务认证失败时 fail-open

**证据**

`framework/provider/rpc/grpc/server.go:236-242` 只有在 `ServiceAuthKey` 成功解析且类型断言成功时才添加认证拦截器。`Make` 返回错误或对象类型错误时，代码不返回错误，gRPC server 仍继续创建和启动。

**影响**

配置上看似启用了服务认证，但配置错误、Provider 初始化失败或绑定类型错误会让服务以无认证状态运行。调用方无法从启动结果察觉降级。

**建议**

- 认证能力被选择或 `ServiceAuthKey` 已绑定时，解析失败必须阻止 server 构建或启动。
- 区分显式 noop 与认证初始化失败，只有显式 noop 才允许无认证启动。
- 增加认证 Provider 构造失败、类型错误、缺失配置的启动回归测试。

### SEC-02：公开默认密钥允许伪造令牌

**证据**

- `framework/provider/auth/jwt/provider.go:37-39` 定义固定 JWT 密钥。
- `framework/provider/auth/jwt/provider.go:113-120` 缺少配置时仅记录日志并继续运行。
- `contrib/serviceauth/token/provider.go:22-24,90-95` 对服务间令牌采用相同策略。

**影响**

任何知道框架源码的人都能使用固定密钥签发合法用户 JWT 或服务身份令牌。日志警告不能构成安全边界。

**建议**

- 认证启用时缺少密钥必须返回错误并阻止启动。
- 拒绝已知占位符和低强度密钥。
- 示例项目可通过显式 development 模式使用临时随机密钥，但不能共享固定值。
- CI 增加模板和示例配置的占位密钥扫描。

### SEC-03：issuer/audience 校验允许缺失字段

**证据**

`framework/provider/auth/jwt/jwt_service.go:127-133` 只在 token claim 非空时比较 issuer/audience。当验证器配置了期望值，但 token 省略字段时，校验被跳过。

**影响**

共享密钥、多发行方或多受众场景中，原本不属于当前服务的 token 可能被接受。

**建议**

验证器配置了 issuer 或 audience 时，要求对应 claim 必须存在且精确匹配，并补充“字段缺失”和“字段不匹配”两组测试。

## 5. P1 发布与运行时问题

### BUILD-01：根模块无法执行测试或静态检查

`go.mod:214` 将 `google.golang.org/genproto` replace 到 `./examples/nop-go/deploy/docker/genproto-dummy`，但该目录不存在且未被 Git 跟踪。

实际执行结果：

```text
go test ./... -> FAILED
go vet ./...  -> FAILED
open examples\nop-go\deploy\docker\genproto-dummy\go.mod:
The system cannot find the path specified.
```

应删除失效 replace，或提交完整且可独立解析的 dummy module。CI 必须从干净 checkout 执行 `go test ./...` 和 `go vet ./...`。

### HTTP-01：timeout 不保证按时返回

`framework/http/middleware/timeout.go:45-50` 在 deadline 到达后继续等待 `<-done`。如果 handler 忽略 context 或永久阻塞，请求不会返回 504。transport contract 版本创建的新 context 也没有写回请求链。

修复需要同时解决：

- 将 deadline context 传递给下游。
- timeout 到达后立即结束上游请求。
- 隔离 response writer，防止超时响应与后台 handler 并发写入。
- 增加永不返回 handler、晚写响应、客户端断连测试。

### CONFIG-01：远程配置能力没有进入默认主线

`framework/provider/config/provider.go:31-32` 始终调用无 source 的 `NewService()`。Bootstrap 后续虽然注册外部 ConfigSource 并调用 Reload，但 `framework/provider/config/service.go:315-327` 仍读取该 Config 实例自身的 nil source。

结果是 Nacos、Etcd、Consul、Apollo 等远程配置不会覆盖默认 Config，热加载和远端兜底承诺不成立。建议让 Config provider 显式解析并持有 ConfigSource，或者建立统一的本地+远端组合配置 Provider。

### DI-01：Provider DAG 只用于展示，不用于装配

`framework/container/container.go:413-419` 按调用方传入顺序逐个注册 Provider；非延迟 Provider 会立即 Register 和 Boot。`DependsOn()` 仅用于生成 DAG，没有参与拓扑排序或缺失依赖校验。

这与 ADR-004 中“注册顺序由框架控制，用户无需关心”的说明冲突。建议先收集 Provider，再按依赖拓扑排序，针对缺失依赖和环路 fail-fast，最后按排序结果执行 Register/Boot。

### LIFE-01：正常关停跳过 Container Destroy

`framework/bootstrap/http_service.go:289-294` 只在启动过程返回错误时 Destroy Container。正常收到信号并完成 host shutdown 后返回 nil，Container closers 不会执行。

DB、Redis、Tracing、Registry、消息队列等 Provider 都会注册 closer，因此正常关停可能丢失 trace flush、遗留连接和后台 goroutine。建议 runtime 创建成功后无条件 defer Destroy，并用 `errors.Join` 合并 shutdown 与 destroy 错误。

### LIFE-02：OnStarted 失败后当前服务泄漏

`framework/lifecycle/manager.go:144-162` 先 Start 服务，再调用 OnStarted，最后才将服务加入 `started`。OnStarted 返回错误时，回滚列表不包含刚刚成功启动的当前服务。

应在 Start 成功后立即加入回滚列表，或者在 OnStarted 失败分支显式 Stop 当前服务。测试需验证当前服务和此前服务均按逆序停止。

### RPC-01：gRPC 连接池存在并发泄漏

`framework/provider/rpc/grpc/client.go:185-201` 的 Close 使用 `mu`，但 `getConn` 没有检查 closed，也没有与 Close 使用相同的生命周期同步。并发 cache miss 会创建多个连接并相互覆盖，Close 的 Range 完成后也可能出现新连接。

建议使用统一状态锁和双重检查，或 singleflight/LoadOrStore；关闭未被采用的连接，并在 closed 后拒绝所有建连操作。

### RETRY-01：资源级重试策略没有完整生效

`framework/provider/retry/service.go:45-55` 已接收资源 policy，但可重试判断调用固定读取 DefaultPolicy 的 `IsRetryable`。资源级 MaxAttempts/退避生效，RetryableErrors/Codes/GRPCCodes 却被默认策略覆盖。

此外 Retry singleton 内部的 `*rand.Rand` 和 ResourcePolicies map 缺少并发保护。建议让 `isRetryable(err, policy)` 使用本次解析出的不可变策略快照，并用并发安全随机源和 RWMutex 保护策略更新。

### OUTBOX-01：MemoryOutbox 会并发重复投递

`framework/provider/outbox/memory.go:82-89` 只快照 pending 消息，没有在锁内原子 claim 为 processing；多个 Process 可同时发送同一消息。每次 Emit 还会启动独立后台 Process，放大重复投递和 goroutine 数量。

建议使用单 worker、有界队列和生命周期 context，并通过锁内状态迁移或持久存储 CAS/租约实现 claim。

### RATE-01：RateLimiter 默认配置实际为全拒绝

`framework/provider/ratelimiter/provider.go:50-64` 使用 100/200 fallback 创建了一个字段 limiter，但没有把规范化值写回 config。后续按资源创建 limiter 时读取的仍是零值 config，得到 rate=0、burst=0。

应在构造阶段规范化并保存 config，或无资源覆盖时复用默认 limiter，并增加零值配置测试。

## 6. P2 并发、生命周期与文档问题

### HEALTH-01：Health Checker 快照和超时不可靠

`framework/provider/health/provider.go:107-110` 解锁后直接持有共享 map 引用；并发 AddChecker/AddDependency 会与 map 迭代竞争。当前实现也依赖 checker 主动监听 context，checker panic 会直接终止进程，忽略 context 的 checker 会让整体检查永久等待。

建议锁内复制 map、包装 panic recovery，并明确采用协作式超时还是框架强制超时语义。

### CONTAINER-01：Make/Destroy 交错会遗漏 closer

Make 只在入口检查 destroyed。factory 可在检查后创建资源，而 Destroy 已经完成 closer 快照；此时 RegisterCloser 不检查 destroyed，新资源将永远不会被回收，Make 仍可能成功返回。

建议用生命周期锁或 active-resolution WaitGroup 阻止 Destroy 与初始化交错；Destroy 后 RegisterCloser 应立即关闭资源或返回错误。

### RPC-02：gRPC Serve 错误被丢弃

`framework/provider/rpc/grpc/server.go:110-115` 在 goroutine 中调用 Serve，但忽略返回值并立即报告启动成功。Serve 异常退出后 running 状态仍为 true。

建议增加 serveErr 通道，更新状态并向 host/lifecycle 报告异步错误。

### APP-01：RunContext 只在启动前检查取消

`framework/application/run.go:47-60` 检查一次 context 后调用不接收 context 的启动函数。运行期间 cancel/deadline 不会触发服务关闭，嵌入式使用和测试只能依赖 OS signal。

应将 context 贯穿 BootHTTPService、RunHTTP 和 Host.Start/Shutdown，关停 select 同时监听调用方 context 与系统信号。

### DOC-01：Config API 文档无法按示例编译

`hade/documents/reference/config.md` 使用了契约中不存在的 `GetDuration`、`GetStringSlice`，并把实际返回 `(ConfigWatcher, error)` 的 Watch 描述为 `<-chan struct{}`。

需要选择一种方向：补齐契约和实现，或把文档改为当前已有的 Unmarshal/显式解析与 ConfigWatcher API。

### DOC-02：HTTP 地址默认值与校验冲突

文档称 `app.address`/`http.addr` 非必填且默认 `:8080`，Gin Provider 也有 fallback；但 `framework/bootstrap/config_schema.go` 在启动前强制要求 `app.address`，存在 `server.http` 时还会额外要求 addr。

应确定唯一 canonical key，并让文档、schema 和 Provider fallback 使用相同语义。

### DOC-03：app.debug 不会开启 pprof

文档声称 `app.debug=true` 后 pprof 可用，实际注册仅由内部 `EnablePprof` 控制，公开 Application Options 没有对应开关，也没有从 app.debug 映射。

考虑到生产安全，建议新增显式 `WithPprof` 并要求独立监听地址或认证保护，而不是隐式绑定 debug。

## 7. 测试与验证结果

### 已执行

| 命令 | 结果 |
| --- | --- |
| `go list ./...` | 成功，包图可解析 |
| `go test ./...` | 失败，被 BUILD-01 阻断 |
| `go vet ./...` | 失败，被 BUILD-01 阻断 |

使用临时 modfile 绕过失效 replace 后，Retry、Outbox、Container 的现有单测及部分 `-race` 测试可以通过。这说明现有测试没有覆盖本报告中的触发场景，不能证明相关实现安全。

### 必须新增的回归测试

- gRPC 认证依赖解析失败必须阻止启动。
- JWT/ServiceAuth 缺密钥、占位密钥、缺失 issuer/audience claim。
- timeout 对永不返回 handler、延迟写响应和客户端取消的行为。
- Provider 乱序注册、缺失依赖、依赖环和 deferred provider。
- 正常关停执行全部 closer；OnStarted 失败完整回滚。
- gRPC 并发 getConn/Close、并发 cache miss、Serve 异常退出。
- Retry 默认策略与资源策略相反时的行为，以及并发 SetPolicy/Do。
- Outbox 并发 claim、重复 Process、失败重试和关停。
- RateLimiter 零值配置和资源级覆盖。
- Health Checker 并发注册、panic、忽略 context。
- Container Make/Destroy/RegisterCloser 交错，并使用 `go test -race` 验证。

## 8. 推荐整改顺序

### 阶段一：恢复安全基线和 CI

1. 修复 BUILD-01，确保干净 checkout 能执行全量测试和 vet。
2. 修复 SEC-01、SEC-02、SEC-03，所有认证错误改为 fail-closed。
3. 将对应安全回归测试加入 CI。

### 阶段二：修复生命周期和资源管理

1. 修复 LIFE-01、LIFE-02、RPC-01、HTTP-01。
2. 修复 Container Destroy 交错和 gRPC Serve 错误传播。
3. 对 lifecycle、container、RPC、timeout 执行 `go test -race`。

### 阶段三：恢复框架核心契约

1. 让 ConfigSource 真正进入默认 Config 主线。
2. 让 DependsOn 参与 Provider 拓扑排序和启动校验。
3. 修复 Retry、Outbox、RateLimiter、Health Checker 语义。

### 阶段四：同步文档与示例

1. 修正 Config API、HTTP 地址、pprof 和 RunContext 文档。
2. 对文档中的 Go 示例增加编译校验。
3. 对 starter 模板执行生成、`go mod tidy`、构建和最小启动测试。

## 9. 发布门槛

建议同时满足以下条件后再发布：

- P0/P1 问题全部修复并有回归测试。
- 干净 checkout 下 `go test ./...`、`go test -race`（核心包）和 `go vet ./...` 通过。
- 认证缺失或初始化失败不会静默降级。
- 正常退出、启动失败和 context 取消三条关停路径均执行完整资源回收。
- Provider DAG 的文档、诊断结果和真实装配顺序一致。
- `hade/documents` 中公开示例能通过自动化编译或契约校验。

## 10. Router 主契约增加路由枚举能力的方案

### 10.1 决策

建议扩展框架，并将路由枚举正式并入 `transport.Router` 主契约，但应作为一次明确的破坏性升级发布。

当前 `framework/contract/transport/router.go:27-65` 只定义路由注册，不提供注册结果的标准读取入口。权限同步、OpenAPI 生成、运行时诊断和治理检查只能依赖 Gin 原生对象或由各业务项目重复实现 Registry 装饰器。

目标状态是让 Contract 层成为路由元数据的唯一标准来源：

- 业务和治理能力不依赖 `gin.Engine.Routes()`。
- Gin、未来的 Fiber/Chi 等 Provider 返回相同的数据模型。
- Group 嵌套路径、Handle 和 Mount 由框架使用统一规则记录。
- 权限同步、文档生成和诊断复用同一份实际注册结果。

由于给 Go interface 增加方法会使所有现有 Provider、Mock 和用户自定义 Router 立即编译失败，该变更不应在 `v0.1.5` 中静默发布。建议目标版本为 **v0.2.0**；如果项目在 v1.0 后执行，则应进入下一个主版本。

### 10.2 主契约设计

在 `framework/contract/transport/router.go` 增加：

```go
// RouteInfo 是框架实际注册路由的只读元数据。
// 不暴露 Handler 和 Middleware 实例，避免 Provider 耦合与敏感实现泄漏。
type RouteInfo struct {
    Method string
    Path   string
    Name   string
}

type Router interface {
    Use(middleware ...Middleware)
    Group(prefix string, middleware ...Middleware) Router

    Handle(method, path string, handler Handler, middleware ...Middleware)
    HandleFunc(method, path string, handlerFunc Handler, middleware ...Middleware)
    GET(path string, handler Handler, middleware ...Middleware)
    POST(path string, handler Handler, middleware ...Middleware)
    PUT(path string, handler Handler, middleware ...Middleware)
    DELETE(path string, handler Handler, middleware ...Middleware)
    Mount(path string, handler http.Handler)

    // Routes 返回当前根 Router 及全部子 Group 已成功注册路由的快照。
    Routes() []RouteInfo
}
```

语义约束：

- `Routes()` 返回副本，调用方修改 slice 或元素不能影响 Router 内部状态。
- 根 Router 和任意子 Group 调用 `Routes()` 都返回同一注册树的完整快照，避免结果随调用入口变化。
- 路由按成功注册顺序返回，保证权限同步和快照测试稳定。
- `Method` 统一为大写标准方法名；`Path` 是包含所有 Group prefix 的绝对路由路径。
- `Name` 是可选稳定标识，未显式命名时为空。第一阶段不能用 handler 函数名自动生成，因为函数名不稳定且可能泄露实现细节。
- 不在 `RouteInfo` 中保存 Handler、Middleware、闭包地址或 Gin handler 名称。
- 每次调用返回当前时刻快照；启动后动态注册是否允许，继续沿用具体 Provider 的限制。

### 10.3 Gin Provider 实现

当前 `framework/provider/gin/router.go:17-25` 的 router 只有 `group` 字段，`Group()` 在 `:51-62` 创建一个全新的 facade。实现主契约时应让整棵路由树共享一个 registry：

```go
type routeRegistry struct {
    mu     sync.RWMutex
    routes []transport.RouteInfo
}

type router struct {
    group    *gin.RouterGroup
    registry *routeRegistry
    prefix   string
}
```

根 Router 初始化一次 registry；Group 创建子 Router 时复用 registry，并由框架计算新 prefix：

```go
func (r *router) Group(prefix string, middleware ...transport.Middleware) transport.Router {
    fullPrefix := joinRoutePath(r.prefix, prefix)
    return &router{
        group:    r.group.Group(prefix, adaptMiddlewares(middleware)...),
        registry: r.registry,
        prefix:   fullPrefix,
    }
}
```

`Handle` 是唯一记录入口。GET/POST/PUT/DELETE/HandleFunc 均继续委托给 Handle，避免重复记录：

```go
func (r *router) Handle(method, path string, handler transport.Handler, middleware ...transport.Middleware) {
    // 参数校验和 Provider 注册逻辑省略。
    r.group.Handle(method, path, handlers...)

    // Gin 注册成功后再记录；若 Gin 因重复路由 panic，不能留下虚假记录。
    r.registry.add(transport.RouteInfo{
        Method: strings.ToUpper(method),
        Path:   joinRoutePath(r.prefix, path),
    })
}
```

`Mount` 当前会注册 GET 和 HEAD，因此应记录两条 RouteInfo。若未来支持 PATCH、OPTIONS、HEAD、ANY 或静态文件 API，也必须最终进入统一 Handle/record 路径。

`joinRoutePath` 不能使用面向文件系统的 `filepath.Join`。它必须保持 HTTP 路径语义，正确处理：

- 空 prefix 和根路径 `/`。
- `/api` + `/users` => `/api/users`。
- 嵌套 Group，例如 `/api` + `/v1` + `/:id`。
- Gin 参数 `:id` 和 catch-all `*path` 原样保留。
- 连续斜杠和尾斜杠采用与 Provider 注册一致的规范化规则，不能让 registry 路径与真实匹配路径不同。

### 10.4 路由命名方案

现有注册方法没有 name 参数，因此加入 `RouteInfo.Name` 后默认只能为空。不要从 handler 名称推导业务权限标识。

建议在同一破坏性版本增加一个独立、显式的命名入口，同时保留现有注册方法：

```go
type RouteOption func(*RouteMetadata)

type RouteMetadata struct {
    Name string
}

type NamedRouteRegistrar interface {
    HandleWithOptions(
        method string,
        path string,
        handler Handler,
        middleware []Middleware,
        options ...RouteOption,
    )
}
```

不过 `NamedRouteRegistrar` 不必立即进入 Router 主契约。第一阶段先保证 Routes()、Method 和 Path 完整可靠；权限系统可以用 `METHOD + Path` 作为稳定键。等命名需求和 API 形态稳定后，再决定是否把命名注册并入主契约，避免一次升级同时扩大两个设计面。

### 10.5 并发与一致性

虽然路由通常在启动阶段注册，Router 仍是公共运行时契约，registry 必须具备基本并发安全：

- `add` 使用写锁，`Routes` 使用读锁并复制 slice。
- RouteInfo 在写入前完成规范化，写入后视为不可变值。
- 不返回内部 slice、map 或指针。
- 不在持有 registry 锁时调用 Gin 或用户 handler，避免锁顺序和重入问题。
- 只记录 Provider 已成功接受的路由；失败或 panic 的注册不能出现在快照中。

第一阶段不建议自动去重。具体 Provider 应继续决定重复路由是 panic、错误还是覆盖；registry 只忠实记录成功注册结果。若后续需要跨 Provider 一致的重复路由策略，应单独设计可返回 error 的注册契约。

### 10.6 破坏面与迁移

已知需要调整的仓库内实现包括：

- `framework/provider/gin/router.go` 的正式 Provider。
- `transport_test.go` 的 `captureRouter`。
- `framework/application/routes_test.go` 的 `testRouter`。
- `framework/bootstrap/http_service_test_helpers.go` 中多个 recording/Gin test Router。
- `framework/http/middleware/health_test.go` 的 `mockRouter`。
- 任何模板、示例或下游项目自行实现的 Router。

迁移步骤：

1. 在 v0.1.x 最后一个版本发布迁移说明，声明 v0.2.0 将给 Router 增加 `Routes()`。
2. 提供可复用的 `transport.RouteRegistry` 或 `transport.NoopRouteSnapshot()` 辅助实现，降低自定义 Provider 和 Mock 的迁移成本。
3. 在框架 Provider、测试替身、starter 模板和示例中实现新方法。
4. 发布 v0.2.0，并在 changelog 中标记 compile-time breaking change。
5. 下游自定义 Router 最低迁移实现可返回空的非 nil slice，但生产 Provider 必须记录真实路由。

不建议通过嵌入临时 `BaseRouter` 强迫用户继承实现。Go 接口的价值是行为契约，框架应提供 helper，但不应把具体继承结构变成隐式要求。

### 10.7 测试方案

主契约升级必须至少覆盖：

- GET/POST/PUT/DELETE/Handle/HandleFunc 每次只记录一条。
- Mount 正确记录 GET 和 HEAD。
- 多层 Group 生成完整路径，参数和 catch-all 不被破坏。
- 从根 Router 或任意子 Group 获取相同完整快照。
- Routes 返回防御性副本，外部修改不会污染内部状态。
- 注册顺序稳定。
- 并发注册和读取通过 `go test -race`。
- Gin 实际 `engine.Routes()` 与 Contract `Routes()` 的 method/path 集合一致，该测试只作为 Provider 一致性验证，不成为生产实现依赖。
- 权限同步消费者只依赖 `transport.Router`，不导入 Gin。
- 自定义最小 Router/Mock 的迁移示例可编译。

建议增加契约级测试套件，所有未来 HTTP Provider 都必须运行：

```go
func RouterContractTests(t *testing.T, factory func() transport.Router)
```

这样新增 Fiber/Chi Provider 时，路由枚举、Group 和 Mount 语义不会各自漂移。

### 10.8 文档与应用层迁移

需要同步更新：

- `hade/documents/reference/api.md`：补充 RouteInfo、Routes 和快照语义。
- `hade/documents/guide/provider.md`：规定所有 HTTP Provider 必须实现路由记录。
- 新增“路由诊断与权限同步”指南，使用 Contract API，不展示 `gin.Engine.Routes()`。
- admin-platform 当前 Registry 装饰器迁移为 `rt.Router.Routes()` 消费者；确认所有所需元数据已进入 RouteInfo 后删除项目级重复记录器。

应用层读取示例：

```go
routes := rt.Router.Routes()
for _, route := range routes {
    permissionKey := route.Method + " " + route.Path
    // 同步权限、生成诊断信息或交给 OpenAPI 工具。
}
```

路由枚举默认不应直接暴露为无认证的生产 HTTP 端点。若提供诊断接口，应受管理面认证和环境开关保护，避免泄露内部管理路径。

### 10.9 验收标准

- 业务和治理代码无需导入 Gin 即可读取完整路由清单。
- Group、Handle、快捷方法和 Mount 的实际路由与快照一致。
- 所有内置 Provider 和 Mock 完成编译期迁移。
- 路由快照通过契约测试和竞态检测。
- 权限同步不再依赖项目级 Registry 装饰器或手写路由清单。
- 文档明确该变更属于 v0.2.0 破坏性升级，并提供下游迁移示例。

## 11. Web 能力边界与扩展优先级

### 11.1 定位与原则

gorp 的目标不是重新实现 Gin，而是提供一条从单体应用平滑演进到微服务治理的路径。HTTP 层继续使用 Gin 作为默认引擎，同时通过 `gorp.Context`、`gorp.Router`、治理中间件、代码生成、OpenAPI、gRPC、服务发现、Tracing、限流、熔断和认证形成稳定的框架契约。

因此，是否把某项 Web 能力提升到 Contract 层，不以“Gin 是否有这个 API”为判断标准，而应看它是否影响：

- 单体 API 的日常开发效率。
- HTTP、OpenAPI 和 gRPC/Proto 契约的一致性。
- 微服务拆分后的调用、观测和错误语义。
- 治理中间件链的稳定挂载。
- 现有 Gin 项目的迁移成本。
- 多 HTTP Provider 之间能否保持稳定语义。

当前框架已经具备治理型 API 框架的主体能力：Gin HTTP engine、Contract/Gin 双模式、全局/分组/接口级中间件、基础治理链、统一响应、参数读取、上传、请求上下文、OpenAPI、Proto-first、Service-first、gRPC 和 WebSocket。`NativeEngine`、`NativeRouterGroup`、`UnwrapContext`、`AdaptMiddleware`、`AdaptHandler` 也提供了必要的 Gin 原生逃生口。

主要缺口集中在 Contract 层：`Router` 和 `Context` 覆盖了基础 API 主线，但部分高频、治理相关能力仍只能下沉到 Gin 或由业务重复实现。

### 11.2 P0：治理型 API 必备能力

这些能力直接影响 API 契约、治理一致性或高频业务开发，应与第 10 节 Router 主契约升级一起进入 v0.2.0。

#### 完整 HTTP 方法

Router 主契约至少补齐：

```go
type Router interface {
    // 现有方法省略。
    PATCH(path string, handler Handler, middleware ...Middleware)
    HEAD(path string, handler Handler, middleware ...Middleware)
    OPTIONS(path string, handler Handler, middleware ...Middleware)
}
```

原因：

- PATCH 已进入 Proto/OpenAPI 生成链，Contract 缺失会造成生成能力与运行时能力割裂。
- HEAD 常用于健康探测、缓存、网关和对象访问。
- OPTIONS 是 CORS、网关预检和 API 基础设施的常见入口。
- 明确方法有利于路由枚举、鉴权、限流、Metrics 和 OpenAPI 生成。

这些快捷方法必须统一委托给 `Handle`，并被第 10 节的 Route Registry 准确记录。

#### 统一 404/405 入口

Router 主契约应提供与 Provider 无关的 fallback 注册能力：

```go
type Router interface {
    // 现有方法省略。
    NoRoute(handler Handler)
    NoMethod(handler Handler)
}
```

要求：

- 框架提供默认 404/405 handler。
- 默认输出使用统一 Responder，与 `httpx.Error` 的错误结构和 request ID 语义一致。
- 业务可以覆盖默认 handler。
- fallback 仍经过适用的全局治理中间件。
- 404 与 405 的业务错误码、HTTP 状态码和日志级别必须稳定。

不建议把 fallback 伪装成普通 RouteInfo；可在未来为路由诊断增加独立的 fallback 元数据，避免权限同步把它当成业务权限点。

#### 请求绑定能力

Context 主契约补齐：

```go
type Context interface {
    // 现有方法省略。
    BindURI(target any) error
    BindHeader(target any) error
    BindForm(target any) error
}
```

契约要求：

- 统一使用 `BindURI` 命名，避免与 Gin 的具体大小写或内部类型绑定。
- 绑定错误返回稳定的框架错误分类，便于统一转换为 400 响应。
- Header 绑定必须保留多值 header 的语义。
- Form 同时明确 `application/x-www-form-urlencoded` 和 multipart 的支持边界。
- 绑定行为应能与 Validator 和 OpenAPI 参数定义衔接，避免 handler 再写一层重复解析。

URI、Header 和 Form 分别承载 REST 参数、租户/认证/Trace/灰度/幂等信息，以及登录、回调、Webhook 等常见业务输入，属于 Contract 主线而不是 Gin 长尾能力。

#### 文件响应与下载

Context 或统一 responder 至少提供：

```go
type Context interface {
    // 现有方法省略。
    File(path string)
    FileAttachment(path, filename string)
}
```

也可以提供更易治理的下载 helper，允许传入 `io.Reader`、内容类型、文件名和长度。需要明确：

- Content-Disposition 文件名编码和 header 注入防护。
- 文件不存在、权限不足和响应已写入时的错误语义。
- 大文件采用流式传输，不能整体读入内存。
- 路径由业务决定时必须提示目录穿越风险。

后台导入导出、报表和模板下载是业务 API 高频能力，应进入 Contract 层。

### 11.3 P1：API 体验增强

这些能力有实际价值，但不阻断单体到微服务治理主线，可在 P0 契约稳定后逐步补齐。

#### Any 路由

```go
ANY(path string, handler Handler, middleware ...Middleware)
```

适用于调试、代理和兼容旧接口。文档必须提醒业务 API 优先使用明确方法，否则权限、Metrics、限流和文档生成的语义会变弱。Route Registry 应展开记录 Provider 实际注册的方法集合，而不是记录虚构的 `ANY` method。

#### 静态文件

建议支持：

```go
Static(relativePath, root string)
StaticFile(relativePath, filePath string)
StaticFS(relativePath string, fs http.FileSystem)
```

适用于 Swagger UI、内部管理页和少量资源托管。大型前端应用仍建议交给构建产物、CDN、Nginx 或网关。静态路由如何进入 Routes 快照需要单独定义，至少记录对外 method/path pattern，不能依赖 Gin 私有表示。

#### SSE 与流式响应

建议提供轻量、Provider 无关的基础能力，而不是在第一阶段引入复杂事件总线：

```go
type StreamWriter interface {
    Write(data []byte) error
    Flush() error
}

type Context interface {
    // 现有方法省略。
    Stream(contentType string, write func(StreamWriter) error) error
    SSE(event SSEEvent) error
}
```

需要统一：

- response header 与 flush 行为。
- 客户端断连和 request context 取消。
- handler 返回后不能继续写入。
- Tracing span、Metrics 和限流如何覆盖长连接生命周期。
- panic、写失败和部分响应后的错误处理。

SSE 适合长任务进度、实时日志、AI 输出和单向事件推送；它继续使用 HTTP，比 WebSocket 更容易复用现有治理链。

#### Cookie helper

可提供 Get/Set/Delete 等轻量 helper，服务管理后台、CSRF、OAuth 回调等场景。不建议把完整 Session 框架作为默认内置能力。默认值应倾向 `HttpOnly`、`Secure`、合理的 SameSite，并让业务显式调整。

#### 内容协商

YAML、Protobuf 和内容协商 helper 可作为后续增强，但优先级低于 JSON、文件下载和 SSE。跨服务强类型通信仍以 gRPC/Proto 为主。

### 11.4 P2：由 Gin 原生模式承接

以下能力不建议完整复制到 gorp Contract：

- HTML 模板加载和服务端页面渲染。
- Gin Context 的全部便利方法。
- 大量 gin-contrib 的逐包包装。
- 复杂 Engine 路径修正、重定向、代理和 multipart 内存细节。
- 第三方 Gin 中间件的完整统一抽象。

只有与治理强相关且需要跨 Provider 稳定语义的中间件适合进入 gorp：CORS、Recovery、Logging、Metrics、Tracing、Security Headers、Body Limit、Rate Limit、Auth、Authorization、Idempotency、Timeout 和 Compression。

长尾能力继续通过 `NativeEngine`、`NativeRouterGroup`、`UnwrapContext` 和 adapter 使用。这样既保留 Gin 生态，又避免 Contract 随 Gin API 膨胀。

### 11.5 原生逃生口的稳定边界

现有 escape hatch 应正式文档化，并承诺清晰边界：

| 场景 | 推荐入口 | 框架责任 |
| --- | --- | --- |
| 常规 JSON API、认证、治理 | gorp Contract | 完整治理和跨 Provider 语义 |
| Provider 特有高级能力 | Native Router/Context | 保留已挂载的上层全局治理，特有行为由业务负责 |
| Gin 第三方中间件 | AdaptMiddleware/Native Group | 适配边界、执行顺序和错误响应需由接入方验证 |
| HTML、复杂代理等长尾能力 | NativeEngine | 不承诺跨 Provider 可移植性 |

需要在文档中明确：

- 下沉到原生 Gin 后哪些全局中间件仍会执行。
- 直接向 NativeEngine 注册的路由是否进入 Contract Routes 快照。建议默认**不进入**，并在诊断中暴露“存在未跟踪原生路由”的差异；若需要统一治理，业务应通过 Contract 注册或显式调用 Registry API。
- 原生 handler 的错误、绑定和响应不自动获得 Contract 的统一语义，除非使用 adapter/helper。
- Provider-specific 代码不具备未来迁移到 Fiber/Chi 的兼容承诺。

### 11.6 与 Router 枚举方案的整合

第 10 节的 Route Registry 是这些 Web 能力的基础设施，而不是孤立功能：

- PATCH/HEAD/OPTIONS 通过 Handle 自动进入快照。
- Any 展开成实际 HTTP 方法后进入快照。
- Static/Mount 记录真实公开 method/path。
- NoRoute/NoMethod 使用独立 fallback 语义，不污染权限路由集合。
- OpenAPI 生成、权限同步、路由诊断共同消费 RouteInfo。
- Contract 路由与 NativeEngine 路由的差异可以用于治理完整性检查。

RouteInfo 第一阶段保持小而稳定：Method、Path、Name。不要为了同时解决 OpenAPI、权限和全部治理元数据而立即加入 handler、middleware 或 Provider 私有字段。后续元数据应通过经过设计的扩展结构演进。

### 11.7 实施顺序

#### v0.2.0 必须完成

1. RouteInfo 与 Router.Routes。
2. PATCH、HEAD、OPTIONS。
3. NoRoute、NoMethod 及统一 404/405 responder。
4. BindURI、BindHeader、BindForm。
5. 文件响应或统一下载 helper。
6. Gin Provider、Mock、代码生成器和文档同步迁移。

#### v0.2.x 后续增强

1. Any。
2. Static、StaticFile、StaticFS。
3. SSE、Stream。
4. Cookie helper。
5. 内容协商 helper。

#### 保持原生模式

1. HTML 模板。
2. 复杂 Gin Engine 配置。
3. 非治理型第三方中间件。
4. Gin Context 长尾方法。

### 11.8 测试与验收

Web 契约扩展除普通单测外，必须建立跨 Provider contract suite：

- 所有 HTTP 方法注册和 Routes 快照一致。
- Group、中间件和 fallback 的执行顺序稳定。
- 404/405 输出结构、HTTP 状态码、错误码和 request ID 一致。
- URI/Header/Form 绑定成功、缺失、格式错误和多值场景一致。
- 文件下载正确设置 headers，并覆盖不存在、非法文件名和大文件流式传输。
- SSE/Stream 覆盖 flush、断连、context cancel、写失败和 handler panic。
- Native 路由与 Contract Registry 差异可被诊断。
- OpenAPI/Proto 生成出的 PATCH 等路由能通过 Contract 注册并实际请求成功。
- 核心测试执行 `go test -race`，确保 Router registry 和流式响应没有并发写问题。

### 11.9 能力准入判断

后续每项 Web 能力进入 Contract 前，应回答：

1. 是否影响 API 契约、OpenAPI 或 Proto 生成一致性？
2. 是否影响认证、限流、Tracing、Metrics 或统一错误等治理链？
3. 是否在单体和微服务业务中都高频使用？
4. 是否能明显减少 handler 重复胶水代码？
5. 是否需要跨 HTTP Provider 保持稳定语义？

多数答案为“是”时，应进入 Contract；否则优先由 Gin 原生模式和 escape hatch 承接。

一句话结论：**gorp 要补齐的是治理型 API 框架的高频稳定契约，而不是复制 Gin；Contract 覆盖单体到微服务演进所需的共同能力，长尾 Web 能力继续留在原生 Provider 层。**

## 12. 整改落地状态（2026-07-21）

本轮已按 review 和第 10、11 节方案完成框架修改。Router 枚举作为 v0.2 主契约直接落地，不再采用 v0.1.x 的可选 `RouteInspector` 过渡方案。

### 已完成

- 构建与测试基线：移除失效的本地 module replace，测试数据库改为内存 SQLite，GORM/SQLX 补齐 SQLite 支持。
- 安全：JWT 与 ServiceAuth Token 缺少密钥或使用模板占位值时 fail-closed；issuer/audience 配置后要求 claims 实际存在；gRPC 服务端和客户端认证依赖解析失败时拒绝启动或建连。
- 生命周期与并发：修复 Lifecycle 回滚、HTTP 服务 Container 销毁、Container Make/Destroy 交错、Provider 状态竞态、gRPC client Close 竞态、DB metrics 重复启动、Outbox 重复投递、Health checker 锁与超时、Retry 并发配置及随机源问题。
- Provider 与配置：Provider 批量注册采用稳定拓扑排序并检测环；远程 ConfigSource 在 bootstrap 后 attach 并 reload；load-shedding Provider 名称与契约键对齐。
- 运行控制：增加 `BootHTTPServiceContext` 和 `RunHTTPContext`，调用方 context 覆盖启动与运行期；HTTP 地址统一回退 `:8080`；pprof 可通过公开选项或 `app.debug=true` 开启。
- Router 主契约：增加 `RouteInfo`、`Routes()`、`PATCH`、`HEAD`、`OPTIONS`、`ANY`、静态资源、`NoRoute` 和 `NoMethod`；Gin Provider 使用父子 Group 共享 registry，并返回防御性快照。
- Context 契约：增加 URI/Header/Form 绑定、文件响应、Cookie、Stream 和 SSE；Gin Provider 与 middleware bridge 已同步实现。
- HTTP 语义：404/405 默认进入统一响应；Timeout middleware 将 deadline 写入 request context，使用线程安全缓冲并在超时后丢弃迟到写入，不再无限等待 handler。
- 文档：本报告已合并 Web 能力边界与 Router 主契约方案；公开 API、HTTP 模式、配置和安全文档已同步实际签名与 fail-closed 行为。

### 验证状态

- 全仓编译 `go test ./... -run '^$'` 通过。
- 框架单元测试及新增定向回归测试通过。
- 核心并发包 `go test -race` 通过，`go vet ./...` 通过。
- 完整 `go test ./...` 中，框架包通过；集成测试仍需要本机启动 Consul、etcd、Kafka、Nacos 和 Redis。
- 明文 gRPC 集成用例已按新安全契约显式设置 `RPCConfig.Insecure=true`，生产配置仍需使用 TLS credentials。

### 后续发布门槛

- 在具备外部依赖的 CI 环境运行完整集成测试。
- 为代码生成模板注入真实 JWT/ServiceAuth 密钥；占位密钥不再是可启动默认值。
- v0.2.0 发布说明必须将 Router/Context 主契约扩展标记为破坏性升级，并给出自定义 Router、Mock 和 Provider 的迁移清单。
