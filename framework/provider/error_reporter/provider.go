package error_reporter

import (
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct {
	config resiliencecontract.ErrorReporterConfig
}

func NewProvider(config resiliencecontract.ErrorReporterConfig) *Provider {
	return &Provider{config: config}
}

func (p *Provider) Name() string { return "error_reporter" }

func (p *Provider) IsDefer() bool { return false }

func (p *Provider) Provides() []string { return []string{resiliencecontract.ErrorReporterKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(resiliencecontract.ErrorReporterKey, func(c runtimecontract.Container) (interface{}, error) {
		if p.config.Enabled && p.config.DSN != "" {
			return NewSentryAdapter(p.config), nil
		}

		logAny, err := c.Make(observabilitycontract.LogKey)
		if err != nil {
			return nil, err
		}
		logger := logAny.(observabilitycontract.Logger)
		return NewLogReporter(logger), nil
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }
