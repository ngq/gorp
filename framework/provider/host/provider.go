package host

import (
	"context"
	"sync"

	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/lifecycle"
)

// Provider 是 Host 服务的提供者。
type Provider struct{}

// NewProvider 创建 Host 服务提供者。
func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string       { return "host" }
func (p *Provider) IsDefer() bool      { return false }
func (p *Provider) Provides() []string { return []string{contract.HostKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.HostKey, func(c contract.Container) (any, error) {
		return NewDefaultHost(c), nil
	}, true)
	return nil
}

func (p *Provider) Boot(contract.Container) error { return nil }

// DefaultHost 是 Host 接口的默认实现。
//
// 中文说明：
// - 统一管理应用的生命周期；
// - 支持注册多个 Hostable 服务；
// - 按优先级启动，按逆序关闭；
// - 使用 lifecycle.Manager 进行统一的生命周期管理。
type DefaultHost struct {
	container contract.Container
	manager   *lifecycle.Manager
	mu        sync.RWMutex
	running   bool
}

// NewDefaultHost 创建默认 Host 实现。
func NewDefaultHost(container contract.Container) *DefaultHost {
	return &DefaultHost{
		container: container,
		manager:   lifecycle.NewManager(),
	}
}

// RegisterService 注册服务。
//
// 中文说明：
// - 将服务注册到生命周期管理器；
// - 默认优先级为 100。
func (h *DefaultHost) RegisterService(name string, service contract.Hostable) error {
	return h.RegisterServiceWithPriority(name, service, nil, 100)
}

// RegisterServiceWithPriority 注册带优先级的服务。
//
// 中文说明：
// - priority 数值小的先启动，后停止；
// - hooks 可以为 nil；
// - HTTP 服务建议优先级 10；
// - gRPC 服务建议优先级 20；
// - Cron 服务建议优先级 30。
func (h *DefaultHost) RegisterServiceWithPriority(name string, service contract.Hostable, hooks contract.Lifecycle, priority int) error {
	h.manager.Register(name, service, hooks, priority)
	return nil
}

// Services 返回所有已注册的服务名称。
func (h *DefaultHost) Services() []string {
	return h.manager.Services()
}

// State 返回当前生命周期状态。
func (h *DefaultHost) State() lifecycle.State {
	return h.manager.State()
}

// Start 启动所有服务。
//
// 中文说明：
// - 使用 lifecycle.Manager 统一管理启动流程；
// - 按优先级启动所有服务。
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

// Stop 停止所有服务。
//
// 中文说明：
// - 使用 lifecycle.Manager 统一管理停止流程；
// - 按优先级逆序停止所有服务。
func (h *DefaultHost) Stop(ctx context.Context) error {
	return h.Shutdown(ctx)
}

// Shutdown 触发优雅关闭。
//
// 中文说明：
// - 使用 lifecycle.Manager 统一管理关闭流程；
// - 触发所有服务的生命周期钩子。
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