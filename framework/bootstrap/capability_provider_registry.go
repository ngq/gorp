// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file maintains provider factory registries for bootstrap selection logic.
// Allows built-in and contributed capability providers to be resolved from config.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件维护 bootstrap 选择逻辑所依赖的 provider factory 注册表。
// 让内建和扩展能力 provider 可以通过配置值解析出来。
package bootstrap

import (
	"fmt"
	"sort"
	"strings"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	circuitbreakernoop "github.com/ngq/gorp/framework/provider/circuitbreaker/noop"
	configsourcelocal "github.com/ngq/gorp/framework/provider/configsource/local"
	configsourcenoop "github.com/ngq/gorp/framework/provider/configsource/noop"
	discoverynoop "github.com/ngq/gorp/framework/provider/discovery/noop"
	dlocknoop "github.com/ngq/gorp/framework/provider/dlock/noop"
	dtmnoop "github.com/ngq/gorp/framework/provider/dtm/noop"
	loadsheddingprovider "github.com/ngq/gorp/framework/provider/loadshedding"
	loadsheddingnoop "github.com/ngq/gorp/framework/provider/loadshedding/noop"
	mqnoop "github.com/ngq/gorp/framework/provider/messagequeue/noop"
	metadatadefault "github.com/ngq/gorp/framework/provider/metadata"
	metadatanoop "github.com/ngq/gorp/framework/provider/metadata/noop"
	retryprovider "github.com/ngq/gorp/framework/provider/retry"
	retrynoop "github.com/ngq/gorp/framework/provider/retry/noop"
	rpcgrpc "github.com/ngq/gorp/framework/provider/rpc/grpc"
	rpchttp "github.com/ngq/gorp/framework/provider/rpc/http"
	rpcnoop "github.com/ngq/gorp/framework/provider/rpc/noop"
	selectornoop "github.com/ngq/gorp/framework/provider/selector/noop"
	selectorp2c "github.com/ngq/gorp/framework/provider/selector/p2c"
	selectorrandom "github.com/ngq/gorp/framework/provider/selector/random"
	selectorwrr "github.com/ngq/gorp/framework/provider/selector/wrr"
	serviceauthnoop "github.com/ngq/gorp/framework/provider/serviceauth/noop"
	tracingnoop "github.com/ngq/gorp/framework/provider/tracing/noop"
	wsnoop "github.com/ngq/gorp/framework/provider/websocket/noop"
)

type providerFactory func() runtimecontract.ServiceProvider

// providerFactoryRegistry maps backend names to provider factory functions.
// Note: This is a map type and NOT safe for concurrent writes. All RegisterXxxProviderFactory
// calls must happen during init() or early main() before any goroutines start.
// If concurrent registration is needed, use sync.Map or add mutex protection.
//
// providerFactoryRegistry 将后端名称映射到 provider 工厂函数。
// 注意：这是 map 类型，并发写入不安全。所有 RegisterXxxProviderFactory 调用
// 必须在 init() 或 main() 早期、任何 goroutine 启动前完成。
// 如果需要并发注册，请使用 sync.Map 或添加互斥锁保护。
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

