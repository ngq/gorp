package gin

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

func MetricsMiddleware() transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
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
