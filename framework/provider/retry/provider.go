package retry

import (
	"errors"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "retry" }

func (p *Provider) IsDefer() bool { return true }

func (p *Provider) Provides() []string { return []string{resiliencecontract.RetryKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(resiliencecontract.RetryKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getRetryConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRetryService(cfg), nil
	}, true)

	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }

func getRetryConfig(c runtimecontract.Container) (*resiliencecontract.RetryConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return &resiliencecontract.RetryConfig{
			Enabled:       true,
			Strategy:      "exponential",
			DefaultPolicy: resiliencecontract.DefaultRetryPolicy(),
		}, nil
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("retry: invalid config service")
	}

	retryCfg := &resiliencecontract.RetryConfig{
		Enabled:       true,
		Strategy:      "exponential",
		DefaultPolicy: resiliencecontract.DefaultRetryPolicy(),
	}

	if v := cfg.Get("retry.enabled"); v != nil {
		retryCfg.Enabled = cfg.GetBool("retry.enabled")
	}
	if v := cfg.Get("retry.strategy"); v != nil {
		retryCfg.Strategy = cfg.GetString("retry.strategy")
	}
	if v := cfg.Get("retry.default_policy.max_attempts"); v != nil {
		retryCfg.DefaultPolicy.MaxAttempts = cfg.GetInt("retry.default_policy.max_attempts")
	}
	if v := cfg.Get("retry.default_policy.initial_delay_ms"); v != nil {
		retryCfg.DefaultPolicy.InitialDelay = time.Duration(cfg.GetInt("retry.default_policy.initial_delay_ms")) * time.Millisecond
	}
	if v := cfg.Get("retry.default_policy.max_delay_ms"); v != nil {
		retryCfg.DefaultPolicy.MaxDelay = time.Duration(cfg.GetInt("retry.default_policy.max_delay_ms")) * time.Millisecond
	}
	if v := cfg.Get("retry.default_policy.multiplier"); v != nil {
		retryCfg.DefaultPolicy.Multiplier = cfg.GetFloat("retry.default_policy.multiplier")
	}

	return retryCfg, nil
}
