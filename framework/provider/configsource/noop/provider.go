// Package noop provides a no-op config source for monolith scenarios.
// This config source returns empty config and does not support Watch.
// Use when config is loaded from local files only.
//
// 空配置源实现包，用于单体应用场景。
// 此配置源返回空配置，不支持 Watch 操作。
// 用于仅从本地文件加载配置的场景。
package noop

import (
	"context"
	"errors"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers a no-op config source contract.
//
// Provider 注册空配置源契约。
type Provider struct{}

// NewProvider creates a new no-op config source provider instance.
//
// NewProvider 创建新的空配置源 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "configsource.noop".
//
// Name 返回 Provider 名称 "configsource.noop"。
func (p *Provider) Name() string { return "configsource.noop" }

// IsDefer returns true, config source can be deferred until first use.
//
// IsDefer 返回 true，配置源可延迟初始化直到首次使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the config source contract key.
//
// Provides 返回配置源契约键。
func (p *Provider) Provides() []string { return []string{datacontract.ConfigSourceKey} }

// Register binds the no-op config source to the container.
//
// Register 将空配置源绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ConfigSourceKey, func(c runtimecontract.Container) (any, error) {
		return &noopConfigSource{}, nil
	}, true)
	return nil
}

// Boot is a no-op for this provider.
//
// Boot 此 Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// noopConfigSource implements datacontract.ConfigSource with no-op behavior.
//
// noopConfigSource 使用空行为实现 datacontract.ConfigSource 接口。
type noopConfigSource struct{}

// Load returns an empty config map.
//
// Load 返回空配置 map。
func (s *noopConfigSource) Load(_ context.Context) (map[string]any, error) {
	return map[string]any{}, nil
}

// Get returns nil (key not found).
//
// Get 返回 nil（键未找到）。
func (s *noopConfigSource) Get(_ context.Context, _ string) (any, error) {
	return nil, nil
}

// Set returns an error (operation not supported).
//
// Set 返回错误（操作不支持）。
func (s *noopConfigSource) Set(_ context.Context, _ string, _ any) error {
	return errors.New("configsource.noop: Set not supported")
}

// Watch returns an error (operation not supported).
//
// Watch 返回错误（操作不支持）。
func (s *noopConfigSource) Watch(_ context.Context, _ string) (datacontract.ConfigWatcher, error) {
	return nil, errors.New("configsource.noop: Watch not supported")
}

// Close does nothing and returns nil.
//
// Close 不执行任何操作并返回 nil。
func (s *noopConfigSource) Close() error { return nil }