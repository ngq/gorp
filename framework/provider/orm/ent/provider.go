package ent

import (
	"fmt"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "orm.ent" }
func (p *Provider) IsDefer() bool {
	return false
}
func (p *Provider) Provides() []string { return []string{datacontract.EntClientKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.EntClientKey, func(c runtimecontract.Container) (any, error) {
		factoryAny, err := c.Make(datacontract.EntClientFactoryKey)
		if err != nil {
			cfgAny, cfgErr := c.Make(datacontract.ConfigKey)
			if cfgErr != nil {
				return nil, err
			}
			cfg := cfgAny.(datacontract.Config)
			var dbc datacontract.DBConfig
			_ = cfg.Unmarshal("database", &dbc)
			return nil, fmt.Errorf("database.backend=ent is selected, but no project-level ent factory is bound at %q (driver=%s)", datacontract.EntClientFactoryKey, dbc.Driver)
		}

		factory, ok := factoryAny.(datacontract.EntClientFactory)
		if !ok {
			return nil, fmt.Errorf("resolved %q does not implement contract.EntClientFactory", datacontract.EntClientFactoryKey)
		}
		return factory.CreateEntClient(c)
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }
