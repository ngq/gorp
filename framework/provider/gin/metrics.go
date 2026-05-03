package gin

import (
	"strconv"
	"time"

	prometheus "github.com/prometheus/client_golang/prometheus"
	promauto "github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/ngq/gorp/framework/contract"
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

// MetricsMiddleware 记录最小 HTTP 请求指标。
//
// 中文说明：
// - 第一阶段先聚焦最稳的通用指标：请求总量 + 请求耗时；
// - 路由标签优先使用 `RoutePath()`，避免把具体 ID 参数打散成高基数；
// - 若当前请求没有命中已注册路由，则退回 URL.Path。
func MetricsMiddleware() contract.HTTPMiddleware {
	return func(c contract.HTTPContext, next contract.HTTPNext) {
		start := time.Now()
		if next != nil {
			next()
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
