// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file selects capability providers from runtime configuration during bootstrap.
// Keeps config-driven microservice assembly centralized and predictable.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件在 bootstrap 阶段根据运行时配置选择能力 provider。
// 将配置驱动的微服务装配逻辑集中管理并保持可预测性。
package bootstrap

import (
	"context"
	"fmt"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// SelectedMicroserviceProviders returns the config-selected microservice capability providers.
// Does NOT process governance.disable/enable/providers.* overrides.
// For governance-aware selection, use SelectedMicroserviceProvidersWithOptions.
// Note: nil providers are filtered out to prevent nil panic in RegisterProviders.
//
// SelectedMicroserviceProviders 返回由配置选择出的微服务能力 provider 集合。
// 不处理 governance.disable/enable/providers.* 覆盖。
// 需要治理覆盖的场景请使用 SelectedMicroserviceProvidersWithOptions。
// 注意：nil provider 会被过滤，防止 RegisterProviders 时 nil panic。
func SelectedMicroserviceProviders(cfg datacontract.Config) []runtimecontract.ServiceProvider {
	mode := DetectGovernanceMode(cfg)
	providers := make([]runtimecontract.ServiceProvider, 0, 12)
	for _, p := range []runtimecontract.ServiceProvider{
		SelectDiscoveryProviderWithMode(cfg, mode),
		SelectSelectorProviderWithMode(cfg, mode),
		SelectRPCProviderWithMode(cfg, mode),
		SelectTracingProviderWithMode(cfg, mode),
		SelectMetadataProviderWithMode(cfg, mode),
		SelectServiceAuthProviderWithMode(cfg, mode),
		SelectCircuitBreakerProviderWithMode(cfg, mode),
		SelectLoadSheddingProviderWithMode(cfg, mode),
		SelectRetryProviderWithMode(cfg, mode),
		SelectDTMProvider(cfg),
		SelectMessageQueueProvider(cfg),
		SelectDistributedLockProvider(cfg),
		SelectWebSocketProvider(cfg),
	} {
		if p != nil {
			providers = append(providers, p)
		}
	}
	return providers
}

// SelectedMicroserviceProvidersWithOptions returns governance-aware microservice providers.
// Processes governance.disable, governance.enable, and governance.providers.* overrides
// before selecting providers.
// Note: nil providers are filtered out to prevent nil panic in RegisterProviders.
//
// SelectedMicroserviceProvidersWithOptions 返回带治理覆盖的微服务 provider 集合。
// 选择 provider 前处理 governance.disable、governance.enable 和 governance.providers.* 覆盖。
// 注意：nil provider 会被过滤，防止 RegisterProviders 时 nil panic。
func SelectedMicroserviceProvidersWithOptions(cfg datacontract.Config, disabled, enabled []string, providerOverrides map[string]string) []runtimecontract.ServiceProvider {
	overlayCfg := overlayGovernanceConfig(cfg, disabled, enabled, providerOverrides)
	mode := DetectGovernanceMode(overlayCfg)
	providers := make([]runtimecontract.ServiceProvider, 0, 12)
	for _, p := range []runtimecontract.ServiceProvider{
		SelectDiscoveryProviderWithMode(overlayCfg, mode),
		SelectSelectorProviderWithMode(overlayCfg, mode),
		SelectRPCProviderWithMode(overlayCfg, mode),
		SelectTracingProviderWithMode(overlayCfg, mode),
		SelectMetadataProviderWithMode(overlayCfg, mode),
		SelectServiceAuthProviderWithMode(overlayCfg, mode),
		SelectCircuitBreakerProviderWithMode(overlayCfg, mode),
		SelectLoadSheddingProviderWithMode(overlayCfg, mode),
		SelectRetryProviderWithMode(overlayCfg, mode),
		SelectDTMProvider(overlayCfg),
		SelectMessageQueueProvider(overlayCfg),
		SelectDistributedLockProvider(overlayCfg),
		SelectWebSocketProvider(overlayCfg),
	} {
		if p != nil {
			providers = append(providers, p)
		}
	}
	return providers
}

// RegisterSelectedMicroserviceProviders resolves and registers microservice capability providers into the container.
//
// RegisterSelectedMicroserviceProviders 解析并将微服务能力 provider 注册到容器。
func RegisterSelectedMicroserviceProviders(c runtimecontract.Container) error {
	return RegisterSelectedMicroserviceProvidersWithMode(c, "")
}

// RegisterSelectedMicroserviceProvidersWithMode resolves and registers microservice providers with an explicit mode override.
//
// RegisterSelectedMicroserviceProvidersWithMode 在显式 mode 覆盖下解析并注册微服务 provider。
func RegisterSelectedMicroserviceProvidersWithMode(c runtimecontract.Container, modeOverride string) error {
	return registerSelectedMicroserviceProvidersWithOptions(c, modeOverride, nil, nil, nil)
}

func registerSelectedMicroserviceProvidersWithOptions(c runtimecontract.Container, modeOverride string, disabled []string, enabled []string, providerOverrides map[string]string) error {
	if c == nil || !c.IsBind(datacontract.ConfigKey) {
		return nil
	}
	cfg, err := container.MakeWith[datacontract.Config](c, datacontract.ConfigKey)
	if err != nil {
		return fmt.Errorf("bootstrap: failed to resolve config for provider selection: %w", err)
	}
	cfg = overlayGovernanceConfig(cfg, disabled, enabled, providerOverrides)

	configSourceProvider := SelectConfigSourceProvider(cfg)
	if configSourceProvider != nil {
		name := configSourceProvider.Name()
		if name != "configsource.local" && name != "configsource.noop" {
			// External config sources should be registered first so reload can refresh the effective config snapshot.
			// 外部配置源需要优先注册，这样后续 reload 才能刷新实际生效的配置快照。
			if err := c.RegisterProvider(configSourceProvider); err != nil {
				return err
			}
			if cfgSvc, ok := cfg.(datacontract.Config); ok {
				if err := cfgSvc.Reload(context.Background()); err != nil {
					return err
				}
				cfg = overlayGovernanceConfig(cfgSvc, disabled, enabled, providerOverrides)
			}
		}
	}

	mode := DetectGovernanceMode(cfg)
	if modeOverride != "" {
		mode = NormalizeGovernanceMode(resiliencecontract.GovernanceMode(modeOverride))
	}

	selectedProviders := []runtimecontract.ServiceProvider{
		SelectDiscoveryProviderWithMode(cfg, mode),
		SelectSelectorProviderWithMode(cfg, mode),
		SelectRPCProviderWithMode(cfg, mode),
		SelectTracingProviderWithMode(cfg, mode),
		SelectMetadataProviderWithMode(cfg, mode),
		SelectServiceAuthProviderWithMode(cfg, mode),
		SelectCircuitBreakerProviderWithMode(cfg, mode),
		SelectLoadSheddingProviderWithMode(cfg, mode),
		SelectDTMProvider(cfg),
		SelectMessageQueueProvider(cfg),
		SelectDistributedLockProvider(cfg),
		SelectWebSocketProvider(cfg),
	}
	for _, p := range selectedProviders {
		if p == nil {
			continue
		}
		if err := c.RegisterProvider(p); err != nil {
			return err
		}
	}
	return nil
}

// SelectConfigSourceProvider selects the config-source provider from config.
//
// SelectConfigSourceProvider 根据配置选择配置源 provider。
func SelectConfigSourceProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	typ := governanceProviderOverride(cfg, "configsource")
	if typ == "" {
		typ = getConfigString(cfg, "configsource.backend", "configsource.type", "config_source.backend", "config_source.type")
	}
	if typ == "" {
		typ = DefaultGovernanceProviderDefaults(DetectGovernanceMode(cfg)).ConfigSource
	}
	return providerFromMap(configSourceProviderFactories, typ, "local")
}

