// Package baseconfigsource provides a base config source provider template.
// Concrete ConfigSource providers embed BaseConfigSourceProvider and only supply
// provider-specific config extraction and source construction logic.
// This eliminates structural duplication across all ConfigSource providers
// and integrates them with the container's Destroy lifecycle.
//
// baseconfigsource 包提供配置源 provider 基础模板。
// 具体 ConfigSource provider 内嵌 BaseConfigSourceProvider，只需提供差异化的配置提取和源构造逻辑。
// 这消除了所有 ConfigSource provider 的结构性重复，
// 并将它们集成到容器的 Destroy 生命周期。
package baseconfigsource

import (
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// BaseConfigSourceProvider eliminates structural duplication across ConfigSource providers.
// Concrete providers embed this struct and only supply Name, GetConfig, and NewSource.
//
// BaseConfigSourceProvider 消除 ConfigSource provider 的结构性重复。
// 具体 provider 内嵌此结构体，只需提供 Name、GetConfig、NewSource 三项差异化逻辑。
type BaseConfigSourceProvider struct {
	// NameStr is the provider identifier, e.g. "configsource.apollo".
	NameStr string

	// GetConfig extracts provider-specific configuration from the container.
	GetConfig func(c runtimecontract.Container) (any, error)

	// NewSource creates a ConfigSource instance from the given config.
	NewSource func(cfg any) (datacontract.ConfigSource, error)
}

// Name returns the provider identifier.
func (p *BaseConfigSourceProvider) Name() string { return p.NameStr }

// IsDefer returns true for lazy initialization.
func (p *BaseConfigSourceProvider) IsDefer() bool { return true }

// Provides returns the contract keys this provider satisfies.
func (p *BaseConfigSourceProvider) Provides() []string {
	return []string{datacontract.ConfigSourceKey}
}

// DependsOn returns the keys this provider depends on.
// BaseConfigSourceProvider depends on Config for configuration.
//
// DependsOn 返回该 provider 依赖的 key。
// BaseConfigSourceProvider 依赖 Config 获取配置。
func (p *BaseConfigSourceProvider) DependsOn() []string {
	return []string{datacontract.ConfigKey}
}

// Register binds ConfigSource to the container as a lazy singleton.
// The source's Close method is registered with the container's Destroy lifecycle.
//
// Register 将 ConfigSource 以延迟单例形式绑定到容器。
// 源的 Close 方法注册到容器的 Destroy 生命周期。
func (p *BaseConfigSourceProvider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ConfigSourceKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := p.GetConfig(c)
		if err != nil {
			return nil, err
		}
		src, err := p.NewSource(cfg)
		if err != nil {
			return nil, err
		}
		c.RegisterCloser(datacontract.ConfigSourceKey, src)
		return src, nil
	}, true)
	return nil
}

// Boot does nothing for lazy providers.
func (p *BaseConfigSourceProvider) Boot(c runtimecontract.Container) error { return nil }
