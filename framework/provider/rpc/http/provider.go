// Package http provides HTTP RPC client and server provider for gorp framework.
// Implements RPCClient and RPCServer contracts with HTTP transport.
// Includes service discovery, metadata propagation, tracing, circuit breaker.
//
// 本包提供 HTTP RPC 客户端和服务端 provider，用于 gorp 框架。
// 实现带 HTTP 传输的 RPCClient 和 RPCServer 契约。
// 包含服务发现、元数据传播、追踪、熔断器。
package http

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

// Provider registers HTTP RPC client/server services.
// Core logic: Create Client with governance, create Server for RPC over HTTP, bind to container.
//
// Provider 注册 HTTP RPC 客户端/服务端服务。
// 核心逻辑：创建带治理的 Client、创建用于 HTTP 上 RPC 的 Server、绑定到容器。
type Provider struct{}

// NewProvider creates a new HTTP RPC provider.
//
// NewProvider 创建新的 HTTP RPC provider。
func NewProvider() *Provider { return &Provider{} }

// Name returns provider name for identification.
//
// Name 返回 provider 名称，用于标识。
func (p *Provider) Name() string  { return "rpc.http" }

// IsDefer indicates HTTP RPC should defer loading.
// RPC can be loaded after core providers.
//
// IsDefer 表示 HTTP RPC 应延迟加载。
// RPC 可以在核心 provider 之后加载。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the capability keys this provider exposes.
// Exposes RPCClientKey, RPCServerKey for HTTP RPC.
//
// Provides 返回 provider 暴露的能力键。
// 暴露 RPCClientKey、RPCServerKey 用于 HTTP RPC。
func (p *Provider) Provides() []string {
	return []string{transportcontract.RPCClientKey, transportcontract.RPCServerKey}
}

// Register binds HTTP RPC client/server factories to the container.
// Core logic: Create client with discovery/governance, create server, bind to container.
//
// Register 将 HTTP RPC 客户端/服务端工厂绑定到容器。
// 核心逻辑：创建带发现/治理的 client、创建 server、绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.RPCClientKey, func(c runtimecontract.Container) (any, error) {
		cfg, _ := getConfig(c)
		return newClientFromContainer(c, cfg), nil
	}, true)

	c.Bind(transportcontract.RPCServerKey, func(c runtimecontract.Container) (any, error) {
		cfg, _ := getConfig(c)
		return NewServer(cfg, c), nil
	}, true)

	return nil
}

// Boot initializes the HTTP RPC provider.
// No additional startup logic required.
//
// Boot 初始化 HTTP RPC provider。
// 无需额外启动逻辑。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// getConfig extracts HTTP RPC configuration from the container's config binding.
// Returns default config if config is not available or missing RPC settings.
//
// getConfig 从容器的 config binding 中提取 HTTP RPC 配置。
// 如果 config 不可用或缺少 RPC 设置，返回默认配置。
func getConfig(c runtimecontract.Container) (*transportcontract.RPCConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return &transportcontract.RPCConfig{Mode: "http"}, nil
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return &transportcontract.RPCConfig{Mode: "http"}, nil
	}

	rpcCfg := &transportcontract.RPCConfig{
		Mode:      "http",
		TimeoutMS: 30000,
	}

	if mode := configprovider.GetStringAny(cfg, "rpc.mode"); mode != "" {
		rpcCfg.Mode = mode
	}
	if baseURL := configprovider.GetStringAny(cfg, "rpc.http.base_url", "rpc.base_url"); baseURL != "" {
		rpcCfg.BaseURL = baseURL
	}
	if timeout := configprovider.GetIntAny(cfg, "rpc.timeout_ms", "rpc.timeout"); timeout > 0 {
		rpcCfg.TimeoutMS = timeout
	}

	return rpcCfg, nil
}

// newClientFromContainer creates an HTTP Client instance by resolving dependencies from the container.
// Dependencies include service registry, selector, metadata propagator, service auth,
// tracer, circuit breaker, and retry policy.
//
// newClientFromContainer 通过从容器解析依赖创建 HTTP Client 实例。
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