package noop

import (
	"context"
	"errors"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop RPC 实现。
//
// 中文说明：
// - 单体项目默认使用此 provider；
// - 不引入任何外部依赖（无 Consul/etcd/gRPC）；
// - RPCClient/RPCServer/Registry 均返回空实现，调用时返回错误。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "rpc.noop" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.RPCClientKey, contract.RPCServerKey, contract.RPCRegistryKey}
}

func (p *Provider) Register(c contract.Container) error {
	// 注册 noop RPCClient
	c.Bind(contract.RPCClientKey, func(c contract.Container) (any, error) {
		return &noopClient{}, nil
	}, true)

	// 注册 noop RPCServer
	c.Bind(contract.RPCServerKey, func(c contract.Container) (any, error) {
		return &noopServer{}, nil
	}, true)

	// 注册 noop Registry
	c.Bind(contract.RPCRegistryKey, func(c contract.Container) (any, error) {
		return &noopRegistry{}, nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// noopClient 是 RPCClient 的空实现。
//
// 中文说明：
// - 所有调用都返回 ErrNoopRPC；
// - 单体项目不应该有跨服务调用，如果触发说明架构设计有问题。
type noopClient struct{}

var ErrNoopRPC = errors.New("rpc: noop mode, service-to-service call not available in monolith")

func (c *noopClient) Call(ctx context.Context, service, method string, req, resp any) error {
	return ErrNoopRPC
}

func (c *noopClient) CallRaw(ctx context.Context, service, method string, data []byte) ([]byte, error) {
	return nil, ErrNoopRPC
}

func (c *noopClient) Close() error { return nil }

// noopServer 是 RPCServer 的空实现。
//
// 中文说明：
// - Register 空操作，Start 返回错误；
// - 单体项目应该使用 HTTP 服务，不需要 RPC 服务端。
type noopServer struct{}

func (s *noopServer) Register(service string, handler any) error {
	return nil // 空操作，忽略注册
}

func (s *noopServer) Start(ctx context.Context) error {
	return ErrNoopRPC
}

func (s *noopServer) Stop(ctx context.Context) error { return nil }

func (s *noopServer) Addr() string { return "" }

// noopRegistry 是 ServiceRegistry 的空实现。
//
// 中文说明：
// - Register/Deregister 空操作；
// - Discover 返回空列表，表示无服务实例可用。
type noopRegistry struct{}

func (r *noopRegistry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	return nil // 空操作，忽略注册
}

func (r *noopRegistry) Deregister(ctx context.Context, name, addr string) error {
	return nil // 空操作，忽略注销
}

func (r *noopRegistry) Discover(ctx context.Context, name string) ([]contract.ServiceInstance, error) {
	return []contract.ServiceInstance{}, nil // 返回空列表
}

func (r *noopRegistry) Close() error { return nil }