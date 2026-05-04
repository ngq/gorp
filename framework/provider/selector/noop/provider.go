package noop

import (
	"context"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "selector.noop" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{discoverycontract.SelectorKey, discoverycontract.SelectorBuilderKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(discoverycontract.SelectorBuilderKey, func(c runtimecontract.Container) (any, error) {
		return &noopBuilder{}, nil
	}, true)

	c.Bind(discoverycontract.SelectorKey, func(c runtimecontract.Container) (any, error) {
		builder := &noopBuilder{}
		return builder.Build(), nil
	}, true)

	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

type noopBuilder struct{}

func (b *noopBuilder) Build() discoverycontract.Selector {
	return &noopSelector{}
}

type noopSelector struct{}

func (s *noopSelector) Select(ctx context.Context, instances []transportcontract.ServiceInstance, opts ...discoverycontract.SelectOption) (
	selected transportcontract.ServiceInstance, done discoverycontract.DoneFunc, err error,
) {
	options := &discoverycontract.SelectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.ForceInstance != nil {
		return *options.ForceInstance, noopDone, nil
	}

	for _, instance := range instances {
		if instance.Healthy {
			return instance, noopDone, nil
		}
	}

	return transportcontract.ServiceInstance{}, noopDone, discoverycontract.ErrNoAvailable
}

func noopDone(ctx context.Context, info discoverycontract.DoneInfo) {}
