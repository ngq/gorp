package noop

import (
	"context"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "discovery.noop" }

func (p *Provider) IsDefer() bool { return true }

func (p *Provider) Provides() []string { return []string{transportcontract.RPCRegistryKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCRegistryKey, func(c runtimecontract.Container) (any, error) {
		return &noopRegistry{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }

type noopRegistry struct{}

func (r *noopRegistry) Register(_ context.Context, _, _ string, _ map[string]string) error {
	return nil
}

func (r *noopRegistry) Deregister(_ context.Context, _, _ string) error {
	return nil
}

func (r *noopRegistry) Discover(_ context.Context, _ string) ([]transportcontract.ServiceInstance, error) {
	return nil, nil
}

func (r *noopRegistry) Close() error { return nil }
