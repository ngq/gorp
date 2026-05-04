package bootstrap

import (
	circuitbreakersentinel "github.com/ngq/gorp/contrib/circuitbreaker/sentinel"
	configsourceapollo "github.com/ngq/gorp/contrib/configsource/apollo"
	configsourceconsul "github.com/ngq/gorp/contrib/configsource/consul"
	configsourceetcd "github.com/ngq/gorp/contrib/configsource/etcd"
	configsourcekubernetes "github.com/ngq/gorp/contrib/configsource/kubernetes"
	configsourcenacos "github.com/ngq/gorp/contrib/configsource/nacos"
	configsourcepolaris "github.com/ngq/gorp/contrib/configsource/polaris"
	dlockredis "github.com/ngq/gorp/contrib/dlock/redis"
	dtmsdk "github.com/ngq/gorp/contrib/dtm/dtmsdk"
	mqredis "github.com/ngq/gorp/contrib/messagequeue/redis"
	discoveryconsul "github.com/ngq/gorp/contrib/registry/consul"
	discoveryetcd "github.com/ngq/gorp/contrib/registry/etcd"
	discoveryeureka "github.com/ngq/gorp/contrib/registry/eureka"
	discoverykubernetes "github.com/ngq/gorp/contrib/registry/kubernetes"
	discoverynacos "github.com/ngq/gorp/contrib/registry/nacos"
	discoverypolaris "github.com/ngq/gorp/contrib/registry/polaris"
	discoveryservicecomb "github.com/ngq/gorp/contrib/registry/servicecomb"
	discoveryzookeeper "github.com/ngq/gorp/contrib/registry/zookeeper"
	serviceauthmtls "github.com/ngq/gorp/contrib/serviceauth/mtls"
	serviceauthtoken "github.com/ngq/gorp/contrib/serviceauth/token"
	tracingotel "github.com/ngq/gorp/contrib/tracing/otel"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	circuitbreakernoop "github.com/ngq/gorp/framework/provider/circuitbreaker/noop"
	configsourcelocal "github.com/ngq/gorp/framework/provider/configsource/local"
	configsourcenoop "github.com/ngq/gorp/framework/provider/configsource/noop"
	discoverynoop "github.com/ngq/gorp/framework/provider/discovery/noop"
	dlocknoop "github.com/ngq/gorp/framework/provider/dlock/noop"
	dtmnoop "github.com/ngq/gorp/framework/provider/dtm/noop"
	mqnoop "github.com/ngq/gorp/framework/provider/messagequeue/noop"
	metadatadefault "github.com/ngq/gorp/framework/provider/metadata"
	metadatanoop "github.com/ngq/gorp/framework/provider/metadata/noop"
	rpcgrpc "github.com/ngq/gorp/framework/provider/rpc/grpc"
	rpchttp "github.com/ngq/gorp/framework/provider/rpc/http"
	rpcnoop "github.com/ngq/gorp/framework/provider/rpc/noop"
	selectornoop "github.com/ngq/gorp/framework/provider/selector/noop"
	selectorp2c "github.com/ngq/gorp/framework/provider/selector/p2c"
	selectorrandom "github.com/ngq/gorp/framework/provider/selector/random"
	selectorwrr "github.com/ngq/gorp/framework/provider/selector/wrr"
	serviceauthnoop "github.com/ngq/gorp/framework/provider/serviceauth/noop"
	tracingnoop "github.com/ngq/gorp/framework/provider/tracing/noop"
)

type providerFactory func() runtimecontract.ServiceProvider

type providerFactoryRegistry map[string]providerFactory

func (r providerFactoryRegistry) register(key string, factory providerFactory) {
	if key == "" || factory == nil {
		return
	}
	r[key] = factory
}

func RegisterConfigSourceProviderFactory(key string, factory providerFactory) {
	configSourceProviderFactories.register(key, factory)
}

func RegisterDiscoveryProviderFactory(key string, factory providerFactory) {
	discoveryProviderFactories.register(key, factory)
}

func RegisterSelectorProviderFactory(key string, factory providerFactory) {
	selectorProviderFactories.register(key, factory)
}

func RegisterRPCProviderFactory(key string, factory providerFactory) {
	rpcProviderFactories.register(key, factory)
}

func RegisterServiceAuthProviderFactory(key string, factory providerFactory) {
	serviceAuthProviderFactories.register(key, factory)
}

func RegisterCircuitBreakerProviderFactory(key string, factory providerFactory) {
	circuitBreakerProviderFactories.register(key, factory)
}

func RegisterDTMProviderFactory(key string, factory providerFactory) {
	dtmProviderFactories.register(key, factory)
}

func RegisterMessageQueueProviderFactory(key string, factory providerFactory) {
	messageQueueProviderFactories.register(key, factory)
}

func RegisterTracingProviderFactory(key string, factory providerFactory) {
	tracingProviderFactories.register(key, factory)
}

func RegisterMetadataProviderFactory(key string, factory providerFactory) {
	metadataProviderFactories.register(key, factory)
}

var (
	configSourceProviderFactories = providerFactoryRegistry{
		"consul":     func() runtimecontract.ServiceProvider { return configsourceconsul.NewProvider() },
		"etcd":       func() runtimecontract.ServiceProvider { return configsourceetcd.NewProvider() },
		"apollo":     func() runtimecontract.ServiceProvider { return configsourceapollo.NewProvider() },
		"nacos":      func() runtimecontract.ServiceProvider { return configsourcenacos.NewProvider() },
		"kubernetes": func() runtimecontract.ServiceProvider { return configsourcekubernetes.NewProvider() },
		"polaris":    func() runtimecontract.ServiceProvider { return configsourcepolaris.NewProvider() },
		"noop":       func() runtimecontract.ServiceProvider { return configsourcenoop.NewProvider() },
		"local":      func() runtimecontract.ServiceProvider { return configsourcelocal.NewProvider() },
		"":           func() runtimecontract.ServiceProvider { return configsourcelocal.NewProvider() },
	}
	discoveryProviderFactories = providerFactoryRegistry{
		"consul":      func() runtimecontract.ServiceProvider { return discoveryconsul.NewProvider() },
		"etcd":        func() runtimecontract.ServiceProvider { return discoveryetcd.NewProvider() },
		"nacos":       func() runtimecontract.ServiceProvider { return discoverynacos.NewProvider() },
		"zookeeper":   func() runtimecontract.ServiceProvider { return discoveryzookeeper.NewProvider() },
		"kubernetes":  func() runtimecontract.ServiceProvider { return discoverykubernetes.NewProvider() },
		"polaris":     func() runtimecontract.ServiceProvider { return discoverypolaris.NewProvider() },
		"eureka":      func() runtimecontract.ServiceProvider { return discoveryeureka.NewProvider() },
		"servicecomb": func() runtimecontract.ServiceProvider { return discoveryservicecomb.NewProvider() },
		"noop":        func() runtimecontract.ServiceProvider { return discoverynoop.NewProvider() },
		"":            func() runtimecontract.ServiceProvider { return discoverynoop.NewProvider() },
	}
	selectorProviderFactories = providerFactoryRegistry{
		"random": func() runtimecontract.ServiceProvider { return selectorrandom.NewProvider() },
		"wrr":    func() runtimecontract.ServiceProvider { return selectorwrr.NewProvider() },
		"p2c":    func() runtimecontract.ServiceProvider { return selectorp2c.NewProvider() },
		"noop":   func() runtimecontract.ServiceProvider { return selectornoop.NewProvider() },
		"":       func() runtimecontract.ServiceProvider { return selectornoop.NewProvider() },
	}
	rpcProviderFactories = providerFactoryRegistry{
		"http": func() runtimecontract.ServiceProvider { return rpchttp.NewProvider() },
		"grpc": func() runtimecontract.ServiceProvider { return rpcgrpc.NewProvider() },
		"noop": func() runtimecontract.ServiceProvider { return rpcnoop.NewProvider() },
		"":     func() runtimecontract.ServiceProvider { return rpcnoop.NewProvider() },
	}
	tracingProviderFactories = providerFactoryRegistry{
		"otel":   func() runtimecontract.ServiceProvider { return tracingotel.NewProvider() },
		"otlp":   func() runtimecontract.ServiceProvider { return tracingotel.NewProvider() },
		"grpc":   func() runtimecontract.ServiceProvider { return tracingotel.NewProvider() },
		"http":   func() runtimecontract.ServiceProvider { return tracingotel.NewProvider() },
		"stdout": func() runtimecontract.ServiceProvider { return tracingotel.NewProvider() },
		"noop":   func() runtimecontract.ServiceProvider { return tracingnoop.NewProvider() },
		"":       func() runtimecontract.ServiceProvider { return tracingnoop.NewProvider() },
	}
	metadataProviderFactories = providerFactoryRegistry{
		"default": func() runtimecontract.ServiceProvider { return metadatadefault.NewProvider() },
		"noop":    func() runtimecontract.ServiceProvider { return metadatanoop.NewProvider() },
		"":        func() runtimecontract.ServiceProvider { return metadatanoop.NewProvider() },
	}
	serviceAuthProviderFactories = providerFactoryRegistry{
		"token": func() runtimecontract.ServiceProvider { return serviceauthtoken.NewProvider() },
		"mtls":  func() runtimecontract.ServiceProvider { return serviceauthmtls.NewProvider() },
		"noop":  func() runtimecontract.ServiceProvider { return serviceauthnoop.NewProvider() },
		"":      func() runtimecontract.ServiceProvider { return serviceauthnoop.NewProvider() },
	}
	circuitBreakerProviderFactories = providerFactoryRegistry{
		"sentinel": func() runtimecontract.ServiceProvider { return circuitbreakersentinel.NewProvider() },
		"noop":     func() runtimecontract.ServiceProvider { return circuitbreakernoop.NewProvider() },
		"":         func() runtimecontract.ServiceProvider { return circuitbreakernoop.NewProvider() },
	}
	dtmProviderFactories = providerFactoryRegistry{
		"sdk":    func() runtimecontract.ServiceProvider { return dtmsdk.NewProvider() },
		"dtmsdk": func() runtimecontract.ServiceProvider { return dtmsdk.NewProvider() },
		"noop":   func() runtimecontract.ServiceProvider { return dtmnoop.NewProvider() },
		"":       func() runtimecontract.ServiceProvider { return dtmnoop.NewProvider() },
	}
	messageQueueProviderFactories = providerFactoryRegistry{
		"redis": func() runtimecontract.ServiceProvider { return mqredis.NewProvider() },
		"noop":  func() runtimecontract.ServiceProvider { return mqnoop.NewProvider() },
		"":      func() runtimecontract.ServiceProvider { return mqnoop.NewProvider() },
	}
	distributedLockProviderFactories = providerFactoryRegistry{
		"redis": func() runtimecontract.ServiceProvider { return dlockredis.NewProvider() },
		"noop":  func() runtimecontract.ServiceProvider { return dlocknoop.NewProvider() },
		"":      func() runtimecontract.ServiceProvider { return dlocknoop.NewProvider() },
	}
)

func providerFromMap(factories map[string]providerFactory, key string, fallback string) runtimecontract.ServiceProvider {
	if factory, ok := factories[key]; ok {
		return factory()
	}
	if factory, ok := factories[fallback]; ok {
		return factory()
	}
	return nil
}

func defaultTracingProvider() runtimecontract.ServiceProvider {
	return tracingnoop.NewProvider()
}

func enabledTracingProvider() runtimecontract.ServiceProvider {
	return tracingotel.NewProvider()
}

func defaultMetadataProvider() runtimecontract.ServiceProvider {
	return metadatanoop.NewProvider()
}

func enabledMetadataProvider() runtimecontract.ServiceProvider {
	return metadatadefault.NewProvider()
}
