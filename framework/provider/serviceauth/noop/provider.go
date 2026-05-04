package noop

import (
	"context"
	"crypto/x509"
	"errors"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "serviceauth.noop" }

func (p *Provider) IsDefer() bool { return true }

func (p *Provider) Provides() []string {
	return []string{securitycontract.ServiceAuthKey, securitycontract.ServiceIdentityKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(securitycontract.ServiceAuthKey, func(c runtimecontract.Container) (any, error) {
		return &noopAuthenticator{}, nil
	}, true)

	c.Bind(securitycontract.ServiceIdentityKey, func(c runtimecontract.Container) (any, error) {
		return &securitycontract.ServiceIdentity{
			ServiceName: "local",
		}, nil
	}, true)

	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }

var ErrNoopAuth = errors.New("serviceauth: noop mode, authentication not available in monolith")

type noopAuthenticator struct{}

func (a *noopAuthenticator) Authenticate(ctx context.Context) (*securitycontract.ServiceIdentity, error) {
	_ = ctx
	return &securitycontract.ServiceIdentity{
		ServiceID:   "local",
		ServiceName: "local",
		Namespace:   "default",
	}, nil
}

func (a *noopAuthenticator) GenerateToken(ctx context.Context, targetService string) (string, error) {
	_, _ = ctx, targetService
	return "noop-token", nil
}

func (a *noopAuthenticator) VerifyToken(ctx context.Context, token string) (*securitycontract.ServiceIdentity, error) {
	_, _ = ctx, token
	return a.Authenticate(ctx)
}

func (a *noopAuthenticator) AuthenticatePeerCertificate(ctx context.Context, cert *x509.Certificate) (*securitycontract.ServiceIdentity, error) {
	_, _ = ctx, cert
	return a.Authenticate(ctx)
}