// SelectDiscoveryProvider selects the discovery provider from config.
//
// SelectDiscoveryProvider 根据配置选择服务发现 provider。
func SelectDiscoveryProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	return SelectDiscoveryProviderWithMode(cfg, DetectGovernanceMode(cfg))
}

// SelectDiscoveryProviderWithMode selects the discovery provider with an explicit governance mode.
//
// SelectDiscoveryProviderWithMode 在显式治理模式下选择服务发现 provider。
func SelectDiscoveryProviderWithMode(cfg datacontract.Config, mode resiliencecontract.GovernanceMode) runtimecontract.ServiceProvider {
	if isGovernanceCapabilityDisabled(cfg, "discovery") {
		return providerFromMap(discoveryProviderFactories, "noop", "noop")
	}
	typ := governanceProviderOverride(cfg, "discovery")
	if typ == "" {
		typ = getConfigString(cfg, "discovery.backend", "discovery.type")
	}
	if typ == "" {
		typ = DefaultGovernanceProviderDefaults(mode).Discovery
	}
	return providerFromMap(discoveryProviderFactories, typ, "noop")
}

// SelectSelectorProvider selects the selector provider from config.
//
// SelectSelectorProvider 根据配置选择 selector provider。
func SelectSelectorProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	return SelectSelectorProviderWithMode(cfg, DetectGovernanceMode(cfg))
}

