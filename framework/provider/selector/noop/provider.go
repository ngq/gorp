// Package noop provides a no-op load balancing selector for monolith scenarios.
// This selector picks the first healthy instance or returns ErrNoAvailable.
// Note: Use in monolith where all services are local.
//
// 空负载均衡选择器实现包，用于单体应用场景。
// 此选择器选择第一个健康实例或返回 ErrNoAvailable。
// 注意：用于单体应用，所有服务都是本地的。
package noop

import (
	"context"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Provider registers no-op selector contracts.
//
// Provider 注册空选择器契约。
type Provider struct{}

// NewProvider creates a new no-op selector provider instance.
//
// NewProvider 创建新的空选择器 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "selector.noop".
//
// Name 返回 Provider 名称 "selector.noop"。
func (p *Provider) Name() string { return "selector.noop" }

// IsDefer returns true, selector can be deferred until first use.
//
// IsDefer 返回 true，选择器可延迟初始化直到首次使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the selector contract keys.
//
// Provides 返回选择器契约键列表。
func (p *Provider) Provides() []string {
	return []string{discoverycontract.SelectorKey, discoverycontract.SelectorBuilderKey}
}

// DependsOn returns the keys this provider depends on.
// Noop selector has no dependencies.
//
// DependsOn 返回该 provider 依赖的 key。
// Noop selector 无依赖。
func (p *Provider) DependsOn() []string { return nil }

// Register binds the no-op selector to the container.
//
// Register 将空选择器绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(discoverycontract.SelectorBuilderKey, func(c runtimecontract.Container) (any, error) {
		return &noopBuilder{}, nil
	}, true)

	c.Bind(discoverycontract.SelectorKey, func(c runtimecontract.Container) (any, error) {
		builder := &noopBuilder{}
		return builder.Build(), nil
	}, true)

	return nil
}

// Boot is a no-op for this provider.
//
// Boot 此 Provider 无启动逻辑。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// noopBuilder builds no-op selector instances.
//
// noopBuilder 构建空选择器实例。
type noopBuilder struct{}

// Build creates a no-op selector.
//
// Build 创建空选择器。
func (b *noopBuilder) Build() discoverycontract.Selector {
	return &noopSelector{}
}

// noopSelector implements discoverycontract.Selector with simple first-match behavior.
//
// noopSelector 使用简单的首次匹配行为实现 discoverycontract.Selector 接口。
type noopSelector struct{}

// Select picks the first healthy instance or returns ErrNoAvailable.
// Core logic: Return forced instance if specified, otherwise pick first healthy.
//
// Select 选择第一个健康实例或返回 ErrNoAvailable。
// 核心逻辑：如果指定强制实例则返回，否则选择第一个健康的实例。
func (s *noopSelector) Select(ctx context.Context, instances []transportcontract.ServiceInstance, opts ...discoverycontract.SelectOption) (
	selected transportcontract.ServiceInstance, done discoverycontract.DoneFunc, err error,
) {
	options := &discoverycontract.SelectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.ForceInstance != nil {
		return *options.ForceInstance, noopDone, nil
	}

	for _, instance := range instances {
		if instance.Healthy {
			return instance, noopDone, nil
		}
	}

	return transportcontract.ServiceInstance{}, noopDone, discoverycontract.ErrNoAvailable
}

// noopDone is a no-op DoneFunc.
//
// noopDone 是空 DoneFunc。
func noopDone(ctx context.Context, info discoverycontract.DoneInfo) {}