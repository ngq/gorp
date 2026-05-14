// Package baseregistry provides a base service registry provider template.
// Concrete registry providers embed BaseRegistryProvider and only supply
// provider-specific config extraction and registry construction logic.
// This eliminates structural duplication across all registry providers
// and integrates them with the container's Destroy lifecycle.
//
// baseregistry 包提供服务注册中心 provider 基础模板。
// 具体 registry provider 内嵌 BaseRegistryProvider，只需提供差异化的配置提取和注册中心构造逻辑。
// 这消除了所有 registry provider 的结构性重复，
// 并将它们集成到容器的 Destroy 生命周期。
package baseregistry

import (
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// BaseRegistryProvider eliminates structural duplication across registry providers.
// Concrete providers embed this struct and only supply Name, GetConfig, and NewRegistry.
//
// BaseRegistryProvider 消除 registry provider 的结构性重复。
// 具体 provider 内嵌此结构体，只需提供 Name、GetConfig、NewRegistry 三项差异化逻辑。
type BaseRegistryProvider struct {
	// NameStr is the provider identifier, e.g. "registry.consul".
	NameStr string

	// GetConfig extracts provider-specific configuration from the container.
	GetConfig func(c runtimecontract.Container) (any, error)

	// NewRegistry creates a ServiceRegistry instance from the given config.
	NewRegistry func(cfg any) (transportcontract.ServiceRegistry, error)
}

// Name returns the provider identifier.
func (p *BaseRegistryProvider) Name() string { return p.NameStr }

// IsDefer returns true for lazy initialization.
func (p *BaseRegistryProvider) IsDefer() bool { return true }

// Provides returns the contract keys this provider satisfies.
func (p *BaseRegistryProvider) Provides() []string {
	return []string{transportcontract.RPCRegistryKey}
}

// Register binds ServiceRegistry to the container as a lazy singleton.
// The registry's Close method is registered with the container's Destroy lifecycle.
//
// Register 将 ServiceRegistry 以延迟单例形式绑定到容器。
// 注册中心的 Close 方法注册到容器的 Destroy 生命周期。
func (p *BaseRegistryProvider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCRegistryKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := p.GetConfig(c)
		if err != nil {
			return nil, err
		}
		reg, err := p.NewRegistry(cfg)
		if err != nil {
			return nil, err
		}
		c.RegisterCloser(transportcontract.RPCRegistryKey, reg)
		return reg, nil
	}, true)
	return nil
}

// Boot does nothing for lazy providers.
func (p *BaseRegistryProvider) Boot(c runtimecontract.Container) error { return nil }
