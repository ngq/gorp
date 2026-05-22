// Application scenarios:
// - Define health check contracts for services and their dependencies.
// - Support aggregated health status reporting for microservices.
// - Provide unified health check endpoint for HTTP and gRPC.
//
// 适用场景：
// - 定义服务及其依赖的健康检查契约。
// - 支持微服务的聚合健康状态报告。
// - 为 HTTP 和 gRPC 提供统一的健康检查端点。
package observability

import (
	"context"
	"time"
)

// HealthCheckerKey is the container key for the health checker capability.
//
// HealthCheckerKey 是健康检查能力的容器键。
const HealthCheckerKey = "framework.health.checker"

// HealthStatus represents the health status of a component or service.
//
// HealthStatus 表示组件或服务的健康状态。
type HealthStatus string

const (
	// HealthStatusHealthy indicates the component is healthy.
	//
	// HealthStatusHealthy 表示组件健康。
	HealthStatusHealthy HealthStatus = "healthy"

	// HealthStatusUnhealthy indicates the component is unhealthy.
	//
	// HealthStatusUnhealthy 表示组件不健康。
	HealthStatusUnhealthy HealthStatus = "unhealthy"

	// HealthStatusDegraded indicates the component is degraded but functional.
	//
	// HealthStatusDegraded 表示组件降级但仍可用。
	HealthStatusDegraded HealthStatus = "degraded"
)

// HealthCheckResult represents the result of a health check.
//
// HealthCheckResult 表示健康检查的结果。
type HealthCheckResult struct {
	// Name is the name of the component being checked.
	//
	// Name 是被检查组件的名称。
	Name string `json:"name"`

	// Status is the health status of the component.
	//
	// Status 是组件的健康状态。
	Status HealthStatus `json:"status"`

	// Message provides additional details about the health status.
	//
	// Message 提供健康状态的额外详情。
	Message string `json:"message,omitempty"`

	// Error contains the error if the health check failed.
	//
	// Error 包含健康检查失败的错误信息。
	Error string `json:"error,omitempty"`

	// Latency is the time taken to perform the health check.
	//
	// Latency 是执行健康检查所花费的时间。
	Latency time.Duration `json:"latency"`

	// Timestamp is when the health check was performed.
	//
	// Timestamp 是健康检查执行的时间戳。
	Timestamp time.Time `json:"timestamp"`

	// Metadata contains additional component-specific information.
	//
	// Metadata 包含额外的组件特定信息。
	Metadata map[string]any `json:"metadata,omitempty"`
}

// HealthReport represents an aggregated health report for a service.
//
// HealthReport 表示服务的聚合健康报告。
type HealthReport struct {
	// Service is the name of the service.
	//
	// Service 是服务名称。
	Service string `json:"service"`

	// Status is the overall health status of the service.
	//
	// Status 是服务的整体健康状态。
	Status HealthStatus `json:"status"`

	// Version is the service version.
	//
	// Version 是服务版本。
	Version string `json:"version"`

	// Timestamp is when the report was generated.
	//
	// Timestamp 是报告生成的时间戳。
	Timestamp time.Time `json:"timestamp"`

	// Checks contains individual health check results.
	//
	// Checks 包含各个健康检查的结果。
	Checks map[string]HealthCheckResult `json:"checks"`

	// Dependencies contains health status of dependent services.
	//
	// Dependencies 包含依赖服务的健康状态。
	Dependencies map[string]DependencyHealth `json:"dependencies,omitempty"`
}

// DependencyHealth represents the health status of a dependency.
//
// DependencyHealth 表示依赖的健康状态。
type DependencyHealth struct {
	// Name is the name of the dependency.
	//
	// Name 是依赖的名称。
	Name string `json:"name"`

	// Type is the type of dependency (e.g., "database", "redis", "grpc").
	//
	// Type 是依赖的类型（如 "database", "redis", "grpc"）。
	Type string `json:"type"`

	// Status is the health status of the dependency.
	//
	// Status 是依赖的健康状态。
	Status HealthStatus `json:"status"`

	// Message provides additional details.
	//
	// Message 提供额外详情。
	Message string `json:"message,omitempty"`

	// Latency is the time taken to check the dependency.
	//
	// Latency 是检查依赖所花费的时间。
	Latency time.Duration `json:"latency"`
}

// HealthChecker defines the health checking contract.
// Implementations aggregate health status from multiple components.
//
// HealthChecker 定义健康检查契约。
// 实现聚合多个组件的健康状态。
type HealthChecker interface {
	// Check performs a health check and returns the aggregated report.
	//
	// Check 执行健康检查并返回聚合报告。
	Check(ctx context.Context) (*HealthReport, error)

	// CheckComponent performs a health check for a specific component.
	//
	// CheckComponent 对特定组件执行健康检查。
	CheckComponent(ctx context.Context, name string) (*HealthCheckResult, error)

	// AddChecker registers a component health checker.
	//
	// AddChecker 注册组件健康检查器。
	AddChecker(name string, checker ComponentChecker)

	// AddDependency registers a dependency health checker.
	//
	// AddDependency 注册依赖健康检查器。
	AddDependency(name string, dep DependencyChecker)
}

// ComponentChecker is a function that checks the health of a component.
//
// ComponentChecker 是检查组件健康的函数。
type ComponentChecker func(ctx context.Context) HealthCheckResult

// DependencyChecker is a function that checks the health of a dependency.
//
// DependencyChecker 是检查依赖健康的函数。
type DependencyChecker func(ctx context.Context) DependencyHealth

// HealthCheckerConfig configures the health checker.
//
// HealthCheckerConfig 配置健康检查器。
type HealthCheckerConfig struct {
	// ServiceName is the name of the service.
	//
	// ServiceName 是服务名称。
	ServiceName string `mapstructure:"service_name"`

	// Version is the service version.
	//
	// Version 是服务版本。
	Version string `mapstructure:"version"`

	// Timeout is the timeout for individual health checks.
	//
	// Timeout 是单个健康检查的超时时间。
	Timeout time.Duration `mapstructure:"timeout"`

	// CheckDependencies indicates whether to check dependencies.
	//
	// CheckDependencies 表示是否检查依赖。
	CheckDependencies bool `mapstructure:"check_dependencies"`
}