// SelectSelectorProviderWithMode selects the selector provider with an explicit governance mode.
//
// SelectSelectorProviderWithMode 在显式治理模式下选择 selector provider。
func SelectSelectorProviderWithMode(cfg datacontract.Config, mode resiliencecontract.GovernanceMode) runtimecontract.ServiceProvider {
	if isGovernanceCapabilityDisabled(cfg, "selector") {
		return providerFromMap(selectorProviderFactories, "noop", "noop")
	}
	algorithm := governanceProviderOverride(cfg, "selector")
	if algorithm == "" {
		algorithm = getConfigString(cfg, "selector.backend", "selector.algorithm", "selector.type")
	}
	if algorithm == "" {
		algorithm = DefaultGovernanceProviderDefaults(mode).Selector
	}
	return providerFromMap(selectorProviderFactories, algorithm, "noop")
}

// SelectRPCProvider selects the RPC provider from config.
//
// SelectRPCProvider 根据配置选择 RPC provider。
func SelectRPCProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	return SelectRPCProviderWithMode(cfg, DetectGovernanceMode(cfg))
}

// SelectRPCProviderWithMode selects the RPC provider with an explicit governance mode.
//
// SelectRPCProviderWithMode 在显式治理模式下选择 RPC provider。
func SelectRPCProviderWithMode(cfg datacontract.Config, mode resiliencecontract.GovernanceMode) runtimecontract.ServiceProvider {
	rpcMode := governanceProviderOverride(cfg, "rpc")
	if rpcMode == "" {
		rpcMode = getConfigString(cfg, "rpc.mode")
	}
	if rpcMode == "" {
		rpcMode = DefaultGovernanceProviderDefaults(mode).RPC
	}
	return providerFromMap(rpcProviderFactories, rpcMode, "noop")
}

// SelectTracingProvider selects the tracing provider from config and governance mode.
//
// SelectTracingProvider 根据配置与治理模式选择 tracing provider。
func SelectTracingProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	return SelectTracingProviderWithMode(cfg, DetectGovernanceMode(cfg))
}

// SelectTracingProviderWithMode selects the tracing provider with an explicit governance mode.
//
// SelectTracingProviderWithMode 在显式治理模式下选择 tracing provider。
func SelectTracingProviderWithMode(cfg datacontract.Config, mode resiliencecontract.GovernanceMode) runtimecontract.ServiceProvider {
	if isGovernanceCapabilityDisabled(cfg, "tracing") {
		return providerFromMap(tracingProviderFactories, "noop", "noop")
	}
	enabled := cfg != nil && cfg.GetBool("tracing.enabled")
	backend := governanceProviderOverride(cfg, "tracing")
	if backend == "" {
		backend = getConfigString(cfg, "tracing.backend", "tracing.type")
	}
	if backend == "" && enabled {
		backend = "otel"
	}
	if backend == "" {
		backend = DefaultGovernanceProviderDefaults(mode).Tracing
	}
	return providerFromMap(tracingProviderFactories, backend, "noop")
}

// SelectMetadataProvider selects the metadata provider from config and governance mode.
//
// SelectMetadataProvider 根据配置与治理模式选择 metadata provider。
func SelectMetadataProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	return SelectMetadataProviderWithMode(cfg, DetectGovernanceMode(cfg))
}

// SelectMetadataProviderWithMode selects the metadata provider with an explicit governance mode.
//
// SelectMetadataProviderWithMode 在显式治理模式下选择 metadata provider。
func SelectMetadataProviderWithMode(cfg datacontract.Config, mode resiliencecontract.GovernanceMode) runtimecontract.ServiceProvider {
	if isGovernanceCapabilityDisabled(cfg, "metadata") {
		return providerFromMap(metadataProviderFactories, "noop", "noop")
	}
	backend := governanceProviderOverride(cfg, "metadata")
	if backend == "" {
		backend = getConfigString(cfg, "metadata.mode", "metadata.backend")
	}
	enabled := cfg != nil && (cfg.Get("metadata.propagate_prefix") != nil || cfg.GetBool("metadata.enabled"))
	if backend == "" && enabled {
		backend = "default"
	}
	if backend == "" {
		backend = DefaultGovernanceProviderDefaults(mode).Metadata
	}
	return providerFromMap(metadataProviderFactories, backend, "noop")
}

