# nop-go 微服务能力使用指南

本项目使用 **gorp 框架** 提供的微服务能力，遵循"单体→微服务渐进演进"的设计原则。

## 框架已提供的能力

| 能力 | noop (单体阶段) | 真实实现 (微服务阶段) | 使用方式 |
|-----|----------------|---------------------|---------|
| **服务发现** | `discovery/noop` | consul, etcd, nacos, k8s | `container.MustMake[contract.Discovery](c)` |
| **负载均衡** | `selector/noop` | random, wrr, p2c | `container.MustMake[contract.Selector](c)` |
| **链路追踪** | `tracing/noop` | otel (OpenTelemetry) | `container.MustMake[contract.Tracer](c)` |
| **熔断器** | `circuitbreaker/noop` | sentinel | `container.MustMake[contract.CircuitBreaker](c)` |
| **限流器** | `circuitbreaker/noop` | sentinel | `container.MustMake[contract.RateLimiter](c)` |
| **RPC** | `rpc/noop` | grpc | `container.MustMake[contract.RPCClient](c)` |
| **分布式锁** | `dlock/noop` | redis, etcd | `container.MustMake[contract.DLock](c)` |
| **消息队列** | `messagequeue/noop` | kafka, rabbitmq | `container.MustMake[contract.MessageQueue](c)` |
| **HTTP Metrics** | 内置 | prometheus | `gin.MetricsMiddleware()` |

## 单体阶段（默认）

使用 `MonolithFriendlyProviders()`，所有微服务能力都是 noop 实现，零依赖：

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

切换到真实实现，只需替换对应的 Provider：

```go
package main

import (
    "github.com/ngq/gorp/framework/bootstrap"
    "github.com/ngq/gorp/framework/contract"
    
    // 真实实现
    discoveryconsul "github.com/ngq/gorp/framework/provider/discovery/consul"
    selectorp2c "github.com/ngq/gorp/framework/provider/selector/p2c"
    tracingotel "github.com/ngq/gorp/framework/provider/tracing/otel"
    cbsentinel "github.com/ngq/gorp/framework/provider/circuitbreaker/sentinel"
)

func main() {
    // 微服务模式：使用真实实现
    providers := []contract.ServiceProvider{
        // 基础能力
        bootstrap.FoundationProviders()...,
        bootstrap.ORMRuntimeProviders()...,
        
        // 微服务能力（替换 noop）
        discoveryconsul.NewProvider(),     // Consul 服务发现
        selectorp2c.NewProvider(),         // P2C 负载均衡
        tracingotel.NewProvider(),         // OpenTelemetry 链路追踪
        cbsentinel.NewProvider(),          // Sentinel 熔断/限流
    }
    
    app, c, err := bootstrap.Init(bootstrap.Options{
        ExtraProviders: providers,
    })
}
```

## 使用示例

### 1. HTTP Metrics（已自动集成）

```go
// framework/provider/gin/metrics.go 已提供
// 在服务启动时自动注册，无需手动配置
// 访问 /metrics 端点即可获取 Prometheus 指标
```

### 2. 熔断器保护下游调用

```go
import "github.com/ngq/gorp/framework/container"

func callDownstream(c contract.Container, url string) error {
    // 获取熔断器
    cb := container.MustMakeCircuitBreaker(c)
    
    // 使用熔断器保护调用
    return cb.Do(context.Background(), "downstream-service", func() error {
        resp, err := http.Get(url)
        // ...
        return err
    })
}
```

### 3. 限流保护 API

```go
func apiHandler(c contract.Container) gin.HandlerFunc {
    rl := container.MustMakeRateLimiter(c)
    
    return func(ctx *gin.Context) {
        // 检查是否允许请求
        if err := rl.Allow(ctx, "api:/v1/users"); err != nil {
            ctx.JSON(429, gin.H{"error": "rate limited"})
            ctx.Abort()
            return
        }
        
        // 处理请求
        // ...
    }
}
```

### 4. 服务发现 + 负载均衡

```go
func callService(c contract.Container, serviceName string) error {
    // 获取服务发现
    disc := container.MustMakeDiscovery(c)
    
    // 发现服务实例
    instances, err := disc.Discover(context.Background(), serviceName)
    if err != nil {
        return err
    }
    
    // 获取负载均衡器
    sel := container.MustMakeSelector(c)
    
    // 选择一个实例
    instance, err := sel.Select(instances)
    if err != nil {
        return err
    }
    
    // 调用服务
    url := fmt.Sprintf("http://%s:%d/api", instance.Address, instance.Port)
    // ...
    return nil
}
```

### 5. 链路追踪

```go
func tracedOperation(c contract.Container) error {
    tracer := container.MustMakeTracer(c)
    
    // 创建 Span
    ctx, span := tracer.Start(context.Background(), "operation-name")
    defer span.End()
    
    // 设置属性
    span.SetAttributes("key", "value")
    
    // 记录错误
    if err != nil {
        span.RecordError(err)
    }
    
    return err
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