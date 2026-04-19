package bootstrap

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
	circuitbreakernoop "github.com/ngq/gorp/framework/provider/circuitbreaker/noop"
	circuitbreakersentinel "github.com/ngq/gorp/framework/provider/circuitbreaker/sentinel"
	configsourceapollo "github.com/ngq/gorp/framework/provider/configsource/apollo"
	configsourceconsul "github.com/ngq/gorp/framework/provider/configsource/consul"
	configsourceetcd "github.com/ngq/gorp/framework/provider/configsource/etcd"
	configsourcekubernetes "github.com/ngq/gorp/framework/provider/configsource/kubernetes"
	configsourcelocal "github.com/ngq/gorp/framework/provider/configsource/local"
	configsourcenacos "github.com/ngq/gorp/framework/provider/configsource/nacos"
	configsourcenoop "github.com/ngq/gorp/framework/provider/configsource/noop"
	configsourcepolaris "github.com/ngq/gorp/framework/provider/configsource/polaris"
	discoveryconsul "github.com/ngq/gorp/framework/provider/discovery/consul"
	discoveryetcd "github.com/ngq/gorp/framework/provider/discovery/etcd"
	discoveryeureka "github.com/ngq/gorp/framework/provider/discovery/eureka"
	discoverykubernetes "github.com/ngq/gorp/framework/provider/discovery/kubernetes"
	discoverynacos "github.com/ngq/gorp/framework/provider/discovery/nacos"
	discoverynoop "github.com/ngq/gorp/framework/provider/discovery/noop"
	discoverypolaris "github.com/ngq/gorp/framework/provider/discovery/polaris"
	discoveryservicecomb "github.com/ngq/gorp/framework/provider/discovery/servicecomb"
	discoveryzookeeper "github.com/ngq/gorp/framework/provider/discovery/zookeeper"
	dtmnoop "github.com/ngq/gorp/framework/provider/dtm/noop"
	dtmsdk "github.com/ngq/gorp/framework/provider/dtm/dtmsdk"
	dlocknoop "github.com/ngq/gorp/framework/provider/dlock/noop"
	dlockredis "github.com/ngq/gorp/framework/provider/dlock/redis"
	mqnoop "github.com/ngq/gorp/framework/provider/messagequeue/noop"
	mqredis "github.com/ngq/gorp/framework/provider/messagequeue/redis"
	metadatadefault "github.com/ngq/gorp/framework/provider/metadata"
	metadatanoop "github.com/ngq/gorp/framework/provider/metadata/noop"
	rpcgrpc "github.com/ngq/gorp/framework/provider/rpc/grpc"
	rpchttp "github.com/ngq/gorp/framework/provider/rpc/http"
	rpcnoop "github.com/ngq/gorp/framework/provider/rpc/noop"
	selectornoop "github.com/ngq/gorp/framework/provider/selector/noop"
	selectorp2c "github.com/ngq/gorp/framework/provider/selector/p2c"
	selectorrandom "github.com/ngq/gorp/framework/provider/selector/random"
	selectorwrr "github.com/ngq/gorp/framework/provider/selector/wrr"
	serviceauthmtls "github.com/ngq/gorp/framework/provider/serviceauth/mtls"
	serviceauthnoop "github.com/ngq/gorp/framework/provider/serviceauth/noop"
	serviceauthtoken "github.com/ngq/gorp/framework/provider/serviceauth/token"
	tracingnoop "github.com/ngq/gorp/framework/provider/tracing/noop"
	tracingotel "github.com/ngq/gorp/framework/provider/tracing/otel"
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

func SelectConfigSourceProvider(cfg contract.Config) contract.ServiceProvider {
	typ := getConfigString(
		cfg,
		"configsource.backend",
		"configsource.type",
		"config_source.backend",
		"config_source.type",
	)
	switch typ {
	case "consul":
		return configsourceconsul.NewProvider()
	case "etcd":
		return configsourceetcd.NewProvider()
	case "apollo":
		return configsourceapollo.NewProvider()
	case "nacos":
		return configsourcenacos.NewProvider()
	case "kubernetes":
		return configsourcekubernetes.NewProvider()
	case "polaris":
		return configsourcepolaris.NewProvider()
	case "noop":
		return configsourcenoop.NewProvider()
	case "", "local":
		fallthrough
	default:
		return configsourcelocal.NewProvider()
	}
}

func SelectDiscoveryProvider(cfg contract.Config) contract.ServiceProvider {
	typ := getConfigString(cfg, "discovery.backend", "discovery.type")
	switch typ {
	case "consul":
		return discoveryconsul.NewProvider()
	case "etcd":
		return discoveryetcd.NewProvider()
	case "nacos":
		return discoverynacos.NewProvider()
	case "zookeeper":
		return discoveryzookeeper.NewProvider()
	case "kubernetes":
		return discoverykubernetes.NewProvider()
	case "polaris":
		return discoverypolaris.NewProvider()
	case "eureka":
		return discoveryeureka.NewProvider()
	case "servicecomb":
		return discoveryservicecomb.NewProvider()
	case "", "noop":
		fallthrough
	default:
		return discoverynoop.NewProvider()
	}
}

