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
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gorp_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})
)

// MetricsMiddleware records basic HTTP request metrics for the current request.
//
// MetricsMiddleware 为当前请求记录基础 HTTP 指标。
func MetricsMiddleware() transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			start := time.Now()
			if next != nil {
				next(c)
			}

			path := c.RoutePath()
			if path == "" && c.Request() != nil && c.Request().URL != nil {
				path = c.Request().URL.Path
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

			httpRequestsTotal.WithLabelValues(method, path, status).Inc()
			httpRequestDuration.WithLabelValues(method, path, status).Observe(duration)
		}
	}
}
