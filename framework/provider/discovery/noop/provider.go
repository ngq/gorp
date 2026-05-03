package noop

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop 服务发现实现。
//
// 中文说明：
// - 单服务 / 单体场景默认使用此 provider；
// - 所有注册、注销、发现操作均为空实现；
// - 需要真实服务发现时，注册 contrib/registry/* 中的实现替换本 provider。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string      { return "discovery.noop" }
func (p *Provider) IsDefer() bool     { return true }
func (p *Provider) Provides() []string { return []string{contract.RPCRegistryKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RPCRegistryKey, func(c contract.Container) (any, error) {
		return &noopRegistry{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(contract.Container) error { return nil }

// noopRegistry 是 contract.ServiceRegistry 的空实现。
type noopRegistry struct{}

func (r *noopRegistry) Register(_ context.Context, _, _ string, _ map[string]string) error {
	return nil
}

func (r *noopRegistry) Deregister(_ context.Context, _, _ string) error {
	return nil
}

func (r *noopRegistry) Discover(_ context.Context, _ string) ([]contract.ServiceInstance, error) {
	return nil, nil
}

func (r *noopRegistry) Close() error { return nil }
