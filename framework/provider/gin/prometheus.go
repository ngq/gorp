package gin

import (
	"errors"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusHandler 返回标准 http.Handler，用于暴露 /metrics 端点。
//
// 中文说明：
// - 默认主线不再直接暴露 Gin HandlerFunc；
// - 通过标准 `http.Handler`，framework 可以统一走 `Mount(path, http.Handler)` 主线；
// - provider 若仍基于 Gin，可在 provider 内部完成适配。
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
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
