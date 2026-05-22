// Package grpc provides the gRPC-based RPC client and server provider for the gorp framework.
// This package implements the core gRPC transport layer with full governance capabilities
// including tracing, metadata propagation, service authentication, circuit breaker and retry.
//
// Package grpc 提供 gorp 框架基于 gRPC 的 RPC 客户端和服务端 provider。
// 本包实现核心 gRPC 传输层，具备完整的治理能力，
// 包括 tracing、metadata 传播、服务认证、熔断器和重试。
package grpc

import (
	datacontract "github.com/ngq/gorp/framework/contract/data"
	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider is the gRPC RPC provider that registers client and server capabilities.
//
// Provider 是 gRPC RPC provider，注册客户端和服务端能力。
type Provider struct{}

// NewProvider creates a new gRPC provider instance.
//
// NewProvider 创建新的 gRPC provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider identifier "rpc.grpc".
//
// Name 返回 provider 标识符 "rpc.grpc"。
func (p *Provider) Name() string { return "rpc.grpc" }

// IsDefer returns true indicating this provider should be deferred during boot.
//
// IsDefer 返回 true，表示此 provider 应在 boot 阶段延迟执行。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the list of keys this provider binds to the container.
// Keys include RPCClientKey, RPCServerKey, GRPCConnFactoryKey, and GRPCServerRegistrarKey.
//
// Provides 返回此 provider 绑定到容器的键列表。
// 键包括 RPCClientKey、RPCServerKey、GRPCConnFactoryKey 和 GRPCServerRegistrarKey。
func (p *Provider) Provides() []string {
	return []string{
		transportcontract.RPCClientKey,
		transportcontract.RPCServerKey,
		transportcontract.GRPCConnFactoryKey,
		transportcontract.GRPCServerRegistrarKey,
	}
}

// DependsOn returns the keys this provider depends on.
// gRPC RPC depends on Config, Discovery, and Tracer.
//
// DependsOn 返回该 provider 依赖的 key。
// gRPC RPC 依赖 Config、Discovery 和 Tracer。
func (p *Provider) DependsOn() []string {
	return []string{datacontract.ConfigKey, transportcontract.RPCRegistryKey, observabilitycontract.TracerKey}
}

// Register binds the gRPC client factory, RPC client, server registrar, and RPC server
// to the container with singleton lifecycle.
//
// Register 将 gRPC client factory、RPC client、server registrar 和 RPC server
// 以 singleton 生命周期绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.GRPCConnFactoryKey, func(c runtimecontract.Container) (any, error) {
		cfg, _ := getGRPCConfig(c)
		return newClientFromContainer(c, cfg), nil
	}, true)

	c.Bind(transportcontract.RPCClientKey, func(c runtimecontract.Container) (any, error) {
		return c.Make(transportcontract.GRPCConnFactoryKey)
	}, true)

	c.Bind(transportcontract.GRPCServerRegistrarKey, func(c runtimecontract.Container) (any, error) {
		cfg, _ := getGRPCConfig(c)
		return NewServer(cfg, c), nil
	}, true)

	c.Bind(transportcontract.RPCServerKey, func(c runtimecontract.Container) (any, error) {
		return c.Make(transportcontract.GRPCServerRegistrarKey)
	}, true)

	return nil
}

// Boot is a no-op for this provider as all setup happens in Register.
//
// Boot 是空操作，因为所有设置都在 Register 中完成。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// getGRPCConfig extracts gRPC configuration from the container's config binding.
// Returns default config if config is not available or missing RPC settings.
//
// getGRPCConfig 从容器的 config binding 中提取 gRPC 配置。
// 如果 config 不可用或缺少 RPC 设置，返回默认配置。
func getGRPCConfig(c runtimecontract.Container) (*transportcontract.RPCConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return &transportcontract.RPCConfig{Mode: "grpc"}, nil
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return &transportcontract.RPCConfig{Mode: "grpc"}, nil
	}

	rpcCfg := &transportcontract.RPCConfig{
		Mode:      "grpc",
		Insecure:  true,
		TimeoutMS: 30000,
	}

	if mode := configprovider.GetStringAny(cfg, "rpc.mode"); mode != "" {
		rpcCfg.Mode = mode
	}
	if target := configprovider.GetStringAny(cfg, "rpc.grpc.target", "rpc.target"); target != "" {
		rpcCfg.Target = target
	}
	if insecureFlag, ok := configprovider.GetBoolAny(cfg, "rpc.grpc.insecure"); ok {
		rpcCfg.Insecure = insecureFlag
	}
	if timeout := configprovider.GetIntAny(cfg, "rpc.timeout_ms", "rpc.timeout"); timeout > 0 {
		rpcCfg.TimeoutMS = timeout
	}
	if addr := configprovider.GetStringAny(cfg, "rpc.grpc.address", "rpc.address"); addr != "" {
		rpcCfg.Address = addr
	}

	return rpcCfg, nil
}

// newClientFromContainer creates a gRPC Client instance by resolving dependencies from the container.
// Dependencies include service registry, selector, metadata propagator, service auth,
// tracer, circuit breaker, and retry policy.
//
// newClientFromContainer 通过从容器解析依赖创建 gRPC Client 实例。
// 依赖包括服务注册、选择器、metadata propagator、服务认证、
// tracer、熔断器和重试策略。
func newClientFromContainer(c runtimecontract.Container, cfg *transportcontract.RPCConfig) *Client {
	var registry transportcontract.ServiceRegistry
	if c.IsBind(transportcontract.RPCRegistryKey) {
		regAny, _ := c.Make(transportcontract.RPCRegistryKey)
		registry, _ = regAny.(transportcontract.ServiceRegistry)
	}

	var selector discoverycontract.Selector
	if c.IsBind(discoverycontract.SelectorKey) {
		selAny, _ := c.Make(discoverycontract.SelectorKey)
		selector, _ = selAny.(discoverycontract.Selector)
	}

	var metadataPropagator transportcontract.MetadataPropagator
	if c.IsBind(transportcontract.MetadataPropagatorKey) {
		mdAny, _ := c.Make(transportcontract.MetadataPropagatorKey)
		metadataPropagator, _ = mdAny.(transportcontract.MetadataPropagator)
	}

	var serviceAuth securitycontract.ServiceTokenIssuer
	if c.IsBind(securitycontract.ServiceAuthKey) {
		authAny, _ := c.Make(securitycontract.ServiceAuthKey)
		serviceAuth, _ = authAny.(securitycontract.ServiceTokenIssuer)
	}

	var tracer observabilitycontract.Tracer
	if c.IsBind(observabilitycontract.TracerKey) {
		tracerAny, _ := c.Make(observabilitycontract.TracerKey)
		tracer, _ = tracerAny.(observabilitycontract.Tracer)
	}

	var circuitBreaker resiliencecontract.CircuitBreaker
	if c.IsBind(resiliencecontract.CircuitBreakerKey) {
		cbAny, _ := c.Make(resiliencecontract.CircuitBreakerKey)
		circuitBreaker, _ = cbAny.(resiliencecontract.CircuitBreaker)
	}
	var retry resiliencecontract.Retry
	if c.IsBind(resiliencecontract.RetryKey) {
		retryAny, _ := c.Make(resiliencecontract.RetryKey)
		retry, _ = retryAny.(resiliencecontract.Retry)
	}

	return NewClient(cfg, registry, selector, metadataPropagator, serviceAuth, tracer, circuitBreaker, retry)
}
