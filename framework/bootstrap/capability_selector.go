package bootstrap

import (
	"context"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

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

func SelectConfigSourceProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	typ := getConfigString(cfg, "configsource.backend", "configsource.type", "config_source.backend", "config_source.type")
	return providerFromMap(configSourceProviderFactories, typ, "local")
}

func SelectDiscoveryProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	typ := getConfigString(cfg, "discovery.backend", "discovery.type")
	return providerFromMap(discoveryProviderFactories, typ, "noop")
}

func SelectSelectorProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	alg := getConfigString(cfg, "selector.backend", "selector.algorithm", "selector.type")
	return providerFromMap(selectorProviderFactories, alg, "noop")
}

func SelectRPCProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	mode := getConfigString(cfg, "rpc.mode")
	return providerFromMap(rpcProviderFactories, mode, "noop")
}

func SelectTracingProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	enabled := cfg != nil && cfg.GetBool("tracing.enabled")
	backend := getConfigString(cfg, "tracing.backend", "tracing.type")
	if backend == "" && enabled {
		backend = "otel"
	}
	return providerFromMap(tracingProviderFactories, backend, "noop")
}

func SelectMetadataProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	mode := getConfigString(cfg, "metadata.mode", "metadata.backend")
	enabled := cfg != nil && (cfg.Get("metadata.propagate_prefix") != nil || cfg.GetBool("metadata.enabled"))
	if mode == "" && enabled {
		mode = "default"
	}
	return providerFromMap(metadataProviderFactories, mode, "noop")
}

func SelectServiceAuthProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	mode := getConfigString(cfg, "service_auth.backend", "service_auth.mode")
	enabled := cfg != nil && cfg.GetBool("service_auth.enabled")
	if mode == "" && enabled {
		mode = "token"
	}
	return providerFromMap(serviceAuthProviderFactories, mode, "noop")
}

func SelectCircuitBreakerProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	backend := getConfigString(cfg, "circuit_breaker.backend", "circuit_breaker.type")
	enabled := cfg != nil && cfg.GetBool("circuit_breaker.enabled")
	if backend == "" && enabled {
		backend = "sentinel"
	}
	return providerFromMap(circuitBreakerProviderFactories, backend, "noop")
}

func SelectDTMProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	backend := getConfigString(cfg, "dtm.backend", "dtm.type", "dtm.driver")
	enabled := cfg != nil && cfg.GetBool("dtm.enabled")
	if backend == "" && enabled {
		backend = "dtmsdk"
	}
	return providerFromMap(dtmProviderFactories, backend, "noop")
}

func SelectMessageQueueProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	backend := getConfigString(cfg, "message_queue.backend", "message_queue.type")
	enabled := cfg != nil && cfg.GetBool("message_queue.enabled")
	if backend == "" && enabled {
		backend = "redis"
	}
	return providerFromMap(messageQueueProviderFactories, backend, "noop")
}

func SelectDistributedLockProvider(cfg datacontract.Config) runtimecontract.ServiceProvider {
	backend := getConfigString(cfg, "distributed_lock.backend", "distributed_lock.type")
	enabled := cfg != nil && cfg.GetBool("distributed_lock.enabled")
	if backend == "" && enabled {
		backend = "redis"
	}
	return providerFromMap(distributedLockProviderFactories, backend, "noop")
}

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
