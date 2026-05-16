// Package observability provides observability provider for gorp framework.
// Registers unified observability service with metrics, tracing, logging, error reporting.
//
// 可观测性包提供可观测性 provider，用于 gorp 框架。
// 注册统一的可观测性服务，包括指标、追踪、日志、错误上报。
package observability

import (
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers unified observability service.
// Core logic: Aggregate metrics, tracer, logger, error reporter into single service.
//
// Provider 注册统一的可观测性服务。
// 核心逻辑：将指标、追踪器、日志器、错误上报器聚合为单一服务。
type Provider struct {
	config observabilitycontract.ObservabilityConfig
}

// NewProvider creates a new observability provider with configuration.
//
// NewProvider 创建新的可观测性 provider，携带配置。
func NewProvider(config observabilitycontract.ObservabilityConfig) *Provider {
	return &Provider{config: config}
}

// Name returns provider name for identification.
//
// Name 返回 provider 名称，用于标识。
func (p *Provider) Name() string { return "observability" }

// IsDefer indicates observability should not defer loading.
// Must be available early for metrics and tracing.
//
// IsDefer 表示可观测性不应延迟加载。
// 必须尽早可用以支持指标和追踪。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the capability keys this provider exposes.
// Exposes ObservabilityKey for unified observability service.
//
// Provides 返回 provider 暴露的能力键。
// 暴露 ObservabilityKey 用于统一可观测性服务。
func (p *Provider) Provides() []string { return []string{observabilitycontract.ObservabilityKey} }

// Register binds the observability factory to the container.
// Core logic: Aggregate all components, create DefaultObservability, bind to container.
//
// Register 将可观测性工厂绑定到容器。
// 核心逻辑：聚合所有组件、创建 DefaultObservability、绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(observabilitycontract.ObservabilityKey, func(c runtimecontract.Container) (interface{}, error) {
		loggerAny, _ := c.Make(observabilitycontract.LogKey)
		logger, _ := loggerAny.(observabilitycontract.Logger)

		errorReporterAny, _ := c.Make(resiliencecontract.ErrorReporterKey)
		errorReporter, _ := errorReporterAny.(resiliencecontract.ErrorReporter)

		var metrics observabilitycontract.Metrics
		if p.config.MetricsEnabled {
			metrics = NewPrometheusMetrics()
		}

		var tracer observabilitycontract.Tracer
		if p.config.TracingEnabled {
			tracer = NewPrometheusTracer() // 启用时使用真实 tracer
		} else {
			tracer = NewNoopTracer()
		}

		return NewDefaultObservability(metrics, tracer, logger, errorReporter), nil
	}, true)
	return nil
}

// Boot initializes the observability provider.
// No additional startup logic required.
//
// Boot 初始化可观测性 provider。
// 无需额外启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }
