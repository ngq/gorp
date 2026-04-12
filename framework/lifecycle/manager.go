package lifecycle

import (
	"context"
	"sync"

	"github.com/ngq/gorp/framework/contract"
)

// Manager 统一管理服务生命周期。
//
// 中文说明：
// - 这是"lifecycle helper 更彻底统一"的核心实现；
// - 统一管理服务的启动、停止、优雅关闭；
// - 支持生命周期钩子（OnStarting/OnStarted/OnStopping/OnStopped）；
// - 可被 Host 或命令层直接使用。
type Manager struct {
	services []ServiceEntry
	mu       sync.RWMutex
	state    State
}

// ServiceEntry 表示一个已注册的服务条目。
type ServiceEntry struct {
	Name     string
	Service  contract.Hostable
	Hooks    contract.Lifecycle
	Priority int // 启动优先级，数值小的先启动
}

// State 表示生命周期状态。
type State int

const (
	StateIdle State = iota
	StateStarting
	StateRunning
	StateStopping
	StateStopped
)

// String 返回状态的字符串表示。
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

// NewManager 创建生命周期管理器。
func NewManager() *Manager {
	return &Manager{
		services: make([]ServiceEntry, 0),
		state:    StateIdle,
	}
}

// Register 注册服务。
//
// 中文说明：
// - 注册一个可被生命周期管理的服务；
// - priority 数值小的先启动，后停止；
// - hooks 可以为 nil，如果服务不需要生命周期钩子。
func (m *Manager) Register(name string, service contract.Hostable, hooks contract.Lifecycle, priority int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.services = append(m.services, ServiceEntry{
		Name:     name,
		Service:  service,
		Hooks:    hooks,
		Priority: priority,
	})
}

// Start 启动所有服务。
//
// 中文说明：
// - 按 priority 从小到大启动服务；
// - 触发 OnStarting 和 OnStarted 钩子；
// - 如果某个服务启动失败，停止已启动的服务并返回错误。
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.state != StateIdle {
		m.mu.Unlock()
		return nil
	}
	m.state = StateStarting
	m.mu.Unlock()

	// 按 priority 排序（小的在前）
	sorted := m.sortedServices()

	// 按顺序启动
	started := make([]ServiceEntry, 0)
	for _, entry := range sorted {
		// 触发 OnStarting 钩子
		if entry.Hooks != nil {
			if err := entry.Hooks.OnStarting(ctx); err != nil {
				_ = m.stopReverse(ctx, started)
				m.mu.Lock()
				m.state = StateIdle
				m.mu.Unlock()
				return err
			}
		}

		// 启动服务
		if err := entry.Service.Start(ctx); err != nil {
			_ = m.stopReverse(ctx, started)
			m.mu.Lock()
			m.state = StateIdle
			m.mu.Unlock()
			return err
		}

		// 触发 OnStarted 钩子
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

// Stop 停止所有服务。
//
// 中文说明：
// - 按 priority 从大到小停止服务；
// - 触发 OnStopping 和 OnStopped 钩子；
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

	// 逆序停止
	for i := len(sorted) - 1; i >= 0; i-- {
		entry := sorted[i]

		// 触发 OnStopping 钩子
		if entry.Hooks != nil {
			if err := entry.Hooks.OnStopping(ctx); err != nil {
				lastErr = err
			}
		}

		// 停止服务
		if err := entry.Service.Stop(ctx); err != nil {
			lastErr = err
		}

		// 触发 OnStopped 钩子
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

// State 返回当前状态。
func (m *Manager) State() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// Services 返回所有已注册的服务名称。
func (m *Manager) Services() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, len(m.services))
	for i, s := range m.services {
		names[i] = s.Name
	}
	return names
}

// sortedServices 返回按 priority 排序的服务列表。
func (m *Manager) sortedServices() []ServiceEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sorted := make([]ServiceEntry, len(m.services))
	copy(sorted, m.services)

	// 简单冒泡排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Priority > sorted[j].Priority {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

// stopReverse 逆序停止服务。
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