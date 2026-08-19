package security

import (
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

// Provider is the DI provider for security Authorizer.
type Provider struct {
	Authorizer securitycontract.Authorizer
}

// NewProvider creates a new security Provider.
func NewProvider(authorizer ...securitycontract.Authorizer) *Provider {
	var a securitycontract.Authorizer
	if len(authorizer) > 0 && authorizer[0] != nil {
		a = authorizer[0]
	} else {
		a = NewMemoryAuthorizer()
	}
	return &Provider{Authorizer: a}
}

// Register binds AuthorizerKey in container.
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(securitycontract.AuthorizerKey, func(c runtimecontract.Container) (any, error) {
		return p.Authorizer, nil
	}, true)
	return nil
}

// Boot initializes provider.
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}