// RegisterLoadShedderProviderFactory registers a load-shedder provider factory.
//
// RegisterLoadShedderProviderFactory 注册过载保护 provider factory。
func RegisterLoadShedderProviderFactory(key string, factory providerFactory) {
	loadShedderProviderFactories.register(key, factory)
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

// RegisterRetryProviderFactory registers a retry provider factory.
//
// RegisterRetryProviderFactory 注册重试 provider factory。
func RegisterRetryProviderFactory(key string, factory providerFactory) {
	retryProviderFactories.register(key, factory)
}

// RegisterWebSocketProviderFactory registers a WebSocket provider factory.
//
// RegisterWebSocketProviderFactory 注册 WebSocket provider factory。
func RegisterWebSocketProviderFactory(key string, factory providerFactory) {
	webSocketProviderFactories.register(key, factory)
}

// RegisterDistributedLockProviderFactory registers a distributed-lock provider factory.
//
// RegisterDistributedLockProviderFactory 注册分布式锁 provider factory。
func RegisterDistributedLockProviderFactory(key string, factory providerFactory) {
	distributedLockProviderFactories.register(key, factory)
}

var (
	// 内建 provider（noop 和 framework provider）
	configSourceProviderFactories = providerFactoryRegistry{
		"noop":  func() runtimecontract.ServiceProvider { return configsourcenoop.NewProvider() },
		"local": func() runtimecontract.ServiceProvider { return configsourcelocal.NewProvider() },
		"":      func() runtimecontract.ServiceProvider { return configsourcelocal.NewProvider() },
	}
	discoveryProviderFactories = providerFactoryRegistry{
		"noop": func() runtimecontract.ServiceProvider { return discoverynoop.NewProvider() },
		"":     func() runtimecontract.ServiceProvider { return discoverynoop.NewProvider() },
	}
	selectorProviderFactories = providerFactoryRegistry{
		"random":   func() runtimecontract.ServiceProvider { return selectorrandom.NewProvider() },
		"wrr":      func() runtimecontract.ServiceProvider { return selectorwrr.NewProvider() },
		"p2c":      func() runtimecontract.ServiceProvider { return selectorp2c.NewProvider() },
		"p2c_ewma": func() runtimecontract.ServiceProvider { return selectorp2c.NewProvider() },
		"noop":     func() runtimecontract.ServiceProvider { return selectornoop.NewProvider() },
		"":         func() runtimecontract.ServiceProvider { return selectornoop.NewProvider() },
	}
	rpcProviderFactories = providerFactoryRegistry{
		"http": func() runtimecontract.ServiceProvider { return rpchttp.NewProvider() },
		"grpc": func() runtimecontract.ServiceProvider { return rpcgrpc.NewProvider() },
		"noop": func() runtimecontract.ServiceProvider { return rpcnoop.NewProvider() },
		"":     func() runtimecontract.ServiceProvider { return rpcnoop.NewProvider() },
	}
	tracingProviderFactories = providerFactoryRegistry{
		"noop": func() runtimecontract.ServiceProvider { return tracingnoop.NewProvider() },
		"":     func() runtimecontract.ServiceProvider { return tracingnoop.NewProvider() },
	}
	metadataProviderFactories = providerFactoryRegistry{
		"default": func() runtimecontract.ServiceProvider { return metadatadefault.NewProvider() },
		"noop":    func() runtimecontract.ServiceProvider { return metadatanoop.NewProvider() },
		"":        func() runtimecontract.ServiceProvider { return metadatanoop.NewProvider() },
	}
	serviceAuthProviderFactories = providerFactoryRegistry{
		"noop": func() runtimecontract.ServiceProvider { return serviceauthnoop.NewProvider() },
		"":     func() runtimecontract.ServiceProvider { return serviceauthnoop.NewProvider() },
	}
	circuitBreakerProviderFactories = providerFactoryRegistry{
		"noop": func() runtimecontract.ServiceProvider { return circuitbreakernoop.NewProvider() },
		"":     func() runtimecontract.ServiceProvider { return circuitbreakernoop.NewProvider() },
	}
	loadShedderProviderFactories = providerFactoryRegistry{
		"semaphore": func() runtimecontract.ServiceProvider { return loadsheddingprovider.NewProvider() },
		"noop":      func() runtimecontract.ServiceProvider { return loadsheddingnoop.NewProvider() },
		"":          func() runtimecontract.ServiceProvider { return loadsheddingnoop.NewProvider() },
	}
	dtmProviderFactories = providerFactoryRegistry{
		"noop": func() runtimecontract.ServiceProvider { return dtmnoop.NewProvider() },
		"":     func() runtimecontract.ServiceProvider { return dtmnoop.NewProvider() },
	}
	messageQueueProviderFactories = providerFactoryRegistry{
		"noop": func() runtimecontract.ServiceProvider { return mqnoop.NewProvider() },
		"":     func() runtimecontract.ServiceProvider { return mqnoop.NewProvider() },
	}
	distributedLockProviderFactories = providerFactoryRegistry{
		"noop": func() runtimecontract.ServiceProvider { return dlocknoop.NewProvider() },
		"":     func() runtimecontract.ServiceProvider { return dlocknoop.NewProvider() },
	}
	retryProviderFactories = providerFactoryRegistry{
		"default": func() runtimecontract.ServiceProvider { return retryprovider.NewProvider() },
		"noop":    func() runtimecontract.ServiceProvider { return retrynoop.NewProvider() },
		"":        func() runtimecontract.ServiceProvider { return retrynoop.NewProvider() },
	}
	webSocketProviderFactories = providerFactoryRegistry{
		"noop": func() runtimecontract.ServiceProvider { return wsnoop.NewProvider() },
		"":     func() runtimecontract.ServiceProvider { return wsnoop.NewProvider() },
	}
)

// providerFromMap 按 key 选择 provider factory。
//
// 日志与失败策略（按用户决策）：
// - 未配置（key 为空）或显式默认（key == fallback）：用默认后端，静默。
// - 显式配置了后端且可用：打印实际生效的后端（便于启动时确认配置）。
// - 配置了后端但不可用（拼错 / provider 未 blank import）：**fail-fast**，
//   注册时返回明确错误使启动失败，绝不静默降级。
func providerFromMap(factories map[string]providerFactory, key string, fallback string) runtimecontract.ServiceProvider {
	key = strings.TrimSpace(key)
	if factory, ok := factories[key]; ok {
		// 显式配置了后端且可用：打印实际生效的后端。
		if key != "" && key != fallback {
			fmt.Printf("[gorp] capability backend: %q\n", key)
		}
		return factory()
	}
	// 未配置或显式默认：用默认后端，静默。
	if key == "" || key == fallback {
		if factory, ok := factories[fallback]; ok {
			return factory()
		}
		return nil
	}
	// 配置了后端但不可用：fail-fast，不静默降级到 noop。
	return &failingProvider{requested: key, available: factoryKeys(factories)}
}

// failingProvider 表示配置了未知/未注册的后端。Register 返回明确错误，
// 使启动失败，避免"配置了 sentinel/etcd 却悄悄跑在 noop 上"的静默失效。
type failingProvider struct {
	requested string
	available []string
}

func (p *failingProvider) Name() string { return "capability.invalid" }
func (p *failingProvider) Register(c runtimecontract.Container) error {
	return fmt.Errorf("capability backend %q is not registered; check the config value or blank-import the provider package (available: %s)", p.requested, strings.Join(p.available, ", "))
}
func (p *failingProvider) Boot(runtimecontract.Container) error { return nil }
func (p *failingProvider) IsDefer() bool                       { return false }
func (p *failingProvider) Provides() []string                  { return nil }
func (p *failingProvider) DependsOn() []string                 { return nil }

// factoryKeys 返回工厂 map 中已注册的后端名（排序，用于错误提示）。
func factoryKeys(factories map[string]providerFactory) []string {
	keys := make([]string, 0, len(factories))
	for k := range factories {
		if k != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// defaultTracingProvider returns the default tracing provider used by bootstrap fallbacks.
//
// defaultTracingProvider 返回 bootstrap 回退路径使用的默认 tracing provider。
func defaultTracingProvider() runtimecontract.ServiceProvider {
	return tracingnoop.NewProvider()
}

// defaultMetadataProvider returns the default metadata provider used by bootstrap fallbacks.
//
// defaultMetadataProvider 返回 bootstrap 回退路径使用的默认 metadata provider。
func defaultMetadataProvider() runtimecontract.ServiceProvider {
	return metadatanoop.NewProvider()
}
