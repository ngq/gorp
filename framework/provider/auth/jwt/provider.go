package jwt

import (
	"errors"
	"strings"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "auth.jwt" }

func (p *Provider) IsDefer() bool { return true }

func (p *Provider) Provides() []string {
	return []string{securitycontract.AuthJWTKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(securitycontract.AuthJWTKey, func(c runtimecontract.Container) (any, error) {
		cfgAny, err := c.Make(datacontract.ConfigKey)
		if err != nil {
			return NewJWTService("default-secret-change-in-production", "gorp", ""), nil
		}

		cfg, ok := cfgAny.(datacontract.Config)
		if !ok {
			return nil, errors.New("auth.jwt: invalid config service")
		}

		secret := JWTSecretFromConfig(cfg)
		if secret == "" {
			secret = "default-secret-change-in-production"
		}

		issuer := strings.TrimSpace(cfg.GetString("auth.jwt.issuer"))
		if issuer == "" {
			issuer = cfg.GetString("service.name")
		}

		audience := strings.TrimSpace(cfg.GetString("auth.jwt.audience"))
		return NewJWTService(secret, issuer, audience), nil
	}, true)

	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}
