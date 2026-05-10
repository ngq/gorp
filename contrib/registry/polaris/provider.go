// Package polaris provides Polaris service registry provider for the gorp framework.
// This provider implements ServiceRegistry contract with Polaris SDK integration.
//
// 本包提供 gorp 框架 Polaris 服务注册中心 provider。
// 本 provider 实现 ServiceRegistry 契约，集成 Polaris SDK。
package polaris

import (
	"errors"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Provider 提供 Polaris 服务发现实现。
//
// 中文说明：
//   - 使用腾讯云 Polaris 实现服务注册与发现；
//   - 支持服务路由、负载均衡、熔断；
//   - 适用于腾讯云环境和私有化部署。
//   - 当前状态：部分可用
//   - 说明：已完成 P2 第一版最小注册/发现闭环，具备 Register / Deregister / Discover 与 fake client 行为测试；
//     但当前仍未覆盖完整路由规则、健康检查与 SDK 产品化语义。
type Provider struct{}

// NewProvider creates a new Polaris provider instance.
//
// NewProvider 创建新的 Polaris provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider identifier "registry.polaris".
//
// Name 返回 provider 标识符 "registry.polaris"。
func (p *Provider) Name() string       { return "registry.polaris" }

// IsDefer returns true for lazy initialization.
//
// IsDefer 返回 true，延迟初始化。
func (p *Provider) IsDefer() bool      { return true }

// Provides returns the contract keys this provider satisfies.
//
// Provides 返回此 provider 满足的契约键。
func (p *Provider) Provides() []string { return []string{transportcontract.RPCRegistryKey} }

// Register binds the Polaris registry to the container.
//
// Register 将 Polaris 注册中心绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCRegistryKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getPolarisConfig(c)
		if err != nil {
			return nil, err
		}
		return NewRegistry(cfg)
	}, true)
	return nil
}

// Boot does nothing for lazy providers.
//
// Boot 延迟 provider 不需要 boot 操作。
func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

// PolarisConfig 定义 Polaris 配置。
type PolarisConfig struct {
	Address            string
	Namespace          string
	Token              string
	ServiceName        string
	ServiceAddr        string
	ServicePort        int
	ServiceMeta        map[string]string
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
		WatchRetryInterval: 200 * time.Millisecond,
	}
	if v := cfg.Get("discovery.polaris.address"); v != nil {
		polarisCfg.Address = cfg.GetString("discovery.polaris.address")
	}
	if v := cfg.Get("discovery.polaris.namespace"); v != nil {
		polarisCfg.Namespace = cfg.GetString("discovery.polaris.namespace")
	}
	if v := cfg.Get("discovery.polaris.token"); v != nil {
		polarisCfg.Token = cfg.GetString("discovery.polaris.token")
	}
	if v := cfg.Get("discovery.polaris.watch_retry_interval_ms"); v != nil {
		if ms := cfg.GetInt("discovery.polaris.watch_retry_interval_ms"); ms > 0 {
			polarisCfg.WatchRetryInterval = time.Duration(ms) * time.Millisecond
		}
	}
	return polarisCfg, nil
}