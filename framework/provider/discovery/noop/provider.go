// Package noop provides a no-op service discovery registry for monolith scenarios.
// This registry does not register or discover services.
// Use in monolith applications where all services are local.
//
// 服务发现注册中心实现包，用于单体应用场景。
// 此注册中心不注册或发现服务。
// 用于单体应用，所有服务都是本地的。
package noop

import (
	"context"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Provider registers a no-op service discovery contract.
//
// Provider 注册空服务发现契约。
type Provider struct{}

// NewProvider creates a new no-op discovery provider instance.
//
// NewProvider 创建新的空服务发现 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "discovery.noop".
//
// Name 返回 Provider 名称 "discovery.noop"。
func (p *Provider) Name() string { return "discovery.noop" }

// IsDefer returns true, discovery can be deferred until first use.
//
// IsDefer 返回 true，服务发现可延迟初始化直到首次使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the RPC registry contract key.
//
// Provides 返回 RPC 注册中心契约键。
func (p *Provider) Provides() []string { return []string{transportcontract.RPCRegistryKey} }

// Register binds the no-op registry to the container.
//
// Register 将空注册中心绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCRegistryKey, func(c runtimecontract.Container) (any, error) {
		return &noopRegistry{}, nil
	}, true)
	return nil
}

// Boot is a no-op for this provider.
//
// Boot 此 Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// noopRegistry implements transportcontract.RPCRegistry with no-op behavior.
//
// noopRegistry 使用空行为实现 transportcontract.RPCRegistry 接口。
type noopRegistry struct{}

// Register does nothing and returns nil.
//
// Register 不执行任何操作并返回 nil。
func (r *noopRegistry) Register(_ context.Context, _, _ string, _ map[string]string) error {
	return nil
}

// Deregister does nothing and returns nil.
//
// Deregister 不执行任何操作并返回 nil。
func (r *noopRegistry) Deregister(_ context.Context, _, _ string) error {
	return nil
}

// Discover returns nil (no instances found).
//
// Discover 返回 nil（未找到实例）。
func (r *noopRegistry) Discover(_ context.Context, _ string) ([]transportcontract.ServiceInstance, error) {
	return nil, nil
}

// Close does nothing and returns nil.
//
// Close 不执行任何操作并返回 nil。
func (r *noopRegistry) Close() error { return nil }