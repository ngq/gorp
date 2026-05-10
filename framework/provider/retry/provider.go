// Package retry provides retry service for gorp framework.
// Supports exponential backoff, jitter, customizable retry policies.
// Handles network errors, gRPC errors, and AppError classification.
//
// 重试包提供重试服务，用于 gorp 框架。
// 支持指数退避、抖动、可自定义重试策略。
// 处理网络错误、gRPC 错误和 AppError 分类。
package retry

import (
	"errors"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers retry service.
// Core logic: Read retry config, create RetryService, bind to container.
//
// Provider 注册重试服务。
// 核心逻辑：读取重试配置、创建 RetryService、绑定到容器。
type Provider struct{}

// NewProvider creates a new retry provider.
//
// NewProvider 创建新的重试 provider。
func NewProvider() *Provider { return &Provider{} }

// Name returns provider name for identification.
//
// Name 返回 provider 名称，用于标识。
func (p *Provider) Name() string { return "retry" }

// IsDefer indicates retry provider should defer loading.
// Retry is typically used by RPC client after initialization.
//
// IsDefer 表示重试 provider 应延迟加载。
// 重试通常在 RPC client 初始化后使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the capability keys this provider exposes.
// Exposes RetryKey for retry service.
//
// Provides 返回 provider 暴露的能力键。
// 暴露 RetryKey 用于重试服务。
func (p *Provider) Provides() []string { return []string{resiliencecontract.RetryKey} }

// Register binds the retry factory to the container.
// Core logic: Read config, create RetryService with policy, bind to container.
//
// Register 将重试工厂绑定到容器。
// 核心逻辑：读取配置、创建带策略的 RetryService、绑定到容器。
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

// Boot initializes the retry provider.
// No additional startup logic required.
//
// Boot 初始化重试 provider。
// 无需额外启动逻辑。
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
