package noop

import (
	"context"
	"errors"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "rpc.noop" }

func (p *Provider) IsDefer() bool { return true }

func (p *Provider) Provides() []string {
	return []string{
		transportcontract.RPCClientKey,
		transportcontract.RPCServerKey,
		transportcontract.RPCRegistryKey,
	}
}

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

func (p *Provider) Boot(runtimecontract.Container) error { return nil }

type noopClient struct{}

var ErrNoopRPC = errors.New("rpc: noop mode, service-to-service call not available in monolith")

func (c *noopClient) Call(ctx context.Context, service, method string, req, resp any) error {
	_, _, _, _, _ = ctx, service, method, req, resp
	return ErrNoopRPC
}

func (c *noopClient) CallRaw(ctx context.Context, service, method string, data []byte) ([]byte, error) {
	_, _, _, _ = ctx, service, method, data
	return nil, ErrNoopRPC
}

func (c *noopClient) Close() error { return nil }

type noopServer struct{}

func (s *noopServer) Register(service string, handler any) error {
	_, _ = service, handler
	return nil
}

func (s *noopServer) Start(ctx context.Context) error {
	_ = ctx
	return ErrNoopRPC
}

func (s *noopServer) Stop(ctx context.Context) error {
	_ = ctx
	return nil
}

func (s *noopServer) Addr() string { return "" }

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
