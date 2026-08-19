// Application scenarios:
// - Collect baseline request count and latency metrics for HTTP routes.
// - Export simple per-method / per-path / per-status observability signals.
// - Provide a low-cost metrics baseline before adding custom business metrics.
//
// 适用场景：
// - 为 HTTP 路由采集基础请求量和耗时指标。
// - 输出按 method / path / status 维度聚合的观测信号。
// - 在接入业务自定义指标前，先提供低成本的通用指标基线。
package middleware

import (
	"strconv"
	"time"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	prometheus "github.com/prometheus/client_golang/prometheus"
	promauto "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gorp_http_requests_total",
		Help: "Total number of HTTP requests handled by Gin.",
	}, []string{"service", "method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gorp_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method", "path", "status"})
)

// MetricsMiddleware records basic HTTP request metrics for the current request.
//
// MetricsMiddleware 为当前请求记录基础 HTTP 指标。
func MetricsMiddleware(serviceNames ...string) transportcontract.Middleware {
	serviceName := "default"
	if len(serviceNames) > 0 && serviceNames[0] != "" {
		serviceName = serviceNames[0]
	}
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			start := time.Now()
			if next != nil {
				next(c)
			}

			path := c.RoutePath()
			if path == "" {
				// 未匹配到路由模板（如 404 或非法路径）时统一收敛为 "unmatched"，
				// 避免原始 URL 中的动态 ID 导致 Prometheus 标签高基数内存爆炸。
				path = "unmatched"
			}
			status := strconv.Itoa(c.ResponseStatus())
			if status == "0" {
				status = strconv.Itoa(200)
			}
			method := ""
			if c.Request() != nil {
				method = c.Request().Method
			}
			duration := time.Since(start).Seconds()

			httpRequestsTotal.WithLabelValues(serviceName, method, path, status).Inc()
			httpRequestDuration.WithLabelValues(serviceName, method, path, status).Observe(duration)
		}
	}
}
