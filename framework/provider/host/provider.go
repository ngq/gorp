// Package host provides the application host service for managing service lifecycle.
// The host manages startup/shutdown of HTTP servers, Cron jobs, and GRPC servers.
// Services can be registered with priority for startup ordering.
//
// Host 服务包，提供应用主机服务，管理多个服务的生命周期。
// Host 管理 HTTP 服务器、Cron 任务和 GRPC 服务器的启动/关闭。
// 服务可按优先级注册，控制启动顺序。
// Eg:
//
//	// 注册 Provider
//	app.Register(host.NewProvider())
//
//	// 注册服务
//	h := c.MustMake(runtimecontract.HostKey).(runtimecontract.Host)
//	h.RegisterService("http", httpService)
//	h.RegisterService("cron", cronService)
//
//	// 启动所有服务
//	h.Start(ctx)
//
//	// 关闭所有服务
//	h.Shutdown(ctx)
package host

import (
	"context"
	"sync"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/lifecycle"
)

// Provider registers the host service contract.
//
// Provider 注册主机服务契约。
type Provider struct{}

// NewProvider creates a new host provider instance.
//
// NewProvider 创建新的主机 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "host".
//
// Name 返回 Provider 名称 "host"。
func (p *Provider) Name() string { return "host" }

// IsDefer returns false, host should be initialized immediately for service registration.
//
// IsDefer 返回 false，主机应立即初始化以便服务注册。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the host contract key.
//
// Provides 返回主机契约键。
func (p *Provider) Provides() []string { return []string{runtimecontract.HostKey} }

// DependsOn returns the keys this provider depends on.
// Host provider has no dependencies.
//
// DependsOn 返回该 provider 依赖的 key。
// Host provider 无依赖。
func (p *Provider) DependsOn() []string { return nil }

// Register binds the host service factory to the container.
//
// Register 将主机服务工厂绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(runtimecontract.HostKey, func(c runtimecontract.Container) (any, error) {
		return NewDefaultHost(c), nil
	}, true)
	return nil
}

// Boot is a no-op for host provider.
//
// Boot 主机 Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// DefaultHost implements runtimecontract.Host interface.
//
// DefaultHost 实现 runtimecontract.Host 接口，管理多个服务的生命周期。
type DefaultHost struct {
	container runtimecontract.Container // container is the DI container.
	                                   //
	                                    // container DI 容器。
	manager   *lifecycle.Manager         // manager is the lifecycle manager.
	                                   //
	                                    // manager 生命周期管理器。
	mu        sync.RWMutex               // mu protects running state.
	                                   //
	                                    // mu 保护运行状态。
	running   bool                       // running indicates if host is started.
	                                   //
	                                    // running 标记主机是否已启动。
}

// NewDefaultHost creates a new host instance with the given container.
//
// NewDefaultHost 根据给定容器创建新的主机实例。
func NewDefaultHost(container runtimecontract.Container) *DefaultHost {
	return &DefaultHost{
		container: container,
		manager:   lifecycle.NewManager(),
	}
}

// RegisterService registers a service with default priority (100).
//
// RegisterService 注册服务，使用默认优先级（100）。
func (h *DefaultHost) RegisterService(name string, service runtimecontract.Hostable) error {
	return h.RegisterServiceWithPriority(name, service, nil, 100)
}

// RegisterServiceWithPriority registers a service with custom priority and lifecycle hooks.
//
// RegisterServiceWithPriority 注册服务，使用自定义优先级和生命周期钩子。
//
// Lower priority numbers start earlier, higher numbers start later.
//
// 优先级数字越小越早启动，数字越大越晚启动。
func (h *DefaultHost) RegisterServiceWithPriority(name string, service runtimecontract.Hostable, hooks runtimecontract.Lifecycle, priority int) error {
	h.manager.Register(name, service, hooks, priority)
	return nil
}

// Services returns all registered service names.
//
// Services 返回所有已注册的服务名称。
func (h *DefaultHost) Services() []string {
	return h.manager.Services()
}

// State returns the current lifecycle state of the host.
//
// State 返回主机当前的生命周期状态。
func (h *DefaultHost) State() lifecycle.State {
	return h.manager.State()
}

// Start starts all registered services in priority order.
// Core logic: Check if already running, then call manager.Start.
//
// Start 按优先级顺序启动所有已注册的服务。
// 核心逻辑：检查是否已运行，然后调用 manager.Start。
func (h *DefaultHost) Start(ctx context.Context) error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return nil
	}
	h.running = true
	h.mu.Unlock()

	return h.manager.Start(ctx)
}

// Stop shuts down all registered services.
//
// Stop 关闭所有已注册的服务。
func (h *DefaultHost) Stop(ctx context.Context) error {
	return h.Shutdown(ctx)
}

// Shutdown gracefully stops all registered services in reverse priority order.
// Core logic: Check running state, call manager.Stop, update running flag.
//
// Shutdown 按优先级逆序优雅关闭所有已注册的服务。
// 核心逻辑：检查运行状态，调用 manager.Stop，更新运行标志。
func (h *DefaultHost) Shutdown(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return nil
	}

	err := h.manager.Stop(ctx)
	h.running = false
	return err
}