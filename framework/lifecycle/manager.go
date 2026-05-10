// Package lifecycle provides service lifecycle management for gorp framework.
// Manages lifecycle of multiple host-managed services in one place.
// Provides startup/stop helper with priority ordering and lifecycle hooks.
//
// 生命周期包提供 gorp 框架的服务生命周期管理能力。
// 在一个位置统一管理多个被 host 托管的服务生命周期。
// 提供带优先级排序和生命周期钩子的通用启动/停止 helper。
package lifecycle

import (
	"context"
	"sync"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Manager manages service lifecycle transitions in one place.
//
// Manager 统一管理服务生命周期流转。
//
// 中文说明：
// - 这是 “lifecycle helper 更彻底统一” 的核心实现。
// - 统一管理服务的启动、停止、优雅关闭。
// - 支持生命周期钩子：OnStarting、OnStarted、OnStopping、OnStopped。
// - 可被 Host 或命令层直接使用。
type Manager struct {
	services []ServiceEntry
	mu       sync.RWMutex
	state    State
}

// ServiceEntry describes one registered service entry.
//
// ServiceEntry 表示一个已注册的服务条目。
//
// 中文说明：
// - Name 用于生命周期日志、排障和服务列表展示。
// - Service 是真正执行 Start/Stop 的运行对象。
// - Hooks 用于在服务前后挂生命周期钩子。
// - Priority 决定启动顺序和停止逆序。
type ServiceEntry struct {
	Name     string
	Service  runtimecontract.Hostable
	Hooks    runtimecontract.Lifecycle
	Priority int // 启动优先级，数值小的先启动。
}

// State represents the current lifecycle state.
//
// State 表示当前生命周期状态。
type State int

const (
	StateIdle State = iota
	StateStarting
	StateRunning
	StateStopping
	StateStopped
)

// String returns the string form of the lifecycle state.
//
// String 返回生命周期状态的字符串表示。
func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateStarting:
		return "starting"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// NewManager creates a lifecycle manager.
//
// NewManager 创建生命周期管理器。
//
// 中文说明：
// - 初始状态为 StateIdle。
// - 默认不注册任何服务，等待 Host 或业务侧逐步填充。
func NewManager() *Manager {
	return &Manager{
		services: make([]ServiceEntry, 0),
		state:    StateIdle,
	}
}

// Register registers one service into the lifecycle manager.
//
// Register 注册一个由生命周期管理器托管的服务。
//
// 中文说明：
// - priority 数值小的先启动，后停止。
// - hooks 可以为 nil，适用于不需要生命周期钩子的服务。
func (m *Manager) Register(name string, service runtimecontract.Hostable, hooks runtimecontract.Lifecycle, priority int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.services = append(m.services, ServiceEntry{
		Name:     name,
		Service:  service,
		Hooks:    hooks,
		Priority: priority,
	})
}

// Start starts all registered services in priority order.
//
// Start 按优先级顺序启动所有已注册服务。
//
// 中文说明：
// - 按 priority 从小到大启动服务。
// - 触发 OnStarting 和 OnStarted 钩子。
// - 如果某个服务启动失败，会回滚已启动的服务。
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.state != StateIdle {
		m.mu.Unlock()
		return nil
	}
	m.state = StateStarting
	m.mu.Unlock()

	sorted := m.sortedServices()
	started := make([]ServiceEntry, 0)

	for _, entry := range sorted {
		if entry.Hooks != nil {
			if err := entry.Hooks.OnStarting(ctx); err != nil {
				_ = m.stopReverse(ctx, started)
				m.mu.Lock()
				m.state = StateIdle
				m.mu.Unlock()
				return err
			}
		}

		if err := entry.Service.Start(ctx); err != nil {
			_ = m.stopReverse(ctx, started)
			m.mu.Lock()
			m.state = StateIdle
			m.mu.Unlock()
			return err
		}

		if entry.Hooks != nil {
			if err := entry.Hooks.OnStarted(ctx); err != nil {
				_ = m.stopReverse(ctx, started)
				m.mu.Lock()
				m.state = StateIdle
				m.mu.Unlock()
				return err
			}
		}

		started = append(started, entry)
	}

	m.mu.Lock()
	m.state = StateRunning
	m.mu.Unlock()
	return nil
}

// Stop stops all registered services in reverse priority order.
//
// Stop 按优先级逆序停止所有已注册服务。
//
// 中文说明：
// - 按 priority 从大到小停止服务。
// - 触发 OnStopping 和 OnStopped 钩子。
// - 即使某个服务停止失败，也继续停止其他服务。
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	if m.state != StateRunning {
		m.mu.Unlock()
		return nil
	}
	m.state = StateStopping
	m.mu.Unlock()

	sorted := m.sortedServices()
	var lastErr error

	for i := len(sorted) - 1; i >= 0; i-- {
		entry := sorted[i]

		if entry.Hooks != nil {
			if err := entry.Hooks.OnStopping(ctx); err != nil {
				lastErr = err
			}
		}

		if err := entry.Service.Stop(ctx); err != nil {
			lastErr = err
		}

		if entry.Hooks != nil {
			if err := entry.Hooks.OnStopped(ctx); err != nil {
				lastErr = err
			}
		}
	}

	m.mu.Lock()
	m.state = StateStopped
	m.mu.Unlock()
	return lastErr
}

// State returns the current lifecycle state.
//
// State 返回当前生命周期状态。
//
// 中文说明：
// - 这是 Host 和外层运行时观察生命周期进度的读取入口。
// - 仅读取当前状态，不改变任何生命周期行为。
func (m *Manager) State() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// Services returns the names of all registered services.
//
// Services 返回所有已注册服务的名称列表。
//
// 中文说明：
// - 这里只返回注册名，不暴露内部 ServiceEntry 细节。
// - 主要供 Host 查询当前托管了哪些服务。
func (m *Manager) Services() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, len(m.services))
	for i, s := range m.services {
		names[i] = s.Name
	}
	return names
}

// sortedServices returns a priority-sorted copy of the service list.
//
// sortedServices 返回按 priority 排序后的服务副本。
//
// 中文说明：
// - 返回的是副本，不会改动原始注册顺序。
// - 统一保证 Start 用正序、Stop 用逆序时看到的是同一套优先级结果。
func (m *Manager) sortedServices() []ServiceEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sorted := make([]ServiceEntry, len(m.services))
	copy(sorted, m.services)

	// Keep ordering logic local and deterministic so startup/shutdown semantics stay stable.
	// 把排序逻辑收口在这里，确保启动/停止语义始终稳定可预测。
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Priority > sorted[j].Priority {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

// stopReverse stops the already-started services in reverse order.
//
// stopReverse 按逆序停止已经成功启动过的服务。
//
// 中文说明：
// - 这是启动失败后的回滚辅助函数。
// - 只处理已经成功启动过的服务，避免把未启动服务也纳入停止流程。
// - 即使中途出错，也继续尽量把后续服务停掉。
func (m *Manager) stopReverse(ctx context.Context, services []ServiceEntry) error {
	var lastErr error
	for i := len(services) - 1; i >= 0; i-- {
		entry := services[i]
		if entry.Hooks != nil {
			_ = entry.Hooks.OnStopping(ctx)
		}
		if err := entry.Service.Stop(ctx); err != nil {
			lastErr = err
		}
		if entry.Hooks != nil {
			_ = entry.Hooks.OnStopped(ctx)
		}
	}
	return lastErr
}
