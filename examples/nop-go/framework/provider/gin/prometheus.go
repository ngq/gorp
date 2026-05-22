// Package gin provides Gin-based HTTP server implementation for gorp framework.
// This file provides Prometheus metrics handler and Go runtime metrics registration.
// Exposes /metrics endpoint and registers GC, memory, goroutine collectors.
//
// Gin HTTP 服务包，提供基于 Gin 的 HTTP 服务器实现，用于 gorp 框架。
// 本文件提供 Prometheus 指标 handler 和 Go 运行时指标注册。
// 暴露 /metrics 端点并注册 GC、内存、goroutine 采集器。
//
// Eg:
//
//	router.Mount("/metrics", ginprovider.PrometheusHandler())
//	ginprovider.RegisterGoRuntimeMetrics()
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