// SelectServiceAuthProvider selects the service-auth provider from config and governance mode.
//
// SelectServiceAuthProvider 根据配置与治理模式选择服务间认证 provider。
func SelectServiceAuthProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	return SelectServiceAuthProviderWithMode(cfg, DetectGovernanceMode(cfg))
}

// SelectServiceAuthProviderWithMode selects the service-auth provider with an explicit governance mode.
//
// SelectServiceAuthProviderWithMode 在显式治理模式下选择服务间认证 provider。
func SelectServiceAuthProviderWithMode(cfg datacontract.Config, mode resiliencecontract.GovernanceMode) runtimecontract.ServiceProvider {
	if isGovernanceCapabilityDisabled(cfg, "serviceauth") {
		return providerFromMap(serviceAuthProviderFactories, "noop", "noop")
	}
	backend := governanceProviderOverride(cfg, "serviceauth")
	if backend == "" {
		backend = getConfigString(cfg, "service_auth.backend", "service_auth.mode")
	}
	enabled := cfg != nil && cfg.GetBool("service_auth.enabled")
	if backend == "" && enabled {
		backend = "token"
	}
	if backend == "" {
		backend = DefaultGovernanceProviderDefaults(mode).ServiceAuth
	}
	return providerFromMap(serviceAuthProviderFactories, backend, "noop")
}

// SelectCircuitBreakerProvider selects the circuit-breaker provider from config and governance mode.
//
// SelectCircuitBreakerProvider 根据配置与治理模式选择熔断 provider。
func SelectCircuitBreakerProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	return SelectCircuitBreakerProviderWithMode(cfg, DetectGovernanceMode(cfg))
}

// SelectCircuitBreakerProviderWithMode selects the circuit-breaker provider with an explicit governance mode.
//
// SelectCircuitBreakerProviderWithMode 在显式治理模式下选择熔断 provider。
func SelectCircuitBreakerProviderWithMode(cfg datacontract.Config, mode resiliencecontract.GovernanceMode) runtimecontract.ServiceProvider {
	if isGovernanceCapabilityDisabled(cfg, "circuitbreaker") {
		return providerFromMap(circuitBreakerProviderFactories, "noop", "noop")
	}
	backend := governanceProviderOverride(cfg, "circuitbreaker")
	if backend == "" {
		backend = getConfigString(cfg, "circuit_breaker.backend", "circuit_breaker.type")
	}
	enabled := cfg != nil && cfg.GetBool("circuit_breaker.enabled")
	if backend == "" && enabled {
		backend = "sentinel"
	}
	if backend == "" {
		backend = DefaultGovernanceProviderDefaults(mode).CircuitBreaker
	}
	return providerFromMap(circuitBreakerProviderFactories, backend, "noop")
}

// SelectLoadSheddingProvider selects the load-shedding provider from config and governance mode.
//
// SelectLoadSheddingProvider 根据配置与治理模式选择过载保护 provider。
func SelectLoadSheddingProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	return SelectLoadSheddingProviderWithMode(cfg, DetectGovernanceMode(cfg))
}

// SelectLoadSheddingProviderWithMode selects the load-shedding provider with an explicit governance mode.
//
// SelectLoadSheddingProviderWithMode 在显式治理模式下选择过载保护 provider。
func SelectLoadSheddingProviderWithMode(cfg datacontract.Config, mode resiliencecontract.GovernanceMode) runtimecontract.ServiceProvider {
	if isGovernanceCapabilityDisabled(cfg, "loadshedding") {
		return providerFromMap(loadShedderProviderFactories, "noop", "noop")
	}
	backend := governanceProviderOverride(cfg, "loadshedding")
	if backend == "" {
		backend = getConfigString(cfg, "load_shedding.backend", "load_shedding.type")
	}
	enabled := cfg != nil && cfg.GetBool("load_shedding.enabled")
	if backend == "" && enabled {
		backend = "semaphore"
	}
	if backend == "" {
		backend = DefaultGovernanceProviderDefaults(mode).LoadShedder
	}
	return providerFromMap(loadShedderProviderFactories, backend, "noop")
}

