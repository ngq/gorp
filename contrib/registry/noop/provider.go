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

var ErrNoopDiscovery = errors.New("discovery: noop mode, service registry not available in monolith")

type noopRegistry struct{}

func (r *noopRegistry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	return nil
}

func (r *noopRegistry) Deregister(ctx context.Context, name, addr string) error {
	return nil
}

func (r *noopRegistry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	return []contract.ServiceInstance{}, nil
}

func (r *noopRegistry) Close() error { return nil }
