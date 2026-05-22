// Package middleware provides HTTP health check endpoints for gorp framework.
// Supports /healthz, /readyz, and /livez endpoints with dependency aggregation.
//
// 提供 HTTP 健康检查端点中间件，支持 /healthz、/readyz 和 /livez 端点。
package middleware

import (
	"net/http"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// HealthCheckEndpoint 配置健康检查端点。
//
// HealthCheckEndpoint 配置健康检查端点。
type HealthCheckEndpoint struct {
	// Path 是健康检查端点路径。
	//
	// Path 是健康检查端点路径。
	Path string

	// Checker 是健康检查器。
	//
	// Checker 是健康检查器。
	Checker observabilitycontract.HealthChecker
}

// HealthCheckHandler 创建健康检查 HTTP handler。
// 返回聚合的健康报告，包含所有组件和依赖的状态。
//
// 使用方式：
//
//	checker, _ := container.Make(observabilitycontract.HealthCheckerKey)
//	router.GET("/healthz", middleware.HealthCheckHandler(checker))
func HealthCheckHandler(checker observabilitycontract.HealthChecker) transportcontract.Handler {
	return func(c transportcontract.Context) {
		if checker == nil {
			c.JSON(http.StatusServiceUnavailable, map[string]any{
				"status":  "unhealthy",
				"message": "health checker not configured",
			})
			return
		}

		report, err := checker.Check(c.Context())
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, map[string]any{
				"status":  "unhealthy",
				"message": "health check failed: " + err.Error(),
			})
			return
		}

		// 根据状态设置 HTTP 状态码
		statusCode := http.StatusOK
		if report.Status == observabilitycontract.HealthStatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		} else if report.Status == observabilitycontract.HealthStatusDegraded {
			statusCode = http.StatusOK // degraded 仍然返回 200，但状态标记为 degraded
		}

		c.JSON(statusCode, report)
	}
}

// HealthCheckHandlerFromContainer 从容器获取健康检查器并创建 handler。
//
// 使用方式：
//
//	router.GET("/healthz", middleware.HealthCheckHandlerFromContainer(container))
func HealthCheckHandlerFromContainer(container runtimecontract.Container) transportcontract.Handler {
	return func(c transportcontract.Context) {
		if container == nil || !container.IsBind(observabilitycontract.HealthCheckerKey) {
			c.JSON(http.StatusServiceUnavailable, map[string]any{
				"status":  "unhealthy",
				"message": "health checker not configured",
			})
			return
		}

		checkerAny, err := container.Make(observabilitycontract.HealthCheckerKey)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, map[string]any{
				"status":  "unhealthy",
				"message": "failed to get health checker",
			})
			return
		}

		checker, ok := checkerAny.(observabilitycontract.HealthChecker)
		if !ok {
			c.JSON(http.StatusServiceUnavailable, map[string]any{
				"status":  "unhealthy",
				"message": "invalid health checker type",
			})
			return
		}

		HealthCheckHandler(checker)(c)
	}
}

// ReadinessHandler 创建就绪检查 HTTP handler。
// 用于 Kubernetes readiness probe，检查服务是否准备好接收流量。
// 如果任何依赖不健康，返回 503。
//
// 使用方式：
//
//	router.GET("/readyz", middleware.ReadinessHandler(checker))
func ReadinessHandler(checker observabilitycontract.HealthChecker) transportcontract.Handler {
	return func(c transportcontract.Context) {
		if checker == nil {
			c.JSON(http.StatusServiceUnavailable, map[string]any{
				"ready":   false,
				"message": "health checker not configured",
			})
			return
		}

		report, err := checker.Check(c.Context())
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, map[string]any{
				"ready":   false,
				"message": "readiness check failed: " + err.Error(),
			})
			return
		}

		// 就绪检查：只有 healthy 状态才返回 ready
		ready := report.Status == observabilitycontract.HealthStatusHealthy

		statusCode := http.StatusOK
		if !ready {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, map[string]any{
			"ready":       ready,
			"status":      report.Status,
			"service":     report.Service,
			"checks":      report.Checks,
			"dependencies": report.Dependencies,
		})
	}
}

// ReadinessHandlerFromContainer 从容器获取健康检查器并创建就绪检查 handler。
//
// 使用方式：
//
//	router.GET("/readyz", middleware.ReadinessHandlerFromContainer(container))
func ReadinessHandlerFromContainer(container runtimecontract.Container) transportcontract.Handler {
	return func(c transportcontract.Context) {
		if container == nil || !container.IsBind(observabilitycontract.HealthCheckerKey) {
			c.JSON(http.StatusServiceUnavailable, map[string]any{
				"ready":   false,
				"message": "health checker not configured",
			})
			return
		}

		checkerAny, err := container.Make(observabilitycontract.HealthCheckerKey)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, map[string]any{
				"ready":   false,
				"message": "failed to get health checker",
			})
			return
		}

		checker, ok := checkerAny.(observabilitycontract.HealthChecker)
		if !ok {
			c.JSON(http.StatusServiceUnavailable, map[string]any{
				"ready":   false,
				"message": "invalid health checker type",
			})
			return
		}

		ReadinessHandler(checker)(c)
	}
}

// LivenessHandler 创建存活检查 HTTP handler。
// 用于 Kubernetes liveness probe，检查服务是否存活。
// 如果服务能响应，就认为存活。
//
// 使用方式：
//
//	router.GET("/livez", middleware.LivenessHandler())
func LivenessHandler() transportcontract.Handler {
	return func(c transportcontract.Context) {
		c.JSON(http.StatusOK, map[string]any{
			"alive":   true,
			"message": "service is alive",
		})
	}
}

// RegisterHealthEndpoints 注册所有健康检查端点。
// 包括 /healthz、/readyz 和 /livez。
//
// 使用方式：
//
//	middleware.RegisterHealthEndpoints(router, container)
func RegisterHealthEndpoints(router transportcontract.Router, container runtimecontract.Container) {
	if router == nil {
		return
	}

	// 注册健康检查端点
	router.GET("/healthz", HealthCheckHandlerFromContainer(container))

	// 注册就绪检查端点
	router.GET("/readyz", ReadinessHandlerFromContainer(container))

	// 注册存活检查端点
	router.GET("/livez", LivenessHandler())
}

// RegisterHealthEndpointsSimple 注册简单的健康检查端点（不检查依赖）。
// 适用于单体应用或不需要依赖检查的场景。
//
// 使用方式：
//
//	middleware.RegisterHealthEndpointsSimple(router, serviceName, version)
func RegisterHealthEndpointsSimple(router transportcontract.Router, serviceName, version string) {
	if router == nil {
		return
	}

	// 简单的健康检查
	router.GET("/healthz", func(c transportcontract.Context) {
		c.JSON(http.StatusOK, map[string]any{
			"status":  "healthy",
			"service": serviceName,
			"version": version,
		})
	})

	// 简单的就绪检查
	router.GET("/readyz", func(c transportcontract.Context) {
		c.JSON(http.StatusOK, map[string]any{
			"ready":   true,
			"service": serviceName,
		})
	})

	// 存活检查
	router.GET("/livez", LivenessHandler())
}