// SelectRetryProvider selects the retry provider from config.
//
// SelectRetryProvider 根据配置选择重试 provider。
func SelectRetryProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	return SelectRetryProviderWithMode(cfg, DetectGovernanceMode(cfg))
}

// SelectRetryProviderWithMode selects the retry provider with an explicit governance mode.
//
// SelectRetryProviderWithMode 在显式治理模式下选择重试 provider。
func SelectRetryProviderWithMode(cfg datacontract.Config, mode resiliencecontract.GovernanceMode) runtimecontract.ServiceProvider {
	if isGovernanceCapabilityDisabled(cfg, "retry") {
		return providerFromMap(retryProviderFactories, "noop", "noop")
	}
	backend := governanceProviderOverride(cfg, "retry")
	if backend == "" {
		backend = getConfigString(cfg, "retry.backend", "retry.type")
	}
	enabled := cfg != nil && cfg.GetBool("retry.enabled")
	if backend == "" && enabled {
		backend = "default"
	}
	if backend == "" {
		backend = DefaultGovernanceProviderDefaults(mode).Retry
	}
	return providerFromMap(retryProviderFactories, backend, "noop")
}

// SelectDTMProvider selects the DTM provider from config and enable flags.
//
// SelectDTMProvider 根据配置与开关状态选择 DTM provider。
func SelectDTMProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	backend := governanceProviderOverride(cfg, "dtm")
	if backend == "" {
		backend = getConfigString(cfg, "dtm.backend", "dtm.type", "dtm.driver")
	}
	enabled := cfg != nil && cfg.GetBool("dtm.enabled")
	if backend == "" && enabled {
		backend = "dtmsdk"
	}
	if backend == "" {
		backend = DefaultGovernanceProviderDefaults(DetectGovernanceMode(cfg)).DTM
	}
	return providerFromMap(dtmProviderFactories, backend, "noop")
}

// SelectMessageQueueProvider selects the message-queue provider from config and enable flags.
//
// SelectMessageQueueProvider 根据配置与开关状态选择消息队列 provider。
func SelectMessageQueueProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	backend := governanceProviderOverride(cfg, "message_queue")
	if backend == "" {
		backend = getConfigString(cfg, "message_queue.backend", "message_queue.type")
	}
	enabled := cfg != nil && cfg.GetBool("message_queue.enabled")
	if backend == "" && enabled {
		backend = "redis"
	}
	if backend == "" {
		backend = DefaultGovernanceProviderDefaults(DetectGovernanceMode(cfg)).MessageQueue
	}
	return providerFromMap(messageQueueProviderFactories, backend, "noop")
}

// SelectDistributedLockProvider selects the distributed-lock provider from config and enable flags.
//
// SelectDistributedLockProvider 根据配置与开关状态选择分布式锁 provider。
func SelectDistributedLockProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	backend := governanceProviderOverride(cfg, "distributed_lock")
	if backend == "" {
		backend = getConfigString(cfg, "distributed_lock.backend", "distributed_lock.type")
	}
	enabled := cfg != nil && cfg.GetBool("distributed_lock.enabled")
	if backend == "" && enabled {
		backend = "redis"
	}
	if backend == "" {
		backend = DefaultGovernanceProviderDefaults(DetectGovernanceMode(cfg)).DistributedLock
	}
	return providerFromMap(distributedLockProviderFactories, backend, "noop")
}

// getConfigString returns the first non-empty config value from the provided keys.
//
// getConfigString 返回给定 key 列表中第一个非空配置值。
func getConfigString(cfg datacontract.Config, keys ...string) string {
	if cfg == nil {
		return ""
	}
	for _, key := range keys {
		if value := cfg.GetString(key); value != "" {
			return value
		}
	}
	return ""
}

// SelectWebSocketProvider selects the WebSocket provider from config and enable flags.
//
// SelectWebSocketProvider 根据配置与开关状态选择 WebSocket provider。
func SelectWebSocketProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	backend := governanceProviderOverride(cfg, "websocket")
	if backend == "" {
		backend = getConfigString(cfg, "websocket.backend", "websocket.type")
	}
	enabled := cfg != nil && cfg.GetBool("websocket.enabled")
	if backend == "" && enabled {
		backend = "gws"
	}
	if backend == "" {
		backend = DefaultGovernanceProviderDefaults(DetectGovernanceMode(cfg)).WebSocket
	}
	return providerFromMap(webSocketProviderFactories, backend, "noop")
}
