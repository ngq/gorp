// Package polaris provides Polaris configuration center provider for the gorp framework.
// This provider implements ConfigSource contract with Polaris SDK integration.
//
// 本包提供 gorp 框架 Polaris 配置中心 provider。
// 本 provider 实现 ConfigSource 契约，集成 Polaris SDK。
package polaris

import (
	"errors"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider 提供 Polaris 配置中心实现。
//
// 中文说明：
//   - 使用腾讯云 Polaris 配置中心；
//   - 支持命名空间隔离；
//   - 支持配置分组管理；
//   - 支持配置热更新；
//   - 适用于腾讯云环境和私有化部署。
//   - 当前状态：部分可用
//   - 说明：已完成 P2 第一版最小 HTTP 配置闭环，具备 Load / Watch 与 fake client 行为测试；
//     但当前仍是轮询桥接态，尚未进入完整 Polaris SDK 产品化能力。
type Provider struct{}

// NewProvider creates a new Polaris provider instance.
//
// NewProvider 创建新的 Polaris provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider identifier "configsource.polaris".
//
// Name 返回 provider 标识符 "configsource.polaris"。
func (p *Provider) Name() string { return "configsource.polaris" }

// IsDefer returns true for lazy initialization.
//
// IsDefer 返回 true，延迟初始化。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the contract keys this provider satisfies.
//
// Provides 返回此 provider 满足的契约键。
func (p *Provider) Provides() []string {
	return []string{datacontract.ConfigSourceKey}
}

// Register binds the Polaris config source to the container.
//
// Register 将 Polaris 配置源绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ConfigSourceKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getPolarisConfig(c)
		if err != nil {
			return nil, err
		}
		return NewConfigSource(cfg)
	}, true)
	return nil
}

// Boot does nothing for lazy providers.
//
// Boot 延迟 provider 不需要 boot 操作。
func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

// defaultPolarisPollInterval is the default polling interval for config updates.
//
// defaultPolarisPollInterval 是配置更新的默认轮询间隔。
const defaultPolarisPollInterval = 5 * time.Second

// PolarisConfig 定义 Polaris 配置。
type PolarisConfig struct {
	ServerAddress      string
	Namespace          string
	FileGroup          string
	FileName           string
	Token              string
	PollInterval       time.Duration
	WatchRetryInterval time.Duration
}

// getPolarisConfig extracts Polaris configuration from the container's config binding.
//
// getPolarisConfig 从容器的 config binding 中提取 Polaris 配置。
func getPolarisConfig(c runtimecontract.Container) (*PolarisConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("polaris: invalid config service")
	}

	polarisCfg := &PolarisConfig{
		Namespace:          "default",
		PollInterval:       defaultPolarisPollInterval,
		WatchRetryInterval: time.Second,
	}

	if v := cfg.Get("config.polaris.server_address"); v != nil {
		polarisCfg.ServerAddress = cfg.GetString("config.polaris.server_address")
	}
	if v := cfg.Get("config.polaris.namespace"); v != nil {
		polarisCfg.Namespace = cfg.GetString("config.polaris.namespace")
	}
	if v := cfg.Get("config.polaris.file_group"); v != nil {
		polarisCfg.FileGroup = cfg.GetString("config.polaris.file_group")
	}
	if v := cfg.Get("config.polaris.file_name"); v != nil {
		polarisCfg.FileName = cfg.GetString("config.polaris.file_name")
	}
	if v := cfg.Get("config.polaris.token"); v != nil {
		polarisCfg.Token = cfg.GetString("config.polaris.token")
	}
	if v := cfg.Get("config.polaris.poll_interval_seconds"); v != nil {
		if seconds := cfg.GetInt("config.polaris.poll_interval_seconds"); seconds > 0 {
			polarisCfg.PollInterval = time.Duration(seconds) * time.Second
		}
	}
	if v := cfg.Get("config.polaris.watch_retry_interval_ms"); v != nil {
		if ms := cfg.GetInt("config.polaris.watch_retry_interval_ms"); ms > 0 {
			polarisCfg.WatchRetryInterval = time.Duration(ms) * time.Millisecond
		}
	}

	return polarisCfg, nil
}