package gin

import (
	"errors"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusHandler 返回一个 Gin HandlerFunc，用于暴露 /metrics 端点。
//
// 中文说明：
// - 内部直接包装 promhttp.Handler()，适配 Gin 的 HandlerFunc 接口；
// - 用于在 app/http/routes.go 中注册 GET /metrics 路由；
// - 会暴露所有已注册的 Prometheus 指标（包括 HTTP 指标 + Go runtime 指标）。
func PrometheusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	}
}

var registerGoRuntimeMetricsOnce sync.Once

// RegisterGoRuntimeMetrics 注册 Go runtime 指标到 Prometheus 默认注册器。
//
// 中文说明：
// - 包括 goroutine 数量、GC 暂停时间、内存分配等关键运行时指标；
// - 这些指标对排查内存泄漏、goroutine 泄漏、GC 问题非常有帮助；
// - 默认 registry 在部分运行场景下可能已经带有 GoCollector；
// - 这里使用 once + AlreadyRegistered 容忍重复注册，避免真实启动时 panic。
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
