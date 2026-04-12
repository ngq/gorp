package retry

import (
	"errors"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 Retry 实现。
//
// 中文说明：
// - 实现指数退避重试策略；
// - 支持可重试错误判断；
// - 可与 CircuitBreaker 组合使用。
type Provider struct{}

// NewProvider 创建 Retry Provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 Provider 名称。
func (p *Provider) Name() string { return "retry" }

// IsDefer 返回是否延迟加载。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回提供的服务 key。
func (p *Provider) Provides() []string {
	return []string{contract.RetryKey}
}

// Register 注册 Retry 服务。
//
// 中文说明：
// - 从容器获取配置，创建 RetryService；
// - 配置格式见 RetryConfig 结构体。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RetryKey, func(c contract.Container) (any, error) {
		cfg, err := getRetryConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRetryService(cfg), nil
	}, true)

	return nil
}

// Boot 启动 Provider。
func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// getRetryConfig 从容器获取重试配置。
//
// 中文说明：
// - 配置路径：retry.enabled、retry.default_policy 等；
// - 未配置时使用默认值。
func getRetryConfig(c contract.Container) (*contract.RetryConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		// 配置服务不可用时使用默认配置
		return &contract.RetryConfig{
			Enabled:        true,
			Strategy:       "exponential",
			DefaultPolicy:  contract.DefaultRetryPolicy(),
		}, nil
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("retry: invalid config service")
	}

	retryCfg := &contract.RetryConfig{
		Enabled:        true,
		Strategy:       "exponential",
		DefaultPolicy:  contract.DefaultRetryPolicy(),
	}

	// 读取配置项
	if v := cfg.Get("retry.enabled"); v != nil {
		retryCfg.Enabled = cfg.GetBool("retry.enabled")
	}
	if v := cfg.Get("retry.strategy"); v != nil {
		retryCfg.Strategy = cfg.GetString("retry.strategy")
	}

	// 读取默认策略配置
	if v := cfg.Get("retry.default_policy.max_attempts"); v != nil {
		retryCfg.DefaultPolicy.MaxAttempts = cfg.GetInt("retry.default_policy.max_attempts")
	}
	if v := cfg.Get("retry.default_policy.initial_delay_ms"); v != nil {
		// 读取毫秒数并转换为 Duration
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