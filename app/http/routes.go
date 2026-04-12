package http

import (
	ginprovider "github.com/ngq/gorp/framework/provider/gin"
	"github.com/ngq/gorp/framework/contract"
)

// RegisterRoutes 把当前项目需要暴露的 HTTP 路由统一挂载到 Gin 引擎上。
func RegisterRoutes(c contract.Container) error {
	_, err := ginprovider.EngineFromContainer(c)
	if err != nil {
		return err
	}

	// Prometheus 指标端点
	if err := RegisterMetrics(c); err != nil {
		return err
	}

	// 健康探针端点：/healthz, /readyz
	if err := RegisterHealthProbes(c); err != nil {
		return err
	}

	// pprof 性能分析端点
	if err := RegisterPprof(c); err != nil {
		return err
	}

	// swagger
	if err := RegisterSwagger(c); err != nil {
		return err
	}

	// 前端托管
	if err := RegisterFrontend(c); err != nil {
		return err
	}
	return nil
}

// RegisterMetrics 注册 Prometheus 指标相关路由和收集器。
func RegisterMetrics(c contract.Container) error {
	engine, err := ginprovider.EngineFromContainer(c)
	if err != nil {
		return err
	}

	ginprovider.RegisterGoRuntimeMetrics()
	engine.GET("/metrics", ginprovider.PrometheusHandler())

	return nil
}
