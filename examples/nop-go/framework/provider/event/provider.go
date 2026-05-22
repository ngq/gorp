// Package event provides event bus provider for gorp framework.
// Registers local in-memory event bus as default implementation.
//
// 事件包提供事件总线 provider，用于 gorp 框架。
// 注册本地内存事件总线作为默认实现。
package event

import (
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers the event bus service contract.
// Core logic: Bind LocalEventBus factory to container.
//
// Provider 注册事件总线服务契约。
// 核心逻辑：将 LocalEventBus 工厂绑定到容器。
type Provider struct{}

// NewProvider creates a new event bus provider.
//
// NewProvider 创建新的事件总线 provider。
func NewProvider() *Provider { return &Provider{} }

// Name returns provider name for identification.
//
// Name 返回 provider 名称，用于标识。
func (p *Provider) Name() string { return "event" }

// IsDefer indicates event bus should not defer loading.
// Events may be needed early in application lifecycle.
//
// IsDefer 表示事件总线不应延迟加载。
// 事件可能在应用生命周期早期就需要。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the capability keys this provider exposes.
// Exposes EventKey for event bus service.
//
// Provides 返回 provider 暴露的能力键。
// 暴露 EventKey 用于事件总线服务。
func (p *Provider) Provides() []string { return []string{integrationcontract.EventKey} }

// Register binds the event bus factory to the container.
// Core logic: Create LocalEventBus, bind to container.
//
// Register 将事件总线工厂绑定到容器。
// 核心逻辑：创建 LocalEventBus、绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.EventKey, func(c runtimecontract.Container) (interface{}, error) {
		return NewLocalEventBus(), nil
	}, true)
	return nil
}

// Boot initializes the event bus provider.
// No additional startup logic required.
//
// Boot 初始化事件总线 provider。
// 无需额外启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }
