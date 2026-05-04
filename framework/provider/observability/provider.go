package observability

import (
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct {
	config observabilitycontract.ObservabilityConfig
}

func NewProvider(config observabilitycontract.ObservabilityConfig) *Provider {
	return &Provider{config: config}
}

func (p *Provider) Name() string       { return "observability" }
func (p *Provider) IsDefer() bool      { return false }
func (p *Provider) Provides() []string { return []string{observabilitycontract.ObservabilityKey} }

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
			tracer = NewNoopTracer()
		} else {
			tracer = NewNoopTracer()
		}

		return NewDefaultObservability(metrics, tracer, logger, errorReporter), nil
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }
