// Application scenarios:
// - Select capability providers from runtime configuration during bootstrap.
// - Keep config-driven microservice assembly centralized and predictable.
// - Support explicit backend selection together with feature-enable fallbacks.
//
// 适用场景：
// - 在 bootstrap 阶段根据运行时配置选择能力 provider。
// - 将配置驱动的微服务装配逻辑集中管理并保持可预测性。
// - 同时支持显式后端选择和基于开关的兜底推断。
package bootstrap

import (
	"context"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// SelectedMicroserviceProviders returns the config-selected microservice capability providers.
//
// SelectedMicroserviceProviders 返回由配置选择出的微服务能力 provider 集合。
func SelectedMicroserviceProviders(cfg datacontract.Config) []runtimecontract.ServiceProvider {
	providers := make([]runtimecontract.ServiceProvider, 0, 10)
	providers = append(providers, SelectDiscoveryProvider(cfg))
	providers = append(providers, SelectSelectorProvider(cfg))
	providers = append(providers, SelectRPCProvider(cfg))
	providers = append(providers, SelectTracingProvider(cfg))
	providers = append(providers, SelectMetadataProvider(cfg))
	providers = append(providers, SelectServiceAuthProvider(cfg))
	providers = append(providers, SelectCircuitBreakerProvider(cfg))
	providers = append(providers, SelectDTMProvider(cfg))
	providers = append(providers, SelectMessageQueueProvider(cfg))
	providers = append(providers, SelectDistributedLockProvider(cfg))
	return providers
}

// RegisterSelectedMicroserviceProviders resolves and registers microservice capability providers into the container.
//
// RegisterSelectedMicroserviceProviders 解析并将微服务能力 provider 注册到容器中。
func RegisterSelectedMicroserviceProviders(c runtimecontract.Container) error {
	if c == nil || !c.IsBind(datacontract.ConfigKey) {
		return nil
	}
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return err
	}
	cfg, _ := cfgAny.(datacontract.Config)

	configSourceProvider := SelectConfigSourceProvider(cfg)
	if configSourceProvider != nil {
		name := configSourceProvider.Name()
		if name != "configsource.local" && name != "configsource.noop" {
			// External config sources should be registered first so reload can refresh the effective config snapshot.
			// 外部配置源需要优先注册，这样后续 reload 才能刷新实际生效的配置快照。
			if err := c.RegisterProvider(configSourceProvider); err != nil {
				return err
			}
			if cfgSvc, ok := cfgAny.(datacontract.Config); ok {
				if err := cfgSvc.Reload(context.Background()); err != nil {
					return err
				}
				cfg = cfgSvc
			}
		}
	}

	providers := []runtimecontract.ServiceProvider{
		SelectDiscoveryProvider(cfg),
		SelectSelectorProvider(cfg),
		SelectRPCProvider(cfg),
		SelectTracingProvider(cfg),
		SelectMetadataProvider(cfg),
		SelectServiceAuthProvider(cfg),
		SelectCircuitBreakerProvider(cfg),
		SelectDTMProvider(cfg),
		SelectMessageQueueProvider(cfg),
		SelectDistributedLockProvider(cfg),
	}
	for _, p := range providers {
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
	typ := getConfigString(cfg, "configsource.backend", "configsource.type", "config_source.backend", "config_source.type")
	return providerFromMap(configSourceProviderFactories, typ, "local")
}

// SelectDiscoveryProvider selects the discovery provider from config.
//
// SelectDiscoveryProvider 根据配置选择服务发现 provider。
func SelectDiscoveryProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	typ := getConfigString(cfg, "discovery.backend", "discovery.type")
	return providerFromMap(discoveryProviderFactories, typ, "noop")
}

// SelectSelectorProvider selects the selector provider from config.
//
// SelectSelectorProvider 根据配置选择 selector provider。
func SelectSelectorProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	alg := getConfigString(cfg, "selector.backend", "selector.algorithm", "selector.type")
	return providerFromMap(selectorProviderFactories, alg, "noop")
}

