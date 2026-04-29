# nop-go 微服务能力使用指南

本项目使用 **gorp 框架** 提供的微服务能力，遵循"单体→微服务渐进演进"的设计原则。

## 框架当前提供的能力位

| 能力位 | 单体阶段默认实现 | 微服务阶段可接入后端 | 默认业务入口 |
|-----|----------------|---------------------|---------|
| **服务发现** | `discovery/noop` | consul, etcd, nacos, k8s 等 | 启动装配层注入，业务默认不直接从容器取 |
| **负载均衡** | `selector/noop` | random, wrr, p2c | 启动装配层注入，业务默认不直接从容器取 |
| **链路追踪** | `tracing/noop` | otel | framework transport / middleware 自动接入 |
| **熔断器** | `circuitbreaker/noop` | sentinel | 按业务边界注入或封装在 service/usecase |
| **限流器** | `circuitbreaker/noop` | sentinel | 优先通过 middleware 接入 |
| **RPC** | `rpc/noop` | grpc / http | Proto-first + `GRPCConnFactory` |
| **分布式锁** | `dlock/noop` | redis, etcd | typed runtime / `container.MakeDistributedLock` 轻入口 |
| **消息队列** | `messagequeue/noop` | redis / kafka / rabbitmq 等 | typed runtime / `container.MakeMessagePublisher` 轻入口 |
| **HTTP Metrics** | 内置 | prometheus | `gin.MetricsMiddleware()` |

## 单体阶段（默认）

单体阶段的主线目标是：保持默认入口更轻，优先通过 typed runtime 与 capability helper 起步；微服务能力默认走 noop，实现零依赖启动。

下面这段代码用于说明框架内部如何装配 noop 能力，不是推荐业务项目在默认起步阶段直接照抄的首屏写法。

```go
package main

import (
    "github.com/ngq/gorp/framework/bootstrap"
    "github.com/ngq/gorp/framework/contract"
)

func main() {
    // 单体模式：使用 noop 实现
    app, c, err := bootstrap.Init(bootstrap.Options{
        ExtraProviders: bootstrap.MonolithFriendlyProviders(),
    })
    // 所有请求直接通过，无熔断/限流/追踪
}
```

## 微服务阶段

切换到真实实现时，业务默认不需要自己在主线代码里手工拼一批具体后端 Provider；更推荐通过配置切换能力位，再由 bootstrap + capability selector 选中对应实现。

```go
package main

import (
    "github.com/ngq/gorp/framework/bootstrap"
)

func main() {
    // 微服务模式：保持默认启动骨架，由配置驱动 capability selector 选中真实实现
    app, c, err := bootstrap.NewDefaultApplication()
    _ = app
    _ = c
    _ = err
}
```

对应思路是：

- 基础骨架仍走 `DefaultProviders()`
- 微服务真实能力由 `RegisterSelectedMicroserviceProviders(...)` 按配置选中
- 业务默认主线优先看到的是 capability helper / typed runtime，而不是一长串具体后端 import

## 使用示例

下面这些示例用于说明 capability 在业务边界层如何被使用；它们表达的是 contract / helper 心智，不要求业务项目自己在启动入口里手工拼装具体后端。

### 1. HTTP Metrics（已自动集成）

```go
// framework/provider/gin/metrics.go 已提供
// 在服务启动时自动注册，无需手动配置
// 访问 /metrics 端点即可获取 Prometheus 指标
```

### 2. 熔断器保护下游调用

```go
func callDownstream(cb contract.CircuitBreaker, url string) error {
    return cb.Do(context.Background(), "downstream-service", func() error {
        resp, err := http.Get(url)
        _ = resp
        return err
    })
}
```

### 3. 限流保护 API

```go
func apiHandler(rl contract.RateLimiter) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        if err := rl.Allow(ctx, "api:/v1/users"); err != nil {
            ctx.JSON(429, gin.H{"error": "rate limited"})
            ctx.Abort()
            return
        }
        // 处理请求
    }
}
```

### 4. 服务发现 + 负载均衡

```go
func callService(disc contract.Discovery, sel contract.Selector, serviceName string) error {
    instances, err := disc.Discover(context.Background(), serviceName)
    if err != nil {
        return err
    }

    instance, err := sel.Select(instances)
    if err != nil {
        return err
    }

    url := fmt.Sprintf("http://%s:%d/api", instance.Address, instance.Port)
    _ = url
    return nil
}
```

### 5. 链路追踪

```go
func tracedOperation(tracer contract.Tracer) error {
    ctx, span := tracer.Start(context.Background(), "operation-name")
    defer span.End()

    span.SetAttributes("key", "value")
    _ = ctx
    return nil
}
```

## 配置参考

### config/app.yaml（微服务模式）

```yaml
# 服务发现配置
discovery:
  driver: consul
  consul:
    address: consul:8500

# 链路追踪配置
tracing:
  driver: otel
  otel:
    endpoint: jaeger:4317
    sampler_rate: 0.1

# 熔断器配置
circuit_breaker:
  enabled: true
  strategy: sentinel
  default:
    threshold: 0.5        # 错误率阈值 50%
    min_request_count: 10 # 最小请求数
    timeout: 10s          # 熔断超时

# 限流配置  
rate_limiter:
  enabled: true
  default:
    qps: 1000
    burst: 100
```

## 监控栈

使用 `docker-compose.monitoring.yml` 启动监控基础设施：

```bash
# 启动微服务 + 监控栈
docker-compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d

# 访问地址
# Grafana:     http://localhost:3000 (admin/nop123456)
# Prometheus:  http://localhost:9090
# Jaeger:      http://localhost:16686
# Consul:      http://localhost:8500
```

## 架构演进路径

```
单体阶段                    微服务阶段
┌─────────────┐            ┌─────────────┐
│ App         │            │ App         │
│ ├─ Gin      │            │ ├─ Gin      │
│ ├─ Gorm     │            │ ├─ Gorm     │
│ └─ Redis    │            │ └─ Redis    │
│             │            │             │
│ noop 实现值：│  ──────▶   │ 真实实现值： │
│ ├─ Discovery│            │ ├─ Consul   │
│ ├─ Tracing  │            │ ├─ Jaeger   │
│ ├─ CircuitBreaker│       │ ├─ Sentinel │
│ └─ Selector │            │ └─ P2C      │
└─────────────┘            └─────────────┘
    零依赖                    完整能力
```

通过切换 Provider，无需修改业务代码即可从单体演进到微服务！