package runtime

import (
	"os"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "orm.runtime" }

func (p *Provider) IsDefer() bool { return false }

func (p *Provider) Provides() []string {
	return []string{
		datacontract.ORMBackendKey,
		datacontract.DBRuntimeKey,
		datacontract.MigratorKey,
		datacontract.SQLExecutorKey,
	}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ORMBackendKey, func(c runtimecontract.Container) (any, error) {
		cfgAny, err := c.Make(datacontract.ConfigKey)
		if err != nil {
			return string(datacontract.RuntimeBackendGorm), nil
		}
		cfg, ok := cfgAny.(datacontract.Config)
		if !ok {
			return string(datacontract.RuntimeBackendGorm), nil
		}

		var dbc datacontract.DBConfig
		if err := cfg.Unmarshal("database", &dbc); err != nil {
			return string(datacontract.RuntimeBackendGorm), nil
		}
		return string(datacontract.NormalizeBackendName(dbc.Backend)), nil
	}, true)

	c.Bind(datacontract.DBRuntimeKey, func(c runtimecontract.Container) (any, error) {
		backendAny, err := c.Make(datacontract.ORMBackendKey)
		if err != nil {
			return c.Make(datacontract.GormKey)
		}
		switch datacontract.NormalizeBackendName(backendAny.(string)) {
		case datacontract.RuntimeBackendSQLX:
			return c.Make(datacontract.SQLXKey)
		case datacontract.RuntimeBackendEnt:
			return c.Make(datacontract.EntClientKey)
		case datacontract.RuntimeBackendGorm:
			fallthrough
		default:
			return c.Make(datacontract.GormKey)
		}
	}, true)

	c.Bind(datacontract.MigratorKey, func(c runtimecontract.Container) (any, error) {
		backendAny, err := c.Make(datacontract.ORMBackendKey)
		if err == nil && datacontract.NormalizeBackendName(backendAny.(string)) == datacontract.RuntimeBackendEnt {
			return nil, os.ErrInvalid
		}
		return c.Make(datacontract.GormKey)
	}, true)

	c.Bind(datacontract.SQLExecutorKey, func(c runtimecontract.Container) (any, error) {
		return c.Make(datacontract.SQLXKey)
	}, true)

	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }
