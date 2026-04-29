package bootstrap

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
)

// SelectedMicroserviceProviders 根据当前配置返回主链路能力对应的单一 provider 集合。
//
// 中文说明：
// - 这是 capability selector 的最小骨架；
// - 目标是不再依赖 provider 注册顺序来碰运气，而是显式按配置选择 discovery / selector / rpc / tracing / metadata / serviceauth；
// - configsource 因为涉及两阶段加载，后续单独接入，此处先只负责其余主链路能力。
func SelectedMicroserviceProviders(cfg contract.Config) []contract.ServiceProvider {
	providers := make([]contract.ServiceProvider, 0, 10)
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

// RegisterSelectedMicroserviceProviders 根据配置把主链路 provider 注册进容器。
//
// 中文说明：
// - 先确保容器里已经有 ConfigKey，避免在无配置场景下误做主链路装配；
// - configsource 单独先注册，因为它可能会改变后续要读取的真实配置；
// - 如果启用了远端配置源，会在注册后先触发一次 Reload，再据此选择 discovery/rpc/tracing 等其余 provider。
func RegisterSelectedMicroserviceProviders(c contract.Container) error {
	if c == nil || !c.IsBind(contract.ConfigKey) {
		return nil
	}
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return err
	}
	cfg, _ := cfgAny.(contract.Config)

	configSourceProvider := SelectConfigSourceProvider(cfg)
	if configSourceProvider != nil {
		name := configSourceProvider.Name()
		if name != "configsource.local" && name != "configsource.noop" {
			if err := c.RegisterProvider(configSourceProvider); err != nil {
				return err
			}
			if cfgSvc, ok := cfgAny.(contract.Config); ok {
				if err := cfgSvc.Reload(context.Background()); err != nil {
					return err
				}
				cfg = cfgSvc
			}
		}
	}

	providers := []contract.ServiceProvider{
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

// SelectConfigSourceProvider 根据配置选择配置源 provider。
func SelectConfigSourceProvider(cfg contract.Config) contract.ServiceProvider {
	typ := getConfigString(
		cfg,
		"configsource.backend",
		"configsource.type",
		"config_source.backend",
		"config_source.type",
	)
	return providerFromMap(configSourceProviderFactories, typ, "local")
}

// SelectDiscoveryProvider 根据配置选择服务发现 provider。
func SelectDiscoveryProvider(cfg contract.Config) contract.ServiceProvider {
	typ := getConfigString(cfg, "discovery.backend", "discovery.type")
	return providerFromMap(discoveryProviderFactories, typ, "noop")
}

// SelectSelectorProvider 根据配置选择负载均衡 selector provider。
func SelectSelectorProvider(cfg contract.Config) contract.ServiceProvider {
	alg := getConfigString(cfg, "selector.backend", "selector.algorithm", "selector.type")
	return providerFromMap(selectorProviderFactories, alg, "noop")
}

// SelectRPCProvider 根据配置选择 RPC 传输 provider。
func SelectRPCProvider(cfg contract.Config) contract.ServiceProvider {
	mode := getConfigString(cfg, "rpc.mode")
	return providerFromMap(rpcProviderFactories, mode, "noop")
}

// SelectTracingProvider 根据配置选择 tracing provider。
func SelectTracingProvider(cfg contract.Config) contract.ServiceProvider {
	enabled := cfg != nil && cfg.GetBool("tracing.enabled")
	backend := getConfigString(cfg,
		"tracing.backend",
		"tracing.type",
	)
	if backend == "" && enabled {
		backend = "otel"
	}
	if backend == "otel" || backend == "otlp" || backend == "grpc" || backend == "http" || backend == "stdout" {
		return providerFromMap(tracingProviderFactories, backend, "noop")
	}
	return providerFromMap(tracingProviderFactories, backend, "noop")
}

// SelectMetadataProvider 根据配置选择 metadata 传播 provider。
func SelectMetadataProvider(cfg contract.Config) contract.ServiceProvider {
	mode := getConfigString(cfg,
		"metadata.mode",
		"metadata.backend",
	)
	enabled := cfg != nil && (cfg.Get("metadata.propagate_prefix") != nil || cfg.GetBool("metadata.enabled"))
	if mode == "" && enabled {
		mode = "default"
	}
	return providerFromMap(metadataProviderFactories, mode, "noop")
}

// SelectServiceAuthProvider 根据配置选择服务间认证 provider。
func SelectServiceAuthProvider(cfg contract.Config) contract.ServiceProvider {
	mode := getConfigString(cfg,
		"service_auth.backend",
		"service_auth.mode",
	)
	enabled := cfg != nil && cfg.GetBool("service_auth.enabled")
	if mode == "" && enabled {
		mode = "token"
	}
	return providerFromMap(serviceAuthProviderFactories, mode, "noop")
}

// SelectCircuitBreakerProvider 根据配置选择熔断 provider。
func SelectCircuitBreakerProvider(cfg contract.Config) contract.ServiceProvider {
	backend := getConfigString(cfg,
		"circuit_breaker.backend",
		"circuit_breaker.type",
	)
	enabled := cfg != nil && cfg.GetBool("circuit_breaker.enabled")
	if backend == "" && enabled {
		backend = "sentinel"
	}
	return providerFromMap(circuitBreakerProviderFactories, backend, "noop")
}

// SelectDTMProvider 根据配置选择分布式事务 provider。
func SelectDTMProvider(cfg contract.Config) contract.ServiceProvider {
	backend := getConfigString(cfg,
		"dtm.backend",
		"dtm.type",
		"dtm.driver",
	)
	enabled := cfg != nil && cfg.GetBool("dtm.enabled")
	if backend == "" && enabled {
		backend = "dtmsdk"
	}
	return providerFromMap(dtmProviderFactories, backend, "noop")
}

// SelectMessageQueueProvider 根据配置选择消息队列 provider。
func SelectMessageQueueProvider(cfg contract.Config) contract.ServiceProvider {
	backend := getConfigString(cfg,
		"message_queue.backend",
		"message_queue.type",
	)
	enabled := cfg != nil && cfg.GetBool("message_queue.enabled")
	if backend == "" && enabled {
		backend = "redis"
	}
	return providerFromMap(messageQueueProviderFactories, backend, "noop")
}

// SelectDistributedLockProvider 根据配置选择分布式锁 provider。
func SelectDistributedLockProvider(cfg contract.Config) contract.ServiceProvider {
	backend := getConfigString(cfg,
		"distributed_lock.backend",
		"distributed_lock.type",
	)
	enabled := cfg != nil && cfg.GetBool("distributed_lock.enabled")
	if backend == "" && enabled {
		backend = "redis"
	}
	return providerFromMap(distributedLockProviderFactories, backend, "noop")
}

// getConfigString 按顺序读取第一个非空配置项。
//
// 中文说明：
// - 用于兼容同一能力在不同命名下的配置键；
// - selector 层只关心“最终有效值”，不在这里做额外规范化。
func getConfigString(cfg contract.Config, keys ...string) string {
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
