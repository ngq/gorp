// Application scenarios:
// - Expose a standard `/metrics` endpoint for Prometheus scraping.
// - Register Go runtime metrics such as GC, memory, and goroutine statistics.
// - Reuse `router.Mount()` with a plain `http.Handler` instead of binding Gin-specific handlers.
//
// 适用场景：
// - 暴露标准 `/metrics` 端点供 Prometheus 抓取。
// - 注册 Go 运行时指标，例如 GC、内存和 goroutine 统计。
// - 通过普通 `http.Handler` 复用 `router.Mount()`，而不是绑定 Gin 专属 handler。
package gin

import (
	"errors"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusHandler returns a standard http.Handler for exposing Prometheus metrics.
//
// PrometheusHandler 返回一个用于暴露 Prometheus 指标的标准 http.Handler。
//
// Example:
//
//	router.Mount("/metrics", ginprovider.PrometheusHandler())
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

var registerGoRuntimeMetricsOnce sync.Once

// RegisterGoRuntimeMetrics registers Go runtime collectors into the default Prometheus registry.
//
// RegisterGoRuntimeMetrics 将 Go 运行时采集器注册到默认 Prometheus 注册表。
func RegisterGoRuntimeMetrics() {
	registerGoRuntimeMetricsOnce.Do(func() {
		collector := collectors.NewGoCollector()
		if err := prometheus.Register(collector); err != nil {
			var alreadyRegisteredErr prometheus.AlreadyRegisteredError
			if errors.As(err, &alreadyRegisteredErr) {
				return
			}
			panic(err)
		}
	})
}
