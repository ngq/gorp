// Package dtmsdk provides DTM distributed transaction provider for the gorp framework.
// This provider registers DTM client capability with lightweight HTTP-based adapter.
//
// 本包提供 gorp 框架 DTM 分布式事务 provider。
// 本 provider 注册 DTM 客户端能力，使用轻量 HTTP 适配器。
package dtmsdk

import (
	"errors"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider implements runtimecontract.ServiceProvider for DTM client.
//
// Provider 实现 DTM 客户端的 runtimecontract.ServiceProvider。
type Provider struct{}

// NewProvider creates a new DTM provider instance.
//
// NewProvider 创建新的 DTM provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider identifier "dtm.sdk".
//
// Name 返回 provider 标识符 "dtm.sdk"。
func (p *Provider) Name() string  { return "dtm.sdk" }

// IsDefer returns true for lazy initialization.
//
// IsDefer 返回 true，延迟初始化。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the contract keys this provider satisfies.
//
// Provides 返回此 provider 满足的契约键。
func (p *Provider) Provides() []string {
	return []string{integrationcontract.DTMKey}
}

// DependsOn returns the keys this provider depends on.
// DTM provider depends on Config for DTM configuration.
//
// DependsOn 返回该 provider 依赖的 key。
// DTM provider 依赖 Config 获取 DTM 配置。
func (p *Provider) DependsOn() []string { return []string{datacontract.ConfigKey} }

// Register binds the DTM client to the container.
//
// Register 将 DTM 客户端绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.DTMKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getDTMConfig(c)
		if err != nil {
			return nil, err
		}
		return NewDTMClient(cfg)
	}, true)
	return nil
}

// Boot does nothing for lazy providers.
//
// Boot 延迟 provider 不需要 boot 操作。
func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

// getDTMConfig extracts DTM configuration from the container's config binding.
//
// getDTMConfig 从容器的 config binding 中提取 DTM 配置。
func getDTMConfig(c runtimecontract.Container) (*integrationcontract.DTMConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("dtm: invalid config service")
	}

	dtmCfg := &integrationcontract.DTMConfig{
		Enabled:         true,
		Endpoint:        "http://localhost:36789",
		Timeout:         10,
		RetryCount:      3,
		RetryInterval:   5,
		CallbackPort:    8080,
		CallbackAddress: "localhost",
	}
	if endpoint := cfg.GetString("dtm.endpoint"); endpoint != "" {
		dtmCfg.Endpoint = endpoint
	}
	if enabled := cfg.GetBool("dtm.enabled"); !enabled {
		dtmCfg.Enabled = false
	}
	if timeout := cfg.GetInt("dtm.timeout"); timeout > 0 {
		dtmCfg.Timeout = timeout
	}
	if port := cfg.GetInt("dtm.callback_port"); port > 0 {
		dtmCfg.CallbackPort = port
	}
	if addr := cfg.GetString("dtm.callback_address"); addr != "" {
		dtmCfg.CallbackAddress = addr
	}
	if retryCount := cfg.GetInt("dtm.retry_count"); retryCount > 0 {
		dtmCfg.RetryCount = retryCount
	}
	if retryInterval := cfg.GetInt("dtm.retry_interval"); retryInterval > 0 {
		dtmCfg.RetryInterval = retryInterval
	}
	return dtmCfg, nil
}