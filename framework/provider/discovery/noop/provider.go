package noop

import (
	"context"
	"errors"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop 服务发现实现。
//
// 中文说明：
// - 单体项目默认使用此 provider；
// - 不引入任何外部依赖（无 Consul/etcd/Nacos）；
// - 所有注册/发现操作返回空结果，单体项目无需服务发现。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "discovery.noop" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.RPCRegistryKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RPCRegistryKey, func(c contract.Container) (any, error) {
		return &noopRegistry{}, nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// ErrNoopDiscovery 表示 noop 服务发现不支持服务注册/发现。
var ErrNoopDiscovery = errors.New("discovery: noop mode, service registry not available in monolith")

// noopRegistry 是 ServiceRegistry 的空实现。
//
// 中文说明：
// - Register/Deregister 空操作，单体项目无需注册服务；
// - Discover 返回空列表，表示无远程服务可用；
// - 单体项目应直接调用本地服务，不应触发服务发现。
type noopRegistry struct{}

// Register 注册服务实例（空操作）。
//
// 中文说明：
// - 单体项目所有服务都在同一进程内，无需注册；
// - 调用此方法会被忽略，返回 nil。
func (r *noopRegistry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	// 空操作，忽略注册请求
	return nil
}

// Deregister 注销服务实例（空操作）。
//
// 中文说明：
// - 单体项目无需注销服务。
func (r *noopRegistry) Deregister(ctx context.Context, name, addr string) error {
	// 空操作，忽略注销请求
	return nil
}

// Discover 发现服务实例。
//
// 中文说明：
// - 返回空列表，表示无远程服务实例；
// - 单体项目应使用本地调用，不应依赖服务发现。
func (r *noopRegistry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	// 返回空列表，单体项目无远程服务
	return []contract.ServiceInstance{}, nil
}

// Close 关闭注册中心连接（空操作）。
func (r *noopRegistry) Close() error { return nil }