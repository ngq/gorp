package observability

import (
	"github.com/ngq/gorp/framework/contract"
)

// Provider 观测服务提供者。
type Provider struct {
	config contract.ObservabilityConfig
}

// NewProvider 创建观测服务提供者。
func NewProvider(config contract.ObservabilityConfig) *Provider {
	return &Provider{config: config}
}

// Name returns the provider name.
func (p *Provider) Name() string { return "observability" }

// IsDefer returns false.
func (p *Provider) IsDefer() bool { return false }

// Provides returns the keys this provider provides.
func (p *Provider) Provides() []string { return []string{contract.ObservabilityKey} }

// Register binds the observability service to the container.
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ObservabilityKey, func(c contract.Container) (interface{}, error) {
		// 获取依赖服务
		loggerAny, _ := c.Make(contract.LogKey)
		logger, _ := loggerAny.(contract.Logger)

		errorReporterAny, _ := c.Make(contract.ErrorReporterKey)
		errorReporter, _ := errorReporterAny.(contract.ErrorReporter)

		// 创建 Metrics 实现
		var metrics contract.Metrics
		if p.config.MetricsEnabled {
			metrics = NewPrometheusMetrics()
		}

		// 创建 Tracer 实现
		var tracer contract.Tracer
		if p.config.TracingEnabled {
			// TODO: 对接 OpenTelemetry
			tracer = NewNoopTracer()
		} else {
			tracer = NewNoopTracer()
		}

		return NewDefaultObservability(metrics, tracer, logger, errorReporter), nil
	}, true)
	return nil
}

// Boot does nothing.
func (p *Provider) Boot(contract.Container) error { return nil }