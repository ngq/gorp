// Package noop provides a no-op RPC client/server for monolith scenarios.
// This RPC implementation returns ErrNoopRPC for all Call operations.
// Note: Service-to-service calls are not available in monolith mode.
//
// 空 RPC 客户端/服务器实现包，用于单体应用场景。
// 此 RPC 实现对所有 Call 操作返回 ErrNoopRPC 错误。
// 注意：服务间调用在单体模式下不可用。
package noop

import (
	"context"
	"errors"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Provider registers no-op RPC client/server contracts.
//
// Provider 注册空 RPC 客户端/服务器契约。
type Provider struct{}

// NewProvider creates a new no-op RPC provider instance.
//
// NewProvider 创建新的空 RPC Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "rpc.noop".
//
// Name 返回 Provider 名称 "rpc.noop"。
func (p *Provider) Name() string { return "rpc.noop" }

// IsDefer returns true, RPC can be deferred until first use.
//
// IsDefer 返回 true，RPC 可延迟初始化直到首次使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the RPC contract keys.
//
// Provides 返回 RPC 契约键列表。
func (p *Provider) Provides() []string {
	return []string{
		transportcontract.RPCClientKey,
		transportcontract.RPCServerKey,
		transportcontract.RPCRegistryKey,
	}
}

// DependsOn returns the keys this provider depends on.
// Noop RPC has no dependencies.
//
// DependsOn 返回该 provider 依赖的 key。
// Noop RPC 无依赖。
func (p *Provider) DependsOn() []string { return nil }

// Register binds the no-op RPC components to the container.
//
// Register 将空 RPC 组件绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCClientKey, func(c runtimecontract.Container) (any, error) {
		return &noopClient{}, nil
	}, true)
	c.Bind(transportcontract.RPCServerKey, func(c runtimecontract.Container) (any, error) {
		return &noopServer{}, nil
	}, true)
	c.Bind(transportcontract.RPCRegistryKey, func(c runtimecontract.Container) (any, error) {
		return &noopRegistry{}, nil
	}, true)
	return nil
}

// Boot is a no-op for this provider.
//
// Boot 此 Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// noopClient implements transportcontract.RPCClient with no-op behavior.
//
// noopClient 使用空行为实现 transportcontract.RPCClient 接口。
type noopClient struct{}

// ErrNoopRPC indicates RPC is not available in monolith mode.
//
// ErrNoopRPC 表示 RPC 在单体模式下不可用。
var ErrNoopRPC = errors.New("rpc: noop mode, service-to-service call not available in monolith")

// Call returns ErrNoopRPC.
//
// Call 返回 ErrNoopRPC。
func (c *noopClient) Call(ctx context.Context, service, method string, req, resp any) error {
	_, _, _, _, _ = ctx, service, method, req, resp
	return ErrNoopRPC
}

// CallRaw returns ErrNoopRPC.
//
// CallRaw 返回 ErrNoopRPC。
func (c *noopClient) CallRaw(ctx context.Context, service, method string, data []byte) ([]byte, error) {
	_, _, _, _ = ctx, service, method, data
	return nil, ErrNoopRPC
}

// Close does nothing and returns nil.
//
// Close 不执行任何操作并返回 nil。
func (c *noopClient) Close() error { return nil }

// noopServer implements transportcontract.RPCServer with no-op behavior.
//
// noopServer 使用空行为实现 transportcontract.RPCServer 接口。
type noopServer struct{}

// Register does nothing and returns nil.
//
// Register 不执行任何操作并返回 nil。
func (s *noopServer) Register(service string, handler any) error {
	_, _ = service, handler
	return nil
}

// Start returns ErrNoopRPC (server cannot start).
//
// Start 返回 ErrNoopRPC（服务器无法启动）。
func (s *noopServer) Start(ctx context.Context) error {
	_ = ctx
	return ErrNoopRPC
}

// Stop does nothing and returns nil.
//
// Stop 不执行任何操作并返回 nil。
func (s *noopServer) Stop(ctx context.Context) error {
	_ = ctx
	return nil
}

// Addr returns empty string.
//
// Addr 返回空字符串。
func (s *noopServer) Addr() string { return "" }

// noopRegistry implements transportcontract.RPCRegistry with no-op behavior.
//
// noopRegistry 使用空行为实现 transportcontract.RPCRegistry 接口。
type noopRegistry struct{}

func (r *noopRegistry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	_, _, _, _ = ctx, name, addr, meta
	return nil
}

func (r *noopRegistry) Deregister(ctx context.Context, name, addr string) error {
	_, _, _ = ctx, name, addr
	return nil
}

func (r *noopRegistry) Discover(ctx context.Context, name string) ([]transportcontract.ServiceInstance, error) {
	_, _ = ctx, name
	return []transportcontract.ServiceInstance{}, nil
}

func (r *noopRegistry) Close() error { return nil }