# Runtime 运行时指南

本文档详细说明 gorp 框架的 Runtime 功能，包括 HTTP 服务运行时、ORM 运行时、微服务能力选择器等核心组件。

## 目录

- [概述](#概述)
- [HTTP Service Runtime](#http-service-runtime)
- [ORM Runtime](#orm-runtime)
- [Capability Selector](#capability-selector)
- [Container Runtime Helpers](#container-runtime-helpers)
- [Provider 分组](#provider-分组)
- [配置详解](#配置详解)
- [启动流程](#启动流程)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)

---

## 概述

框架 Runtime 提供以下核心能力：

1. **HTTP Service Runtime** - HTTP 服务启动封装
2. **ORM Runtime** - 统一的数据库运行时抽象
3. **Capability Selector** - 根据配置自动选择微服务能力 Provider
4. **Container Helpers** - 简化容器服务获取的工具函数

---

## HTTP Service Runtime

### HTTPServiceRuntime 结构

```go
type HTTPServiceRuntime struct {
    App         *framework.Application  // 框架应用实例
    Container   contract.Container      // 服务容器
    Logger      contract.Logger         // 日志服务
    Engine      *gin.Engine             // Gin 引擎
    DB          *gorm.DB                // GORM 数据库实例
    Redis       contract.Redis          // Redis 服务
    JWT         contract.JWTService     // JWT 认证服务
    Config      contract.Config         // 配置服务
    ServiceName string                  // 服务名称
}
```

### HTTPServiceOptions 选项

```go
type HTTPServiceOptions struct {
    // ExtraProviders 服务专属 Provider 列表
    // 用于注册服务特有的 Provider（如业务 Provider）
    ExtraProviders []contract.ServiceProvider

    // DisableRedis 禁用 Redis Provider
    // 适用于不需要 Redis 的服务
    DisableRedis bool

    // DisableGorm 禁用 Gorm Provider
    // 适用于不需要数据库的服务
    DisableGorm bool

    // DisableMetrics 禁用 Prometheus 指标采集
    // 适用于不需要监控指标的服务
    DisableMetrics bool
}
```

### 核心函数

#### NewHTTPServiceRuntime

创建 HTTP 服务运行时实例。

```go
func NewHTTPServiceRuntime(serviceName string, opts HTTPServiceOptions) (*HTTPServiceRuntime, error)
```

**参数**：
- `serviceName`: 服务名称，用于日志和健康检查标识
- `opts`: 启动选项

**示例**：

```go
package main

import (
    "github.com/ngq/gorp/framework/bootstrap"
    serviceprovider "your-project/internal/provider"
)

func main() {
    rt, err := bootstrap.NewHTTPServiceRuntime("user-service", bootstrap.HTTPServiceOptions{
        ExtraProviders: []contract.ServiceProvider{
            serviceprovider.NewProvider(),
        },
    })
    if err != nil {
        panic(err)
    }

    // 使用 Runtime 对象
    rt.Logger.Info("服务启动中...")
    rt.Engine.GET("/users", getUsers)
}
```

#### BootHTTPService

完整的 HTTP 服务启动流程（推荐使用）。

```go
func BootHTTPService(
    serviceName string,
    opts HTTPServiceOptions,
    migrate func(*HTTPServiceRuntime) error,
    setup func(*HTTPServiceRuntime) error,
) error
```

**参数**：
- `serviceName`: 服务名称
- `opts`: 启动选项
- `migrate`: 数据库迁移回调（可选）
- `setup`: 服务装配回调（路由注册等）

**完整示例**：

```go
package main

import (
    "github.com/ngq/gorp/framework/bootstrap"
    "your-project/internal/data"
    "your-project/internal/handler"
)

func main() {
    err := bootstrap.BootHTTPService(
        "user-service",
        bootstrap.HTTPServiceOptions{},
        migrate,
        setup,
    )
    if err != nil {
        panic(err)
    }
}

// migrate 数据库迁移
func migrate(rt *bootstrap.HTTPServiceRuntime) error {
    return bootstrap.AutoMigrateModels(rt,
        &data.User{},
        &data.Order{},
        &data.Product{},
    )
}

// setup 服务装配（路由注册）
func setup(rt *bootstrap.HTTPServiceRuntime) error {
    // 注册路由
    userHandler := handler.NewUserHandler(rt.DB, rt.Logger)

    rt.Engine.GET("/users", userHandler.List)
    rt.Engine.GET("/users/:id", userHandler.Get)
    rt.Engine.POST("/users", userHandler.Create)
    rt.Engine.PUT("/users/:id", userHandler.Update)
    rt.Engine.DELETE("/users/:id", userHandler.Delete)

    return nil
}
```

#### AutoMigrateModels

简化 GORM 自动迁移调用。

```go
func AutoMigrateModels(rt *HTTPServiceRuntime, models ...any) error
```

**示例**：

```go
// 单个模型
bootstrap.AutoMigrateModels(rt, &data.User{})

// 多个模型
bootstrap.AutoMigrateModels(rt,
    &data.User{},
    &data.Order{},
    &data.Product{},
)
```

#### RegisterHealthCheck

注册健康检查端点 `/healthz`。

```go
func RegisterHealthCheck(engine *gin.Engine, serviceName string)
```

**返回格式**：

```json
{
    "status": "healthy",
    "service": "user-service",
    "version": "1.0.0"
}
```

#### RegisterMetricsEndpoint

注册 Prometheus 指标端点 `/metrics`。

```go
func RegisterMetricsEndpoint(engine *gin.Engine)
```

**包含指标**：
- Go runtime 指标（内存、GC、goroutine 等）
- HTTP 请求指标（请求数、耗时、状态码）

#### RunHTTP

启动 HTTP 服务（含信号处理和优雅关闭）。

```go
func RunHTTP(c contract.Container, logger contract.Logger) error
```

**特性**：
- 自动注册信号处理（SIGINT、SIGTERM）
- 优雅关闭（10 秒超时）
- 支持 Host 服务集成

---

## ORM Runtime

### ORM Runtime Provider

提供统一的 ORM 能力绑定，位于 `framework/provider/orm/runtime/provider.go`。

**提供的 Key**：

| Key | 类型 | 说明 |
|-----|------|------|
| `ORMBackendKey` | string | ORM 后端类型：gorm/sqlx/ent |
| `DBRuntimeKey` | any | 数据库运行时实例 |
| `MigratorKey` | any | 数据迁移能力 |
| `SQLExecutorKey` | any | SQL 执行能力 |

**Backend 类型**：

```go
type RuntimeBackend string

const (
    RuntimeBackendGorm RuntimeBackend = "gorm"
    RuntimeBackendSQLX RuntimeBackend = "sqlx"
    RuntimeBackendEnt  RuntimeBackend = "ent"
)
```

### 配置切换 Backend

```yaml
# config.yaml
database:
  backend: gorm  # 可选: gorm, sqlx, ent
```

### DBRuntime 行为

```go
// Provider.Register 中的绑定逻辑
c.Bind(contract.DBRuntimeKey, func(c contract.Container) (any, error) {
    backend := c.Make(contract.ORMBackendKey)
    switch backend {
    case "sqlx":
        return c.Make(contract.SQLXKey)  // 返回 *sqlx.DB
    case "ent":
        return c.Make(contract.EntClientKey)  // 返回 Ent Client
    default:
        return c.Make(contract.GormKey)  // 返回 *gorm.DB
    }
}, true)
```

### 使用示例

```go
package main

import (
    "github.com/ngq/gorp/framework/bootstrap"
    "github.com/ngq/gorp/framework/container"
    "github.com/ngq/gorp/framework/contract"
    ormruntime "github.com/ngq/gorp/framework/provider/orm/runtime"
)

func main() {
    // 注册 ORM Runtime Provider
    app := framework.NewApplication()
    c := app.Container()

    // 注册 Provider
    c.RegisterProviders(
        bootstrap.ORMRuntimeProviders(),  // 包含 orm.runtime
    )

    // 获取统一数据库运行时
    dbRuntime, err := container.MakeDBRuntime(c)
    if err != nil {
        panic(err)
    }

    // 健康检查
    if err := container.PingDBRuntime(dbRuntime); err != nil {
        panic(err)
    }
}
```

---

## Capability Selector

Capability Selector 根据配置文件自动选择合适的微服务能力 Provider，避免手动注册和冲突。

### RegisterSelectedMicroserviceProviders

自动注册微服务能力 Provider。

```go
func RegisterSelectedMicroserviceProviders(c contract.Container) error
```

**流程**：
1. 检查配置是否已绑定
2. 选择并注册 ConfigSource Provider
3. 如果是远程配置源，重新加载配置
4. 根据配置选择并注册其他 Provider

**示例**：

```go
package main

import (
    "github.com/ngq/gorp/framework/bootstrap"
    "github.com/ngq/gorp/framework/container"
)

func main() {
    app := framework.NewApplication()
    c := app.Container()

    // 1. 注册基础 Provider（包含 config）
    c.RegisterProviders(bootstrap.DefaultProviders()...)

    // 2. 自动注册微服务能力（根据配置选择）
    if err := bootstrap.RegisterSelectedMicroserviceProviders(c); err != nil {
        panic(err)
    }

    // 服务已具备配置指定的微服务能力
}
```

### 各能力选择器

#### ConfigSource Provider

```go
func SelectConfigSourceProvider(cfg contract.Config) contract.ServiceProvider
```

| 配置值 | Provider |
|-------|----------|
| `consul` | configsourceconsul |
| `etcd` | configsourceetcd |
| `apollo` | configsourceapollo |
| `nacos` | configsourcenacos |
| `kubernetes` | configsourcekubernetes |
| `polaris` | configsourcepolaris |
| `noop` | configsourcenoop |
| `local`（默认） | configsourcelocal |

#### Discovery Provider

```go
func SelectDiscoveryProvider(cfg contract.Config) contract.ServiceProvider
```

| 配置值 | Provider |
|-------|----------|
| `consul` | discoveryconsul |
| `etcd` | discoveryetcd |
| `nacos` | discoverynacos |
| `zookeeper` | discoveryzookeeper |
| `kubernetes` | discoverykubernetes |
| `polaris` | discoverypolaris |
| `eureka` | discoveryeureka |
| `servicecomb` | discoveryservicecomb |
| `noop`（默认） | discoverynoop |

#### Selector Provider（负载均衡）

```go
func SelectSelectorProvider(cfg contract.Config) contract.ServiceProvider
```

| 配置值 | Provider | 算法 |
|-------|----------|------|
| `random` | selectorrandom | 随机选择 |
| `wrr` | selectorwrr | 加权轮询 |
| `p2c` | selectorp2c | Power of Two Choices |
| `noop`（默认） | selectornoop | 无选择 |

#### RPC Provider

```go
func SelectRPCProvider(cfg contract.Config) contract.ServiceProvider
```

| 配置值 | Provider |
|-------|----------|
| `http` | rpchttp |
| `grpc` | rpcgrpc |
| `noop`（默认） | rpcnoop |

#### Tracing Provider

```go
func SelectTracingProvider(cfg contract.Config) contract.ServiceProvider
```

| 配置值 | Provider |
|-------|----------|
| `otel`, `otlp`, `grpc`, `http`, `stdout` | tracingotel |
| `noop`（默认） | tracingnoop |

#### Metadata Provider

```go
func SelectMetadataProvider(cfg contract.Config) contract.ServiceProvider
```

| 配置值 | Provider |
|-------|----------|
| `default` 或 `enabled=true` | metadatadefault |
| `noop`（默认） | metadatanoop |

#### ServiceAuth Provider

```go
func SelectServiceAuthProvider(cfg contract.Config) contract.ServiceProvider
```

| 配置值 | Provider |
|-------|----------|
| `token` 或 `enabled=true` | serviceauthtoken |
| `mtls` | serviceauthmtls |
| `noop`（默认） | serviceauthnoop |

#### CircuitBreaker Provider

```go
func SelectCircuitBreakerProvider(cfg contract.Config) contract.ServiceProvider
```

| 配置值 | Provider |
|-------|----------|
| `sentinel` 或 `enabled=true` | circuitbreakersentinel |
| `noop`（默认） | circuitbreakernoop |

#### DTM Provider（分布式事务）

```go
func SelectDTMProvider(cfg contract.Config) contract.ServiceProvider
```

| 配置值 | Provider |
|-------|----------|
| `sdk`, `dtmsdk` 或 `enabled=true` | dtmsdk |
| `noop`（默认） | dtmnoop |

#### MessageQueue Provider

```go
func SelectMessageQueueProvider(cfg contract.Config) contract.ServiceProvider
```

| 配置值 | Provider |
|-------|----------|
| `redis` 或 `enabled=true` | mqredis |
| `noop`（默认） | mqnoop |

#### DistributedLock Provider

```go
func SelectDistributedLockProvider(cfg contract.Config) contract.ServiceProvider
```

| 配置值 | Provider |
|-------|----------|
| `redis` 或 `enabled=true` | dlockredis |
| `noop`（默认） | dlocknoop |

---

## Container Runtime Helpers

### 安全获取函数（返回 error）

```go
// 数据库运行时
func MakeDBRuntime(c contract.Container) (any, error)

// Redis 服务
func MakeRedis(c contract.Container) (contract.Redis, error)

// 统一缓存服务
func MakeCache(c contract.Container) (contract.Cache, error)

// GORM 实例
func MakeGormDB(c contract.Container) (*gorm.DB, error)

// SQLX 实例
func MakeSQLX(c contract.Container) (*sqlx.DB, error)

// 消息发布者
func MakeMessagePublisher(c contract.Container) (contract.MessagePublisher, error)

// 消息订阅者
func MakeMessageSubscriber(c contract.Container) (contract.MessageSubscriber, error)

// 分布式锁
func MakeDistributedLock(c contract.Container) (contract.DistributedLock, error)

// Proto-first gRPC 连接工厂
func MakeGRPCConnFactory(c contract.Container) (contract.GRPCConnFactory, error)

// Proto-first gRPC 服务端注册器
func MakeGRPCServerRegistrar(c contract.Container) (contract.GRPCServerRegistrar, error)

// Cron 服务
func MakeCron(c contract.Container) (contract.Cron, error)

// 日志服务
func MakeLogger(c contract.Container) (contract.Logger, error)

// Host 服务
func MakeHost(c contract.Container) (contract.Host, error)

// HTTP 服务
func MakeHTTP(c contract.Container) (contract.HTTP, error)

// Gin Engine
func MakeGinEngine(c contract.Container) (*gin.Engine, error)
```

### 强制获取函数（panic）

```go
// 日志服务
func MustMakeLogger(c contract.Container) contract.Logger

// GORM 实例
func MustMakeGorm(c contract.Container) *gorm.DB

// Gin Engine
func MustMakeEngine(c contract.Container) *gin.Engine

// 消息发布者
func MustMakeMessagePublisher(c contract.Container) contract.MessagePublisher

// 消息订阅者
func MustMakeMessageSubscriber(c contract.Container) contract.MessageSubscriber

// 分布式锁
func MustMakeDistributedLock(c contract.Container) contract.DistributedLock

// Proto-first gRPC 连接工厂
func MustMakeGRPCConnFactory(c contract.Container) contract.GRPCConnFactory

// Proto-first gRPC 服务端注册器
func MustMakeGRPCServerRegistrar(c contract.Container) contract.GRPCServerRegistrar

// 配置服务
func MustMakeConfig(c contract.Container) contract.Config

// 缓存服务
func MustMakeCache(c contract.Container) contract.Cache
```

### 工具函数

```go
// 数据库健康检查
func PingDBRuntime(dbAny any) error
```

支持类型：
- `*gorm.DB`
- `*sql.DB`
- 实现了 `Ping() error` 接口的类型

### 使用示例

```go
package main

import (
    "github.com/ngq/gorp/framework/bootstrap"
    "github.com/ngq/gorp/framework/container"
)

func setup(rt *bootstrap.HTTPServiceRuntime) error {
    c := rt.Container

    // 强制获取（启动阶段必需）
    logger := container.MustMakeLogger(c)
    db := container.MustMakeGorm(c)

    // 安全获取（可选能力）
    redis, err := container.MakeRedis(c)
    if err != nil {
        logger.Info("Redis not configured")
    }

    cache, err := container.MakeCache(c)
    if err != nil {
        logger.Info("Cache not configured")
    }

    // 数据库健康检查
    if err := container.PingDBRuntime(db); err != nil {
        return err
    }

    return nil
}
```

---

## Provider 分组

### FoundationProviders

基础框架能力，所有服务都需要。

```go
func FoundationProviders() []contract.ServiceProvider {
    return []contract.ServiceProvider{
        appProvider.NewProvider(),      // 应用配置
        configProvider.NewProvider(),   // 配置服务
        logProvider.NewProvider(),      // 日志服务
        ginProvider.NewProvider(),      // HTTP 引擎
        hostProvider.NewProvider(),     // 进程宿主
        cronProvider.NewProvider(),     // 定时任务
    }
}
```

### ORMRuntimeProviders

数据库和 ORM 能力。

```go
func ORMRuntimeProviders() []contract.ServiceProvider {
    return []contract.ServiceProvider{
        gormProvider.NewProvider(),        // GORM
        sqlxProvider.NewProvider(),        // SQLX
        runtimeORMProvider.NewProvider(),  // ORM Runtime
        inspectProvider.NewProvider(),     // 数据库检查
    }
}
```

### AuthProviders

业务认证能力。

```go
func AuthProviders() []contract.ServiceProvider {
    return []contract.ServiceProvider{
        authJWTProvider.NewProvider(),  // 业务 JWT
    }
}
```

### BusinessSimplificationProviders

业务开发简化能力。

```go
func BusinessSimplificationProviders() []contract.ServiceProvider {
    return []contract.ServiceProvider{
        authJWTProvider.NewProvider(),  // 业务 JWT
        cacheProvider.NewProvider(),    // 统一缓存
    }
}
```

### DefaultProviders

完整的默认 Provider 组。

```go
func DefaultProviders() []contract.ServiceProvider {
    providers := make([]contract.ServiceProvider, 0, 24)
    providers = append(providers, FoundationProviders()...)
    providers = append(providers, ORMRuntimeProviders()...)
    providers = append(providers, BusinessSimplificationProviders()...)
    return providers
}
```

---

## 配置详解

### 完整配置示例

```yaml
# config.yaml

# 服务基本信息
service:
  name: user-service
  version: 1.0.0

# 数据库配置
database:
  backend: gorm  # 可选: gorm, sqlx, ent
  driver: mysql
  dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True"

# Redis 配置
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

# HTTP 配置
http:
  port: 8080
  mode: release  # debug, release, test

# 配置源（远程配置）
configsource:
  backend: local  # consul, etcd, apollo, nacos, kubernetes, polaris, noop
  # consul:
  #   address: 127.0.0.1:8500
  #   path: config/user-service

# 服务发现
discovery:
  backend: noop  # consul, etcd, nacos, zookeeper, kubernetes, polaris, eureka
  # consul:
  #   address: 127.0.0.1:8500

# 负载均衡选择器
selector:
  algorithm: noop  # random, wrr, p2c

# RPC 配置
rpc:
  mode: noop  # http, grpc
  # grpc:
  #   address: :9090
  #   insecure: true
  #   timeout_ms: 30000

# 链路追踪
tracing:
  enabled: false
  backend: noop  # otel, otlp, grpc, http, stdout
  # otel:
  #   endpoint: localhost:4317
  #   service_name: user-service

# Metadata 传播
metadata:
  enabled: false
  mode: noop  # default
  propagate_prefix: "x-"

# 服务间认证
service_auth:
  enabled: false
  backend: noop  # token, mtls
  # token:
  #   secret: your-secret-key

# 熔断器
circuit_breaker:
  enabled: false
  backend: noop  # sentinel
  # sentinel:
  #   rules:
  #     - name: user-service
  #       threshold: 0.5

# 分布式事务（DTM）
dtm:
  enabled: false
  backend: noop  # dtmsdk
  # dtmsdk:
  #   address: localhost:36789

# 消息队列
message_queue:
  enabled: false
  backend: noop  # redis
  # redis:
  #   channel_prefix: "mq:"

# 分布式锁
distributed_lock:
  enabled: false
  backend: noop  # redis
  # redis:
  #   key_prefix: "lock:"
```

### 配置优先级

当多个配置 key 可用时，按顺序优先：

```go
// getConfigString 支持多个候选 key
func getConfigString(cfg contract.Config, keys ...string) string

// 示例：discovery.backend 或 discovery.type 都可以
SelectDiscoveryProvider(cfg)  // 检查 discovery.backend, discovery.type
```

---

## 启动流程

### HTTP 服务启动流程

```
┌─────────────────────────────────────────────────────────────┐
│                     BootHTTPService                          │
├─────────────────────────────────────────────────────────────┤
│  1. NewHTTPServiceRuntime                                    │
│     ├── framework.NewApplication()                          │
│     ├── buildHTTPProviders(opts)                            │
│     │   ├── FoundationProviders()                           │
│     │   ├── ORMRuntimeProviders() (if !DisableGorm)         │
│     │   ├── BusinessSimplificationProviders()               │
│     │   ├── redisProvider (if !DisableRedis)                │
│     │   └── ExtraProviders                                  │
│     ├── RegisterSelectedMicroserviceProviders(c)            │
│     │   ├── SelectConfigSourceProvider → 注册并 Reload      │
│     │   ├── SelectDiscoveryProvider                         │
│     │   ├── SelectSelectorProvider                          │
│     │   ├── SelectRPCProvider                               │
│     │   ├── SelectTracingProvider                           │
│     │   ├── SelectMetadataProvider                          │
│     │   ├── SelectServiceAuthProvider                       │
│     │   ├── SelectCircuitBreakerProvider                    │
│     │   ├── SelectDTMProvider                               │
│     │   ├── SelectMessageQueueProvider                      │
│     │   └── SelectDistributedLockProvider                   │
│     └── 构建 HTTPServiceRuntime                             │
│         ├── Logger = MustMakeLogger                         │
│         ├── Engine = MustMakeEngine                         │
│         ├── Config = MustMakeConfig                         │
│         ├── JWT = MustMakeJWTService                        │
│         ├── DB = MustMakeGorm (if !DisableGorm)             │
│         └── Redis = MakeRedis (if !DisableRedis)            │
├─────────────────────────────────────────────────────────────┤
│  2. 执行 migrate(rt)                                         │
│     └── AutoMigrateModels 或自定义迁移逻辑                   │
├─────────────────────────────────────────────────────────────┤
│  3. 执行 setup(rt)                                           │
│     └── 注册路由、中间件等                                   │
├─────────────────────────────────────────────────────────────┤
│  4. RegisterHealthCheck(rt.Engine, serviceName)             │
│     └── GET /healthz → {status, service, version}           │
├─────────────────────────────────────────────────────────────┤
│  5. RegisterMetricsEndpoint(rt.Engine)                      │
│     └── GET /metrics → Prometheus 指标                      │
│     └── MetricsMiddleware() → 收集 HTTP 指标                │
├─────────────────────────────────────────────────────────────┤
│  6. RunHTTP(c, logger)                                       │
│     ├── 获取 Host 和 HTTP 服务                               │
│     ├── Register HTTP to Host                                │
│     ├── Host.Start(ctx)                                      │
│     ├── signal.NotifyContext (SIGINT, SIGTERM)              │
│     ├── <-ctx.Done() 等待信号                                │
│     └── Host.Shutdown(ctx, 10s timeout)                     │
└─────────────────────────────────────────────────────────────┘
```

### 微服务能力注册流程

```
┌─────────────────────────────────────────────────────────────┐
│         RegisterSelectedMicroserviceProviders                │
├─────────────────────────────────────────────────────────────┤
│  1. 检查 Config 是否已绑定                                   │
│     └── if !IsBind(ConfigKey) → return nil                  │
├─────────────────────────────────────────────────────────────┤
│  2. 选择并注册 ConfigSource Provider                         │
│     ├── SelectConfigSourceProvider(cfg)                     │
│     ├── 如果是远程配置源（非 local/noop）                    │
│     │   ├── RegisterProvider                                │
│     │   └── cfg.Reload(ctx) 重新加载配置                    │
│     └── cfg = 重新加载后的配置                               │
├─────────────────────────────────────────────────────────────┤
│  3. 依次注册其他 Provider                                    │
│     ├── SelectDiscoveryProvider                             │
│     ├── SelectSelectorProvider                              │
│     ├── SelectRPCProvider                                   │
│     ├── SelectTracingProvider                               │
│     ├── SelectMetadataProvider                              │
│     ├── SelectServiceAuthProvider                           │
│     ├── SelectCircuitBreakerProvider                        │
│     ├── SelectDTMProvider                                   │
│     ├── SelectMessageQueueProvider                          │
│     └── SelectDistributedLockProvider                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 最佳实践

### 1. 最简服务启动

```go
package main

import (
    "github.com/ngq/gorp/framework/bootstrap"
)

func main() {
    // 最简启动：无数据库、无 Redis
    err := bootstrap.BootHTTPService(
        "simple-service",
        bootstrap.HTTPServiceOptions{
            DisableGorm:  true,
            DisableRedis: true,
        },
        nil,  // 无迁移
        setup,
    )
    if err != nil {
        panic(err)
    }
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
    rt.Engine.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })
    return nil
}
```

### 2. 标准服务启动

```go
package main

import (
    "github.com/ngq/gorp/framework/bootstrap"
    "github.com/ngq/gorp/framework/container"
    serviceprovider "your-project/internal/provider"
    "your-project/internal/data"
    "your-project/internal/handler"
)

func main() {
    err := bootstrap.BootHTTPService(
        "user-service",
        bootstrap.HTTPServiceOptions{
            ExtraProviders: []contract.ServiceProvider{
                serviceprovider.NewProvider(),
            },
        },
        migrate,
        setup,
    )
    if err != nil {
        panic(err)
    }
}

func migrate(rt *bootstrap.HTTPServiceRuntime) error {
    return bootstrap.AutoMigrateModels(rt,
        &data.User{},
        &data.UserProfile{},
    )
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
    // 注册中间件
    rt.Engine.Use(authMiddleware(rt.JWT))

    // 注册路由
    h := handler.NewUserHandler(rt.DB, rt.Logger)
    api := rt.Engine.Group("/api/v1")
    {
        api.GET("/users", h.List)
        api.GET("/users/:id", h.Get)
        api.POST("/users", h.Create)
    }

    return nil
}
```

### 3. 使用可选能力

```go
package main

import (
    "github.com/ngq/gorp/framework/bootstrap"
    "github.com/ngq/gorp/framework/container"
)

func setup(rt *bootstrap.HTTPServiceRuntime) error {
    c := rt.Container
    logger := rt.Logger

    // 可选：Redis
    redis, err := container.MakeRedis(c)
    if err != nil {
        logger.Info("Redis not configured, using in-memory cache")
    } else {
        // 使用 Redis
        rt.Engine.Use(rateLimitMiddleware(redis))
    }

    // 可选：缓存
    cache, err := container.MakeCache(c)
    if err != nil {
        logger.Info("Cache not configured")
    } else {
        // 使用缓存
        rt.Engine.Use(cacheMiddleware(cache))
    }

    // 可选：分布式锁
    lock, err := container.MakeDistributedLock(c)
    if err != nil {
        logger.Info("Distributed lock not configured")
    } else {
        // 使用分布式锁
        rt.Engine.Use(idempotencyMiddleware(lock))
    }

    return nil
}
```

### 4. 微服务模式启动

```yaml
# config.yaml
discovery:
  backend: consul

selector:
  algorithm: wrr

rpc:
  mode: grpc

tracing:
  enabled: true
  backend: otel

metadata:
  enabled: true

service_auth:
  enabled: true
  backend: token

circuit_breaker:
  enabled: true
```

```go
package main

import (
    "github.com/ngq/gorp/framework/bootstrap"
    "github.com/ngq/gorp/framework/container"
)

func setup(rt *bootstrap.HTTPServiceRuntime) error {
    c := rt.Container

    // 自动具备微服务能力
    // - 服务发现（Consul）
    // - 负载均衡（WRR）
    // - RPC（gRPC）
    // - 链路追踪（OpenTelemetry）
    // - Metadata 传播
    // - 服务间认证（Token）
    // - 熔断器（Sentinel）

    // 获取 gRPC 连接工厂
    connFactory, err := container.MakeGRPCConnFactory(c)
    if err != nil {
        return err
    }

    // 获取 gRPC Server 注册器
    registrar, err := container.MakeGRPCServerRegistrar(c)
    if err != nil {
        return err
    }

    // 注册 gRPC 服务
    registrar.RegisterProto(func(s *grpc.Server) error {
        pb.RegisterUserServiceServer(s, &userServiceServer{})
        return nil
    })

    return nil
}
```

### 5. Cron Worker 启动

```go
package main

import (
    "context"

    "github.com/ngq/gorp/framework/bootstrap"
    "github.com/ngq/gorp/framework/container"
)

func main() {
    // 创建 Application
    app, c, err := bootstrap.NewDefaultApplication()
    if err != nil {
        panic(err)
    }

    // 获取 Cron 服务
    cron, err := container.MakeCron(c)
    if err != nil {
        panic(err)
    }

    // 注册定时任务
    cron.AddFunc("0 */5 * * *", func() {
        // 每 5 分钟执行
        cleanupExpiredSessions(c)
    })

    cron.AddFunc("0 0 * * *", func() {
        // 每天执行
        generateDailyReport(c)
    })

    // 启动
    cron.Start()
    app.Run()
}

func cleanupExpiredSessions(c contract.Container) {
    db := container.MustMakeGorm(c)
    logger := container.MustMakeLogger(c)

    result := db.Where("expires_at < ?", time.Now()).Delete(&Session{})
    logger.Info("清理过期 Session", map[string]any{"count": result.RowsAffected})
}
```

---

## 常见问题

### Q1: 如何禁用某个默认 Provider？

```go
// 方式1：使用 Options
bootstrap.HTTPServiceOptions{
    DisableGorm:    true,  // 禁用 Gorm
    DisableRedis:   true,  // 禁用 Redis
    DisableMetrics: true,  // 禁用 Prometheus
}

// 方式2：手动构建 Provider 列表
providers := []contract.ServiceProvider{}
providers = append(providers, bootstrap.FoundationProviders()...)
// 不添加 ORMRuntimeProviders()
providers = append(providers, bootstrap.BusinessSimplificationProviders()...)
providers = append(providers, opts.ExtraProviders...)
```

### Q2: 如何覆盖默认 Provider？

```go
// 方式：在 ExtraProviders 中添加同名 Provider
// 后注册的会覆盖先注册的（如果支持覆盖）
bootstrap.HTTPServiceOptions{
    ExtraProviders: []contract.ServiceProvider{
        myCustomLogProvider.NewProvider(),  // 覆盖默认 log Provider
    },
}
```

### Q3: 如何获取服务实例地址？

```go
// 在服务启动后，HTTP 地址由 Host 管理
host, err := container.MakeHost(c)
if err != nil {
    return err
}

// 获取 HTTP 服务地址
httpHostable := host.GetService("http")
addr := httpHostable.Addr()
```

### Q4: 如何自定义优雅关闭超时？

```go
// 当前默认 10 秒超时
// 如需自定义，使用 Host.Shutdown
host, _ := container.MakeHost(c)

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

host.Shutdown(ctx)
```

### Q5: 多数据库如何配置？

```go
// ORM Runtime 只支持单一 Backend
// 如需多数据库，直接使用 GORM 创建多个连接

// 主数据库（ORM Runtime 管理）
mainDB := container.MustMakeGorm(c)

// 第二数据库（手动创建）
secondDB, err := gorm.Open(mysql.Open(secondDSN), &gorm.Config{})
if err != nil {
    panic(err)
}
```

### Q6: Capability Selector 何时执行？

```
注册时机：
├── NewHTTPServiceRuntime 内部调用
├── 在 DefaultProviders 注册后执行
├── Config Provider 已绑定，配置已加载
└── 自动根据配置选择微服务 Provider
```

### Q7: 如何查看当前启用的能力？

```go
func debugCapabilities(c contract.Container) {
    keys := []string{
        contract.DiscoveryKey,
        contract.SelectorKey,
        contract.RPCClientKey,
        contract.TracerKey,
        contract.MetadataPropagatorKey,
        contract.ServiceAuthKey,
        contract.CircuitBreakerKey,
        contract.DistributedLockKey,
        contract.MessagePublisherKey,
    }

    logger := container.MustMakeLogger(c)

    for _, key := range keys {
        if c.IsBind(key) {
            logger.Info("能力已启用", map[string]any{"key": key})
        } else {
            logger.Info("能力未启用", map[string]any{"key": key})
        }
    }
}
```

---

## 总结

### Runtime 功能矩阵

| 功能 | 入口函数 | 适用场景 |
|------|---------|---------|
| HTTP 服务启动 | `BootHTTPService` | 所有 HTTP 服务 |
| HTTP 运行时创建 | `NewHTTPServiceRuntime` | 需要自定义启动流程 |
| ORM Runtime | `ORMRuntimeProviders` | 数据库服务 |
| 微服务能力 | `RegisterSelectedMicroserviceProviders` | 微服务架构 |
| 服务获取 | `Make*/MustMake*` | 业务代码中获取服务 |

### 启动方式选择

| 场景 | 推荐方式 |
|------|---------|
| 单体应用 | `BootHTTPService` + 禁用微服务能力 |
| 微服务应用 | `BootHTTPService` + 配置微服务能力 |
| Cron Worker | `NewDefaultApplication` + `MakeCron` |
| 纯 gRPC 服务 | 自定义启动 + `MakeGRPCServerRegistrar` |
| 最简服务 | `BootHTTPService` + 禁用 DB/Redis |