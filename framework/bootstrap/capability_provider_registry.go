package bootstrap

import (
	"github.com/ngq/gorp/framework/contract"
	circuitbreakernoop "github.com/ngq/gorp/framework/provider/circuitbreaker/noop"
	circuitbreakersentinel "github.com/ngq/gorp/contrib/circuitbreaker/sentinel"
	configsourceapollo "github.com/ngq/gorp/contrib/configsource/apollo"
	configsourceconsul "github.com/ngq/gorp/contrib/configsource/consul"
	configsourceetcd "github.com/ngq/gorp/contrib/configsource/etcd"
	configsourcekubernetes "github.com/ngq/gorp/contrib/configsource/kubernetes"
	configsourcelocal "github.com/ngq/gorp/framework/provider/configsource/local"
	configsourcenacos "github.com/ngq/gorp/contrib/configsource/nacos"
	configsourcenoop "github.com/ngq/gorp/contrib/configsource/noop"
	configsourcepolaris "github.com/ngq/gorp/contrib/configsource/polaris"
	discoveryconsul "github.com/ngq/gorp/contrib/registry/consul"
	discoveryetcd "github.com/ngq/gorp/contrib/registry/etcd"
	discoveryeureka "github.com/ngq/gorp/contrib/registry/eureka"
	discoverykubernetes "github.com/ngq/gorp/contrib/registry/kubernetes"
	discoverynacos "github.com/ngq/gorp/contrib/registry/nacos"
	discoverynoop "github.com/ngq/gorp/contrib/registry/noop"
	discoverypolaris "github.com/ngq/gorp/contrib/registry/polaris"
	discoveryservicecomb "github.com/ngq/gorp/contrib/registry/servicecomb"
	discoveryzookeeper "github.com/ngq/gorp/contrib/registry/zookeeper"
	dtmnoop "github.com/ngq/gorp/framework/provider/dtm/noop"
	dtmsdk "github.com/ngq/gorp/framework/provider/dtm/dtmsdk"
	dlocknoop "github.com/ngq/gorp/framework/provider/dlock/noop"
	dlockredis "github.com/ngq/gorp/contrib/dlock/redis"
	mqnoop "github.com/ngq/gorp/framework/provider/messagequeue/noop"
	mqredis "github.com/ngq/gorp/contrib/mq/redis"
	metadatadefault "github.com/ngq/gorp/framework/provider/metadata"
	metadatanoop "github.com/ngq/gorp/framework/provider/metadata/noop"
	rpcgrpc "github.com/ngq/gorp/framework/provider/rpc/grpc"
	rpchttp "github.com/ngq/gorp/framework/provider/rpc/http"
	rpcnoop "github.com/ngq/gorp/framework/provider/rpc/noop"
	selectornoop "github.com/ngq/gorp/framework/provider/selector/noop"
	selectorp2c "github.com/ngq/gorp/framework/provider/selector/p2c"
	selectorrandom "github.com/ngq/gorp/framework/provider/selector/random"
	selectorwrr "github.com/ngq/gorp/framework/provider/selector/wrr"
	serviceauthmtls "github.com/ngq/gorp/contrib/serviceauth/mtls"
	serviceauthnoop "github.com/ngq/gorp/framework/provider/serviceauth/noop"
	serviceauthtoken "github.com/ngq/gorp/contrib/serviceauth/token"
	tracingnoop "github.com/ngq/gorp/contrib/tracing/noop"
	tracingotel "github.com/ngq/gorp/contrib/tracing/otel"
)

type providerFactory func() contract.ServiceProvider

type providerFactoryRegistry map[string]providerFactory

func (r providerFactoryRegistry) register(key string, factory providerFactory) {
	if key == "" || factory == nil {
		return
	}
	r[key] = factory
}

// RegisterConfigSourceProviderFactory 注册配置源 provider 工厂。
func RegisterConfigSourceProviderFactory(key string, factory providerFactory) {
	configSourceProviderFactories.register(key, factory)
}

