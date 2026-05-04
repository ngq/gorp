package ssh

import (
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string       { return "ssh" }
func (p *Provider) IsDefer() bool      { return true }
func (p *Provider) Provides() []string { return []string{integrationcontract.SSHKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.SSHKey, func(c runtimecontract.Container) (any, error) {
		return NewService(c)
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }
