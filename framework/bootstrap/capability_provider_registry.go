// Application scenarios:
// - Maintain the provider factory registries used by bootstrap selection logic.
// - Allow built-in and contributed capability providers to be resolved from config values.
// - Support optional factory extension points without hard-coding every provider choice into selectors.
//
// 适用场景：
// - 维护 bootstrap 选择逻辑所依赖的 provider factory 注册表。
// - 让内建和扩展能力 provider 可以通过配置值解析出来。
// - 支持可选 factory 扩展点，而不是把所有 provider 选择写死在 selector 里。
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

// RegisterConfigSourceProviderFactory registers a config-source provider factory.
//
// RegisterConfigSourceProviderFactory 注册配置源 provider factory。
func RegisterConfigSourceProviderFactory(key string, factory providerFactory) {
	configSourceProviderFactories.register(key, factory)
}

// RegisterDiscoveryProviderFactory registers a discovery provider factory.
//
// RegisterDiscoveryProviderFactory 注册服务发现 provider factory。
func RegisterDiscoveryProviderFactory(key string, factory providerFactory) {
	discoveryProviderFactories.register(key, factory)
}

// RegisterSelectorProviderFactory registers a selector provider factory.
//
// RegisterSelectorProviderFactory 注册选择器 provider factory。
func RegisterSelectorProviderFactory(key string, factory providerFactory) {
	selectorProviderFactories.register(key, factory)
}

// RegisterRPCProviderFactory registers an RPC provider factory.
//
// RegisterRPCProviderFactory 注册 RPC provider factory。
func RegisterRPCProviderFactory(key string, factory providerFactory) {
	rpcProviderFactories.register(key, factory)
}

// RegisterServiceAuthProviderFactory registers a service-auth provider factory.
//
// RegisterServiceAuthProviderFactory 注册服务鉴权 provider factory。
func RegisterServiceAuthProviderFactory(key string, factory providerFactory) {
	serviceAuthProviderFactories.register(key, factory)
}

// RegisterCircuitBreakerProviderFactory registers a circuit-breaker provider factory.
//
// RegisterCircuitBreakerProviderFactory 注册熔断器 provider factory。
func RegisterCircuitBreakerProviderFactory(key string, factory providerFactory) {
	circuitBreakerProviderFactories.register(key, factory)
}

// RegisterDTMProviderFactory registers a DTM provider factory.
//
// RegisterDTMProviderFactory 注册 DTM provider factory。
func RegisterDTMProviderFactory(key string, factory providerFactory) {
	dtmProviderFactories.register(key, factory)
}

// RegisterMessageQueueProviderFactory registers a message-queue provider factory.
//
// RegisterMessageQueueProviderFactory 注册消息队列 provider factory。
func RegisterMessageQueueProviderFactory(key string, factory providerFactory) {
	messageQueueProviderFactories.register(key, factory)
}

// RegisterTracingProviderFactory registers a tracing provider factory.
//
// RegisterTracingProviderFactory 注册 tracing provider factory。
func RegisterTracingProviderFactory(key string, factory providerFactory) {
	tracingProviderFactories.register(key, factory)
}

// RegisterMetadataProviderFactory registers a metadata provider factory.
//
// RegisterMetadataProviderFactory 注册 metadata provider factory。
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
	// Fall back to the default mapping when the requested backend is absent.
	// 当请求后端不存在时，回退到默认映射。
	if factory, ok := factories[fallback]; ok {
		return factory()
	}
	return nil
}

// defaultTracingProvider returns the default tracing provider used by bootstrap fallbacks.
//
// defaultTracingProvider 返回 bootstrap 回退路径使用的默认 tracing provider。
func defaultTracingProvider() runtimecontract.ServiceProvider {
	return tracingnoop.NewProvider()
}

// enabledTracingProvider returns the tracing provider used when tracing is explicitly enabled.
//
// enabledTracingProvider 返回 tracing 显式启用时使用的 provider。
func enabledTracingProvider() runtimecontract.ServiceProvider {
	return tracingotel.NewProvider()
}

// defaultMetadataProvider returns the default metadata provider used by bootstrap fallbacks.
//
// defaultMetadataProvider 返回 bootstrap 回退路径使用的默认 metadata provider。
func defaultMetadataProvider() runtimecontract.ServiceProvider {
	return metadatanoop.NewProvider()
}

// enabledMetadataProvider returns the metadata provider used when propagation is explicitly enabled.
//
// enabledMetadataProvider 返回 metadata 透传显式启用时使用的 provider。
func enabledMetadataProvider() runtimecontract.ServiceProvider {
	return metadatadefault.NewProvider()
}
