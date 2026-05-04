package event

import (
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string       { return "event" }
func (p *Provider) IsDefer() bool      { return false }
func (p *Provider) Provides() []string { return []string{integrationcontract.EventKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.EventKey, func(c runtimecontract.Container) (interface{}, error) {
		return NewLocalEventBus(), nil
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }
