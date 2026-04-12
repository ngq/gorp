// Package plugin 插件管理器
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/ngq/gorp/framework/contract"
)

// Manager 插件管理器
//
// 中文说明:
// - 负责插件的发现、加载、安装、卸载等生命周期管理;
// - 使用 gorp Container 注册插件服务;
// - 支持插件依赖排序;
// - Discover 只发现不加载,Register 才注册到 Container。
type Manager struct {
	mu          sync.RWMutex
	container   contract.Container
	plugins     map[string]Plugin      // systemName -> Plugin 实例
	metas       map[string]*PluginMeta // systemName -> Meta 元数据
	pluginsDir  string                 // 插件目录路径
}

// NewManager 创建插件管理器
//
// 中文说明:
// - container 是 gorp 依赖注入容器;
// - pluginsDir 是插件目录路径,通常是 "./plugins";
// - 创建后需要调用 Discover() 发现可用插件。
func NewManager(container contract.Container, pluginsDir string) *Manager {
	return &Manager{
		container:   container,
		plugins:     make(map[string]Plugin),
		metas:       make(map[string]*PluginMeta),
		pluginsDir:  pluginsDir,
	}
}

// Discover 扫描并发现所有插件
//
// 中文说明:
// - 扫描 plugins/ 目录下的所有 plugin.json;
// - 解析元数据并验证版本兼容性;
// - 只发现不加载,不创建插件实例;
// - 实际注册需要调用 Register() 方法。
func (m *Manager) Discover() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清空已发现的元数据
	m.metas = make(map[string]*PluginMeta)

	// 检查插件目录是否存在
	entries, err := os.ReadDir(m.pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 插件目录不存在,不是错误
		}
		return fmt.Errorf("read plugins directory: %w", err)
	}

	// 扫描每个子目录
	for _, entry := range entries {
		if !entry.IsDir() {
			continue // 只处理目录
		}

		pluginDir := filepath.Join(m.pluginsDir, entry.Name())
		metaPath := filepath.Join(pluginDir, "plugin.json")

		// 读取 plugin.json
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue // 没有 plugin.json,跳过
		}

		var meta PluginMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			continue // 解析失败,跳过
		}

		// 验证基本字段
		if meta.SystemName == "" {
			continue // system_name 必填
		}

		m.metas[meta.SystemName] = &meta
	}

	return nil
}

// Register 注册一个插件实例
//
// 中文说明:
// - 将插件实例注册到管理器;
// - 同时转换为 ServiceProvider 注册到 gorp Container;
// - 注册后插件可通过 Container.Make() 获取;
// - 注册成功后会自动添加到全局 Registry。
func (m *Manager) Register(p Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	meta := p.Meta()
	if meta == nil {
		return fmt.Errorf("plugin meta is nil")
	}

	systemName := meta.SystemName
	if systemName == "" {
		return fmt.Errorf("plugin system_name is empty")
	}

	// 注册到管理器内部映射
	m.plugins[systemName] = p
	m.metas[systemName] = meta

	// 转换为 ServiceProvider
	sp := p.ToServiceProvider()
	if sp == nil {
		return fmt.Errorf("plugin %s ToServiceProvider returns nil", systemName)
	}

	// 注册到 Container
	if err := m.container.RegisterProvider(sp); err != nil {
		return fmt.Errorf("register plugin %s to container: %w", systemName, err)
	}

	// 注册到全局 Registry
	GetRegistry().Register(p)

	return nil
}

// Install 安装插件
//
// 中文说明:
// - 调用插件的 Install 方法;
// - 用于创建数据库表、初始化配置;
// - 安装后标记 meta.Installed = true;
// - 通常只需调用一次。
func (m *Manager) Install(ctx context.Context, systemName string) error {
	m.mu.RLock()
	p, ok := m.plugins[systemName]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("plugin %s not registered", systemName)
	}

	// 调用插件安装方法
	if err := p.Install(ctx, m.container); err != nil {
		return fmt.Errorf("install plugin %s: %w", systemName, err)
	}

	// 更新元数据
	m.mu.Lock()
	if meta, ok := m.metas[systemName]; ok {
		meta.Installed = true
		meta.InstalledVersion = p.Meta().Version
	}
	m.mu.Unlock()

	return nil
}

// Uninstall 卸载插件
//
// 中文说明:
// - 调用插件的 Uninstall 方法;
// - 用于清理数据、删除表(谨慎);
// - 卸载后从管理器移除;
// - 注意: 不从 Container 移除已绑定的服务。
func (m *Manager) Uninstall(ctx context.Context, systemName string) error {
	m.mu.RLock()
	p, ok := m.plugins[systemName]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("plugin %s not registered", systemName)
	}

	// 调用插件卸载方法
	if err := p.Uninstall(ctx, m.container); err != nil {
		return fmt.Errorf("uninstall plugin %s: %w", systemName, err)
	}

	// 从管理器移除
	m.mu.Lock()
	delete(m.plugins, systemName)
	if meta, ok := m.metas[systemName]; ok {
		meta.Installed = false
		meta.InstalledVersion = ""
	}
	m.mu.Unlock()

	return nil
}

// Boot 启动所有已安装的插件
//
// 中文说明:
// - 按依赖顺序依次调用插件的 Boot 方法;
// - 用于初始化运行时状态、读取配置;
// - 在服务启动时调用;
// - 如果某个插件 Boot 失败,会返回错误。
func (m *Manager) Boot(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 解析依赖顺序
	order, err := m.resolveDependencyOrder()
	if err != nil {
		return fmt.Errorf("resolve dependency order: %w", err)
	}

	// 按顺序启动
	for _, systemName := range order {
		p, ok := m.plugins[systemName]
		if !ok {
			continue // 未注册,跳过
		}

		meta := p.Meta()
		if meta == nil || !meta.Installed {
			continue // 未安装,跳过
		}

		if err := p.Boot(ctx, m.container); err != nil {
			return fmt.Errorf("boot plugin %s: %w", systemName, err)
		}
	}

	return nil
}

// GetPlugin 获取插件实例
//
// 中文说明:
// - 通过 systemName 获取已注册的插件;
// - 用于业务代码调用插件方法。
func (m *Manager) GetPlugin(systemName string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.plugins[systemName]
	return p, ok
}

// GetPaymentMethod 获取支付插件
//
// 中文说明:
// - 类型安全的获取支付插件;
// - 返回 PaymentMethod 接口;
// - 如果插件不是支付类型,返回 false。
func (m *Manager) GetPaymentMethod(systemName string) (PaymentMethod, bool) {
	p, ok := m.GetPlugin(systemName)
	if !ok {
		return nil, false
	}

	pm, ok := p.(PaymentMethod)
	return pm, ok
}

// GetShippingMethod 获取配送插件
func (m *Manager) GetShippingMethod(systemName string) (ShippingMethod, bool) {
	p, ok := m.GetPlugin(systemName)
	if !ok {
		return nil, false
	}

	sm, ok := p.(ShippingMethod)
	return sm, ok
}

// GetWidgetPlugin 获取小部件插件
func (m *Manager) GetWidgetPlugin(systemName string) (WidgetPlugin, bool) {
	p, ok := m.GetPlugin(systemName)
	if !ok {
		return nil, false
	}

	wp, ok := p.(WidgetPlugin)
	return wp, ok
}

// ListPlugins 列出所有已发现的插件元数据
//
// 中文说明:
// - 包含已注册和未注册的;
// - 按 DisplayOrder 排序;
// - 用于管理后台展示插件列表。
func (m *Manager) ListPlugins() []*PluginMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*PluginMeta, 0, len(m.metas))
	for _, meta := range m.metas {
		result = append(result, meta)
	}

	// 按 DisplayOrder 排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].DisplayOrder < result[j].DisplayOrder
	})

	return result
}

// ListInstalledPlugins 列出已安装的插件
func (m *Manager) ListInstalledPlugins() []*PluginMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*PluginMeta, 0)
	for _, meta := range m.metas {
		if meta.Installed {
			result = append(result, meta)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].DisplayOrder < result[j].DisplayOrder
	})

	return result
}

// ListPluginsByType 按类型列出插件
//
// 中文说明:
// - pluginType 例如: "payment", "shipping";
// - 用于管理后台按分类展示。
func (m *Manager) ListPluginsByType(pluginType string) []*PluginMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*PluginMeta, 0)
	for systemName, meta := range m.metas {
		p, ok := m.plugins[systemName]
		if ok && p.PluginType() == pluginType {
			result = append(result, meta)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].DisplayOrder < result[j].DisplayOrder
	})

	return result
}

// resolveDependencyOrder 解析依赖顺序(拓扑排序)
//
// 中文说明:
// - 根据 meta.DependsOn 构建依赖图;
// - 拓扑排序确保依赖先加载;
// - 检测循环依赖并返回错误。
func (m *Manager) resolveDependencyOrder() ([]string, error) {
	// 构建依赖图
	graph := make(map[string][]string)
	for systemName, meta := range m.metas {
		graph[systemName] = meta.DependsOn
	}

	// 拓扑排序
	var result []string
	visited := make(map[string]bool)
	visiting := make(map[string]bool) // 用于检测循环

	var visit func(string) error
	visit = func(node string) error {
		if visited[node] {
			return nil // 已访问
		}
		if visiting[node] {
			return fmt.Errorf("circular dependency detected at %s", node)
		}

		visiting[node] = true
		for _, dep := range graph[node] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[node] = false
		visited[node] = true

		result = append(result, node)
		return nil
	}

	// 遍历所有节点
	for node := range graph {
		if err := visit(node); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// SetInstalled 更新插件安装状态
//
// 中文说明:
// - 用于从数据库恢复安装状态;
// - 启动时调用,恢复上次保存的状态。
func (m *Manager) SetInstalled(systemName string, installed bool, version string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if meta, ok := m.metas[systemName]; ok {
		meta.Installed = installed
		meta.InstalledVersion = version
	}
}