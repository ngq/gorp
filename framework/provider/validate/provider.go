package validate

import (
	"errors"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "validate" }

func (p *Provider) IsDefer() bool { return true }

func (p *Provider) Provides() []string { return []string{datacontract.ValidatorKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ValidatorKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getValidateConfig(c)
		if err != nil {
			return nil, err
		}
		return NewValidatorService(cfg)
	}, true)

	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }

func getValidateConfig(c runtimecontract.Container) (*datacontract.ValidatorConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return &datacontract.ValidatorConfig{
			Enabled: true,
			Locale:  "zh",
		}, nil
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("validate: invalid config service")
	}

	vCfg := &datacontract.ValidatorConfig{
		Enabled: true,
		Locale:  "zh",
	}

	if v := cfg.Get("validation.enabled"); v != nil {
		vCfg.Enabled = cfg.GetBool("validation.enabled")
	}
	if v := cfg.Get("validation.locale"); v != nil {
		vCfg.Locale = cfg.GetString("validation.locale")
	}

	return vCfg, nil
}