// SelectRPCProvider selects the RPC provider from config.
//
// SelectRPCProvider 根据配置选择 RPC provider。
func SelectRPCProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	mode := getConfigString(cfg, "rpc.mode")
	return providerFromMap(rpcProviderFactories, mode, "noop")
}

// SelectTracingProvider selects the tracing provider from config and enable flags.
//
// SelectTracingProvider 根据配置和开关状态选择 tracing provider。
func SelectTracingProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	enabled := cfg != nil && cfg.GetBool("tracing.enabled")
	backend := getConfigString(cfg, "tracing.backend", "tracing.type")
	if backend == "" && enabled {
		// Promote to the mainstream otel implementation when tracing is enabled but no backend is declared.
		// tracing 已启用但未声明后端时，自动提升到主流的 otel 实现。
		backend = "otel"
	}
	return providerFromMap(tracingProviderFactories, backend, "noop")
}

// SelectMetadataProvider selects the metadata provider from config and propagation hints.
//
// SelectMetadataProvider 根据配置和透传迹象选择 metadata provider。
func SelectMetadataProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	mode := getConfigString(cfg, "metadata.mode", "metadata.backend")
	enabled := cfg != nil && (cfg.Get("metadata.propagate_prefix") != nil || cfg.GetBool("metadata.enabled"))
	if mode == "" && enabled {
		mode = "default"
	}
	return providerFromMap(metadataProviderFactories, mode, "noop")
}

// SelectServiceAuthProvider selects the service-auth provider from config and enable flags.
//
// SelectServiceAuthProvider 根据配置和开关状态选择服务鉴权 provider。
func SelectServiceAuthProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	mode := getConfigString(cfg, "service_auth.backend", "service_auth.mode")
	enabled := cfg != nil && cfg.GetBool("service_auth.enabled")
	if mode == "" && enabled {
		mode = "token"
	}
	return providerFromMap(serviceAuthProviderFactories, mode, "noop")
}

// SelectCircuitBreakerProvider selects the circuit-breaker provider from config and enable flags.
//
// SelectCircuitBreakerProvider 根据配置和开关状态选择熔断 provider。
func SelectCircuitBreakerProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	backend := getConfigString(cfg, "circuit_breaker.backend", "circuit_breaker.type")
	enabled := cfg != nil && cfg.GetBool("circuit_breaker.enabled")
	if backend == "" && enabled {
		backend = "sentinel"
	}
	return providerFromMap(circuitBreakerProviderFactories, backend, "noop")
}

// SelectDTMProvider selects the DTM provider from config and enable flags.
//
// SelectDTMProvider 根据配置和开关状态选择 DTM provider。
func SelectDTMProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	backend := getConfigString(cfg, "dtm.backend", "dtm.type", "dtm.driver")
	enabled := cfg != nil && cfg.GetBool("dtm.enabled")
	if backend == "" && enabled {
		backend = "dtmsdk"
	}
	return providerFromMap(dtmProviderFactories, backend, "noop")
}

// SelectMessageQueueProvider selects the message-queue provider from config and enable flags.
//
// SelectMessageQueueProvider 根据配置和开关状态选择消息队列 provider。
func SelectMessageQueueProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	backend := getConfigString(cfg, "message_queue.backend", "message_queue.type")
	enabled := cfg != nil && cfg.GetBool("message_queue.enabled")
	if backend == "" && enabled {
		backend = "redis"
	}
	return providerFromMap(messageQueueProviderFactories, backend, "noop")
}

// SelectDistributedLockProvider selects the distributed-lock provider from config and enable flags.
//
// SelectDistributedLockProvider 根据配置和开关状态选择分布式锁 provider。
func SelectDistributedLockProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	backend := getConfigString(cfg, "distributed_lock.backend", "distributed_lock.type")
	enabled := cfg != nil && cfg.GetBool("distributed_lock.enabled")
	if backend == "" && enabled {
		backend = "redis"
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
		// Return the first non-empty candidate so legacy and new key names can coexist.
		// 返回第一个非空候选值，以便兼容多套新旧配置键名。
		if value := cfg.GetString(key); value != "" {
			return value
		}
	}
	return ""
}
