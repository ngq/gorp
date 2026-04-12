package ent

import (
	"fmt"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 为 ent backend 提供正式 runtime 接入点。
//
// 中文说明：
// - 框架层仍然不直接依赖 ent.Client；
// - 当 `database.backend=ent` 时，这个 provider 会去解析 `contract.EntClientFactoryKey`；
// - 真实的 ent client 创建逻辑由业务项目自己实现并注入；
// - 这样外围 wiring（`EntClientKey` / `DBRuntimeKey`）已经固定，项目只需要补自己的 factory 即可。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "orm.ent" }
func (p *Provider) IsDefer() bool {
	return false
}
func (p *Provider) Provides() []string { return []string{contract.EntClientKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.EntClientKey, func(c contract.Container) (any, error) {
		factoryAny, err := c.Make(contract.EntClientFactoryKey)
		if err != nil {
			cfgAny, cfgErr := c.Make(contract.ConfigKey)
			if cfgErr != nil {
				return nil, err
			}
			cfg := cfgAny.(contract.Config)
			var dbc contract.DBConfig
			_ = cfg.Unmarshal("database", &dbc)
			return nil, fmt.Errorf("database.backend=ent is selected, but no project-level ent factory is bound at %q (driver=%s)", contract.EntClientFactoryKey, dbc.Driver)
		}

		factory, ok := factoryAny.(contract.EntClientFactory)
		if !ok {
			return nil, fmt.Errorf("resolved %q does not implement contract.EntClientFactory", contract.EntClientFactoryKey)
		}
		return factory.CreateEntClient(c)
	}, true)
	return nil
}

func (p *Provider) Boot(contract.Container) error { return nil }