// RegisterDiscoveryProviderFactory 注册服务发现 provider 工厂。
func RegisterDiscoveryProviderFactory(key string, factory providerFactory) {
	discoveryProviderFactories.register(key, factory)
}

// RegisterSelectorProviderFactory 注册负载均衡 provider 工厂。
func RegisterSelectorProviderFactory(key string, factory providerFactory) {
	selectorProviderFactories.register(key, factory)
}

// RegisterRPCProviderFactory 注册 RPC provider 工厂。
func RegisterRPCProviderFactory(key string, factory providerFactory) {
	rpcProviderFactories.register(key, factory)
}

// RegisterServiceAuthProviderFactory 注册服务间认证 provider 工厂。
func RegisterServiceAuthProviderFactory(key string, factory providerFactory) {
	serviceAuthProviderFactories.register(key, factory)
}

// RegisterCircuitBreakerProviderFactory 注册熔断 provider 工厂。
func RegisterCircuitBreakerProviderFactory(key string, factory providerFactory) {
	circuitBreakerProviderFactories.register(key, factory)
}

// RegisterDTMProviderFactory 注册分布式事务 provider 工厂。
func RegisterDTMProviderFactory(key string, factory providerFactory) {
	dtmProviderFactories.register(key, factory)
}

// RegisterMessageQueueProviderFactory 注册消息队列 provider 工厂。
func RegisterMessageQueueProviderFactory(key string, factory providerFactory) {
	messageQueueProviderFactories.register(key, factory)
}

// RegisterTracingProviderFactory 注册 tracing provider 工厂。
func RegisterTracingProviderFactory(key string, factory providerFactory) {
	tracingProviderFactories.register(key, factory)
}

// RegisterMetadataProviderFactory 注册 metadata provider 工厂。
func RegisterMetadataProviderFactory(key string, factory providerFactory) {
	metadataProviderFactories.register(key, factory)
}