func SelectSelectorProvider(cfg contract.Config) contract.ServiceProvider {
	alg := getConfigString(cfg, "selector.backend", "selector.algorithm", "selector.type")
	switch alg {
	case "random":
		return selectorrandom.NewProvider()
	case "wrr":
		return selectorwrr.NewProvider()
	case "p2c":
		return selectorp2c.NewProvider()
	case "", "noop":
		fallthrough
	default:
		return selectornoop.NewProvider()
	}
}

func SelectRPCProvider(cfg contract.Config) contract.ServiceProvider {
	mode := getConfigString(cfg, "rpc.mode")
	switch mode {
	case "http":
		return rpchttp.NewProvider()
	case "grpc":
		return rpcgrpc.NewProvider()
	case "", "noop":
		fallthrough
	default:
		return rpcnoop.NewProvider()
	}
}

func SelectTracingProvider(cfg contract.Config) contract.ServiceProvider {
	enabled := cfg != nil && cfg.GetBool("tracing.enabled")
	backend := getConfigString(cfg,
		"tracing.backend",
		"tracing.type",
	)
	if enabled || backend == "otel" || backend == "otlp" || backend == "grpc" || backend == "http" || backend == "stdout" {
		return tracingotel.NewProvider()
	}
	return tracingnoop.NewProvider()
}

func SelectMetadataProvider(cfg contract.Config) contract.ServiceProvider {
	mode := getConfigString(cfg,
		"metadata.mode",
		"metadata.backend",
	)
	enabled := cfg != nil && (cfg.Get("metadata.propagate_prefix") != nil || cfg.GetBool("metadata.enabled"))
	if mode == "default" || enabled {
		return metadatadefault.NewProvider()
	}
	return metadatanoop.NewProvider()
}

func SelectServiceAuthProvider(cfg contract.Config) contract.ServiceProvider {
	mode := getConfigString(cfg,
		"service_auth.backend",
		"service_auth.mode",
	)
	enabled := cfg != nil && cfg.GetBool("service_auth.enabled")
	switch mode {
	case "token":
		return serviceauthtoken.NewProvider()
	case "mtls":
		return serviceauthmtls.NewProvider()
	case "", "noop":
		if enabled {
			return serviceauthtoken.NewProvider()
		}
		fallthrough
	default:
		return serviceauthnoop.NewProvider()
	}
}

func SelectCircuitBreakerProvider(cfg contract.Config) contract.ServiceProvider {
	backend := getConfigString(cfg,
		"circuit_breaker.backend",
		"circuit_breaker.type",
	)
	enabled := cfg != nil && cfg.GetBool("circuit_breaker.enabled")
	switch backend {
	case "sentinel":
		return circuitbreakersentinel.NewProvider()
	case "", "noop":
		if enabled {
			return circuitbreakersentinel.NewProvider()
		}
		fallthrough
	default:
		return circuitbreakernoop.NewProvider()
	}
}

func SelectDTMProvider(cfg contract.Config) contract.ServiceProvider {
	backend := getConfigString(cfg,
		"dtm.backend",
		"dtm.type",
		"dtm.driver",
	)
	enabled := cfg != nil && cfg.GetBool("dtm.enabled")
	switch backend {
	case "sdk", "dtmsdk":
		return dtmsdk.NewProvider()
	case "", "noop":
		if enabled {
			return dtmsdk.NewProvider()
		}
		fallthrough
	default:
		return dtmnoop.NewProvider()
	}
}

func SelectMessageQueueProvider(cfg contract.Config) contract.ServiceProvider {
	backend := getConfigString(cfg,
		"message_queue.backend",
		"message_queue.type",
	)
	enabled := cfg != nil && cfg.GetBool("message_queue.enabled")
	switch backend {
	case "redis":
		return mqredis.NewProvider()
	case "", "noop":
		if enabled {
			return mqredis.NewProvider()
		}
		fallthrough
	default:
		return mqnoop.NewProvider()
	}
}

func SelectDistributedLockProvider(cfg contract.Config) contract.ServiceProvider {
	backend := getConfigString(cfg,
		"distributed_lock.backend",
		"distributed_lock.type",
	)
	enabled := cfg != nil && cfg.GetBool("distributed_lock.enabled")
	switch backend {
	case "redis":
		return dlockredis.NewProvider()
	case "", "noop":
		if enabled {
			return dlockredis.NewProvider()
		}
		fallthrough
	default:
		return dlocknoop.NewProvider()
	}
}

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
