package bootstrap

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
	configsourceconsul "github.com/ngq/gorp/framework/provider/configsource/consul"
	configsourceetcd "github.com/ngq/gorp/framework/provider/configsource/etcd"
	configsourcelocal "github.com/ngq/gorp/framework/provider/configsource/local"
	configsourcenoop "github.com/ngq/gorp/framework/provider/configsource/noop"
	discoveryconsul "github.com/ngq/gorp/framework/provider/discovery/consul"
	discoveryetcd "github.com/ngq/gorp/framework/provider/discovery/etcd"
	discoverynacos "github.com/ngq/gorp/framework/provider/discovery/nacos"
	discoverynoop "github.com/ngq/gorp/framework/provider/discovery/noop"
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
	providers := make([]contract.ServiceProvider, 0, 8)
	providers = append(providers, SelectDiscoveryProvider(cfg))
	providers = append(providers, SelectSelectorProvider(cfg))
	providers = append(providers, SelectRPCProvider(cfg))
	providers = append(providers, SelectTracingProvider(cfg))
	providers = append(providers, SelectMetadataProvider(cfg))
	providers = append(providers, SelectServiceAuthProvider(cfg))
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
	typ := getConfigString(cfg, "configsource.type", "config_source.type")
	switch typ {
	case "consul":
		return configsourceconsul.NewProvider()
	case "etcd":
		return configsourceetcd.NewProvider()
	case "noop":
		return configsourcenoop.NewProvider()
	case "", "local":
		fallthrough
	default:
		return configsourcelocal.NewProvider()
	}
}

func SelectDiscoveryProvider(cfg contract.Config) contract.ServiceProvider {
	typ := getConfigString(cfg, "discovery.type")
	switch typ {
	case "consul":
		return discoveryconsul.NewProvider()
	case "etcd":
		return discoveryetcd.NewProvider()
	case "nacos":
		return discoverynacos.NewProvider()
	case "", "noop":
		fallthrough
	default:
		return discoverynoop.NewProvider()
	}
}

func SelectSelectorProvider(cfg contract.Config) contract.ServiceProvider {
	alg := getConfigString(cfg, "selector.algorithm")
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
	if cfg != nil && cfg.GetBool("tracing.enabled") {
		return tracingotel.NewProvider()
	}
	return tracingnoop.NewProvider()
}

func SelectMetadataProvider(cfg contract.Config) contract.ServiceProvider {
	mode := getConfigString(cfg, "metadata.mode")
	enabled := cfg != nil && cfg.Get("metadata.propagate_prefix") != nil
	if mode == "default" || enabled {
		return metadatadefault.NewProvider()
	}
	return metadatanoop.NewProvider()
}

func SelectServiceAuthProvider(cfg contract.Config) contract.ServiceProvider {
	mode := getConfigString(cfg, "serviceauth.mode", "service_auth.mode")
	switch mode {
	case "token":
		return serviceauthtoken.NewProvider()
	case "mtls":
		return serviceauthmtls.NewProvider()
	case "", "noop":
		fallthrough
	default:
		return serviceauthnoop.NewProvider()
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