var (
	configSourceProviderFactories = providerFactoryRegistry{
		"consul":     func() contract.ServiceProvider { return configsourceconsul.NewProvider() },
		"etcd":       func() contract.ServiceProvider { return configsourceetcd.NewProvider() },
		"apollo":     func() contract.ServiceProvider { return configsourceapollo.NewProvider() },
		"nacos":      func() contract.ServiceProvider { return configsourcenacos.NewProvider() },
		"kubernetes": func() contract.ServiceProvider { return configsourcekubernetes.NewProvider() },
		"polaris":    func() contract.ServiceProvider { return configsourcepolaris.NewProvider() },
		"noop":       func() contract.ServiceProvider { return configsourcenoop.NewProvider() },
		"local":      func() contract.ServiceProvider { return configsourcelocal.NewProvider() },
		"":           func() contract.ServiceProvider { return configsourcelocal.NewProvider() },
	}
	discoveryProviderFactories = providerFactoryRegistry{
		"consul":      func() contract.ServiceProvider { return discoveryconsul.NewProvider() },
		"etcd":        func() contract.ServiceProvider { return discoveryetcd.NewProvider() },
		"nacos":       func() contract.ServiceProvider { return discoverynacos.NewProvider() },
		"zookeeper":   func() contract.ServiceProvider { return discoveryzookeeper.NewProvider() },
		"kubernetes":  func() contract.ServiceProvider { return discoverykubernetes.NewProvider() },
		"polaris":     func() contract.ServiceProvider { return discoverypolaris.NewProvider() },
		"eureka":      func() contract.ServiceProvider { return discoveryeureka.NewProvider() },
		"servicecomb": func() contract.ServiceProvider { return discoveryservicecomb.NewProvider() },
		"noop":        func() contract.ServiceProvider { return discoverynoop.NewProvider() },
		"":            func() contract.ServiceProvider { return discoverynoop.NewProvider() },
	}
	selectorProviderFactories = providerFactoryRegistry{
		"random": func() contract.ServiceProvider { return selectorrandom.NewProvider() },
		"wrr":    func() contract.ServiceProvider { return selectorwrr.NewProvider() },
		"p2c":    func() contract.ServiceProvider { return selectorp2c.NewProvider() },
		"noop":   func() contract.ServiceProvider { return selectornoop.NewProvider() },
		"":       func() contract.ServiceProvider { return selectornoop.NewProvider() },
	}
	rpcProviderFactories = providerFactoryRegistry{
		"http": func() contract.ServiceProvider { return rpchttp.NewProvider() },
		"grpc": func() contract.ServiceProvider { return rpcgrpc.NewProvider() },
		"noop": func() contract.ServiceProvider { return rpcnoop.NewProvider() },
		"":     func() contract.ServiceProvider { return rpcnoop.NewProvider() },
	}
	tracingProviderFactories = providerFactoryRegistry{
		"otel":   func() contract.ServiceProvider { return tracingotel.NewProvider() },
		"otlp":   func() contract.ServiceProvider { return tracingotel.NewProvider() },
		"grpc":   func() contract.ServiceProvider { return tracingotel.NewProvider() },
		"http":   func() contract.ServiceProvider { return tracingotel.NewProvider() },
		"stdout": func() contract.ServiceProvider { return tracingotel.NewProvider() },
		"noop":   func() contract.ServiceProvider { return tracingnoop.NewProvider() },
		"":       func() contract.ServiceProvider { return tracingnoop.NewProvider() },
	}
	metadataProviderFactories = providerFactoryRegistry{
		"default": func() contract.ServiceProvider { return metadatadefault.NewProvider() },
		"noop":    func() contract.ServiceProvider { return metadatanoop.NewProvider() },
		"":        func() contract.ServiceProvider { return metadatanoop.NewProvider() },
	}
	serviceAuthProviderFactories = providerFactoryRegistry{
		"token": func() contract.ServiceProvider { return serviceauthtoken.NewProvider() },
		"mtls":  func() contract.ServiceProvider { return serviceauthmtls.NewProvider() },
		"noop":  func() contract.ServiceProvider { return serviceauthnoop.NewProvider() },
		"":      func() contract.ServiceProvider { return serviceauthnoop.NewProvider() },
	}
	circuitBreakerProviderFactories = providerFactoryRegistry{
		"sentinel": func() contract.ServiceProvider { return circuitbreakersentinel.NewProvider() },
		"noop":     func() contract.ServiceProvider { return circuitbreakernoop.NewProvider() },
		"":         func() contract.ServiceProvider { return circuitbreakernoop.NewProvider() },
	}
	dtmProviderFactories = providerFactoryRegistry{
		"sdk":    func() contract.ServiceProvider { return dtmsdk.NewProvider() },
		"dtmsdk": func() contract.ServiceProvider { return dtmsdk.NewProvider() },
		"noop":   func() contract.ServiceProvider { return dtmnoop.NewProvider() },
		"":       func() contract.ServiceProvider { return dtmnoop.NewProvider() },
	}
	messageQueueProviderFactories = providerFactoryRegistry{
		"redis": func() contract.ServiceProvider { return mqredis.NewProvider() },
		"noop":  func() contract.ServiceProvider { return mqnoop.NewProvider() },
		"":      func() contract.ServiceProvider { return mqnoop.NewProvider() },
	}
	distributedLockProviderFactories = providerFactoryRegistry{
		"redis": func() contract.ServiceProvider { return dlockredis.NewProvider() },
		"noop":  func() contract.ServiceProvider { return dlocknoop.NewProvider() },
		"":      func() contract.ServiceProvider { return dlocknoop.NewProvider() },
	}
)

func providerFromMap(factories map[string]providerFactory, key string, fallback string) contract.ServiceProvider {
	if factory, ok := factories[key]; ok {
		return factory()
	}
	if factory, ok := factories[fallback]; ok {
		return factory()
	}
	return nil
}

func defaultTracingProvider() contract.ServiceProvider {
	return tracingnoop.NewProvider()
}

func enabledTracingProvider() contract.ServiceProvider {
	return tracingotel.NewProvider()
}

func defaultMetadataProvider() contract.ServiceProvider {
	return metadatanoop.NewProvider()
}

func enabledMetadataProvider() contract.ServiceProvider {
	return metadatadefault.NewProvider()
}
