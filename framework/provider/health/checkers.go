// Package health provides dependency health checkers for common components.
// Includes database, Redis, and gRPC service health checkers.
//
// 提供常用组件的依赖健康检查器，包括数据库、Redis 和 gRPC 服务。
package health

import (
	"context"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// --- 数据库健康检查器 ---

// DatabaseHealthChecker 创建数据库健康检查器。
// 使用 DBInspector 契约的 Ping 方法检查数据库连接。
//
// 使用方式：
//
//	checker := health.DatabaseHealthChecker(container, "mysql")
//	healthChecker.AddDependency("mysql", checker)
func DatabaseHealthChecker(c runtimecontract.Container, name string) observabilitycontract.DependencyChecker {
	return func(ctx context.Context) observabilitycontract.DependencyHealth {
		start := time.Now()

		// 检查容器是否绑定 DBInspector
		if c == nil || !c.IsBind(datacontract.DBInspectorKey) {
			return observabilitycontract.DependencyHealth{
				Name:    name,
				Type:    "database",
				Status:  observabilitycontract.HealthStatusDegraded,
				Message: "database inspector not configured",
				Latency: time.Since(start),
			}
		}

		// 获取 DBInspector
		inspectorAny, err := c.Make(datacontract.DBInspectorKey)
		if err != nil {
			return observabilitycontract.DependencyHealth{
				Name:    name,
				Type:    "database",
				Status:  observabilitycontract.HealthStatusUnhealthy,
				Message: "failed to get database inspector",
				Latency: time.Since(start),
			}
		}

		inspector, ok := inspectorAny.(datacontract.DBInspector)
		if !ok {
			return observabilitycontract.DependencyHealth{
				Name:    name,
				Type:    "database",
				Status:  observabilitycontract.HealthStatusUnhealthy,
				Message: "invalid database inspector type",
				Latency: time.Since(start),
			}
		}

		// 执行 Ping 检查
		if err := inspector.Ping(ctx); err != nil {
			return observabilitycontract.DependencyHealth{
				Name:    name,
				Type:    "database",
				Status:  observabilitycontract.HealthStatusUnhealthy,
				Message: "database ping failed",
				Latency: time.Since(start),
			}
		}

		return observabilitycontract.DependencyHealth{
			Name:    name,
			Type:    "database",
			Status:  observabilitycontract.HealthStatusHealthy,
			Message: "database connection is healthy",
			Latency: time.Since(start),
		}
	}
}

// --- Redis 健康检查器 ---

// RedisHealthChecker 创建 Redis 健康检查器。
// 使用 Redis 契约的 Ping 方法检查 Redis 连接。
//
// 使用方式：
//
//	checker := health.RedisHealthChecker(container, "redis")
//	healthChecker.AddDependency("redis", checker)
func RedisHealthChecker(c runtimecontract.Container, name string) observabilitycontract.DependencyChecker {
	return func(ctx context.Context) observabilitycontract.DependencyHealth {
		start := time.Now()

		// 检查容器是否绑定 Redis
		if c == nil || !c.IsBind(datacontract.RedisKey) {
			return observabilitycontract.DependencyHealth{
				Name:    name,
				Type:    "redis",
				Status:  observabilitycontract.HealthStatusDegraded,
				Message: "redis not configured",
				Latency: time.Since(start),
			}
		}

		// 获取 Redis 服务
		redisAny, err := c.Make(datacontract.RedisKey)
		if err != nil {
			return observabilitycontract.DependencyHealth{
				Name:    name,
				Type:    "redis",
				Status:  observabilitycontract.HealthStatusUnhealthy,
				Message: "failed to get redis service",
				Latency: time.Since(start),
			}
		}

		// 尝试 Ping Redis
		// Redis 契约可能有不同的 Ping 方法签名
		// 使用类型断言检查是否支持 Ping
		switch redis := redisAny.(type) {
		case interface{ Ping(ctx context.Context) error }:
			if err := redis.Ping(ctx); err != nil {
				return observabilitycontract.DependencyHealth{
					Name:    name,
					Type:    "redis",
					Status:  observabilitycontract.HealthStatusUnhealthy,
					Message: "redis ping failed",
					Latency: time.Since(start),
				}
			}
		case interface{ Ping() error }:
			if err := redis.Ping(); err != nil {
				return observabilitycontract.DependencyHealth{
					Name:    name,
					Type:    "redis",
					Status:  observabilitycontract.HealthStatusUnhealthy,
					Message: "redis ping failed",
					Latency: time.Since(start),
				}
			}
		default:
			// 如果 Redis 不支持 Ping，标记为 degraded
			return observabilitycontract.DependencyHealth{
				Name:    name,
				Type:    "redis",
				Status:  observabilitycontract.HealthStatusDegraded,
				Message: "redis does not support ping",
				Latency: time.Since(start),
			}
		}

		return observabilitycontract.DependencyHealth{
			Name:    name,
			Type:    "redis",
			Status:  observabilitycontract.HealthStatusHealthy,
			Message: "redis connection is healthy",
			Latency: time.Since(start),
		}
	}
}

// --- gRPC 服务健康检查器 ---

// GRPCServiceHealthChecker 创建 gRPC 服务健康检查器。
// 使用 gRPC 健康检查协议检查远程服务状态。
//
// 使用方式：
//
//	checker := health.GRPCServiceHealthChecker(container, "user-service", "user-service")
//	healthChecker.AddDependency("user-service", checker)
func GRPCServiceHealthChecker(c runtimecontract.Container, serviceName string, target string) observabilitycontract.DependencyChecker {
	return func(ctx context.Context) observabilitycontract.DependencyHealth {
		start := time.Now()

		// 检查容器是否绑定 RPC Client
		// 如果没有 RPC Client，标记为 degraded（非必需依赖）
		if c == nil || !c.IsBind("framework.rpc.client") {
			return observabilitycontract.DependencyHealth{
				Name:    serviceName,
				Type:    "grpc",
				Status:  observabilitycontract.HealthStatusDegraded,
				Message: "rpc client not configured",
				Latency: time.Since(start),
			}
		}

		// 获取 RPC Client
		clientAny, err := c.Make("framework.rpc.client")
		if err != nil {
			return observabilitycontract.DependencyHealth{
				Name:    serviceName,
				Type:    "grpc",
				Status:  observabilitycontract.HealthStatusUnhealthy,
				Message: "failed to get rpc client",
				Latency: time.Since(start),
			}
		}

		// 尝试调用健康检查
		// 使用简单的 Call 方法检查服务是否可达
		// 注意：这里不使用标准的 gRPC 健康检查协议，
		// 因为 RPC Client 契约是通用的
		client, ok := clientAny.(interface {
			Call(ctx context.Context, service, method string, req, resp any) error
		})
		if !ok {
			return observabilitycontract.DependencyHealth{
				Name:    serviceName,
				Type:    "grpc",
				Status:  observabilitycontract.HealthStatusDegraded,
				Message: "rpc client does not support Call",
				Latency: time.Since(start),
			}
		}

		// 执行健康检查调用
		// 使用空的请求和响应，只检查连接是否可达
		healthCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		// 尝试调用一个健康检查方法
		// 如果服务不可达，会返回错误
		err = client.Call(healthCtx, target, "grpc.health.v1.Health/Check", nil, nil)
		if err != nil {
			return observabilitycontract.DependencyHealth{
				Name:    serviceName,
				Type:    "grpc",
				Status:  observabilitycontract.HealthStatusUnhealthy,
				Message: "grpc service is unreachable",
				Latency: time.Since(start),
			}
		}

		return observabilitycontract.DependencyHealth{
			Name:    serviceName,
			Type:    "grpc",
			Status:  observabilitycontract.HealthStatusHealthy,
			Message: "grpc service is healthy",
			Latency: time.Since(start),
		}
	}
}

// --- 自定义健康检查器 ---

// CustomDependencyChecker 创建自定义依赖健康检查器。
// 用户可以传入自定义的检查函数。
//
// 使用方式：
//
//	checker := health.CustomDependencyChecker("custom", "http", func(ctx context.Context) (bool, string) {
//	    // 自定义检查逻辑
//	    return true, "custom service is healthy"
//	})
//	healthChecker.AddDependency("custom", checker)
func CustomDependencyChecker(name string, depType string, checkFunc func(ctx context.Context) (bool, string)) observabilitycontract.DependencyChecker {
	return func(ctx context.Context) observabilitycontract.DependencyHealth {
		start := time.Now()

		if checkFunc == nil {
			return observabilitycontract.DependencyHealth{
				Name:    name,
				Type:    depType,
				Status:  observabilitycontract.HealthStatusDegraded,
				Message: "check function not provided",
				Latency: time.Since(start),
			}
		}

		healthy, message := checkFunc(ctx)
		status := observabilitycontract.HealthStatusHealthy
		if !healthy {
			status = observabilitycontract.HealthStatusUnhealthy
		}

		return observabilitycontract.DependencyHealth{
			Name:    name,
			Type:    depType,
			Status:  status,
			Message: message,
			Latency: time.Since(start),
		}
	}
}