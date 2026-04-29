package observability

import (
	"github.com/ngq/gorp/framework/contract"
)

// Provider 观测服务提供者。
//
// 中文说明：
// - 对外统一暴露 contract.ObservabilityKey；
// - 把 metrics、tracer、logger、error reporter 聚合成一个总入口。
type Provider struct {
	config contract.ObservabilityConfig
}

// NewProvider 创建观测服务提供者。
func NewProvider(config contract.ObservabilityConfig) *Provider {
	return &Provider{config: config}
}

// Name 返回 provider 名称。
func (p *Provider) Name() string { return "observability" }

// IsDefer 表示 observability provider 不走延迟加载。
func (p *Provider) IsDefer() bool { return false }

// Provides 返回当前 provider 暴露的能力 key。
func (p *Provider) Provides() []string { return []string{contract.ObservabilityKey} }

// Register 绑定统一观测服务。
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

// Boot observability provider 无额外启动逻辑。
func (p *Provider) Boot(contract.Container) error { return nil }