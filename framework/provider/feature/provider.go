package feature

import (
	featurecontract "github.com/ngq/gorp/framework/contract/feature"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider is the DI provider for feature flag capabilities.
type Provider struct {
	Service *Service
}

// NewProvider creates a new feature Provider instance.
func NewProvider(initialFlags ...map[string]FlagRule) *Provider {
	return &Provider{
		Service: NewService(initialFlags...),
	}
}

// Register binds FeatureKey in the DI container.
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(featurecontract.FeatureKey, func(c runtimecontract.Container) (any, error) {
		if p.Service != nil {
			return p.Service, nil
		}
		return NewService(), nil
	}, true)
	return nil
}

// Boot initializes the provider.
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}
