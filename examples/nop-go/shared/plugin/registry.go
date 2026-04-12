// Package plugin 插件注册表
package plugin

import (
	"sort"
	"sync"
)

// Registry 插件注册表(全局单例)
//
// 中文说明:
// - 用于按类型查找已注册的插件;
// - 例如: 获取所有支付插件、所有配送插件;
// - 与 Manager 不同,Registry 专注于查找;
// - Manager.Register() 会自动注册到 Registry。
type Registry struct {
	mu           sync.RWMutex
	byType       map[string][]Plugin // pluginType -> []Plugin
	bySystemName map[string]Plugin   // systemName -> Plugin
}

// globalRegistry 全局注册表实例
var globalRegistry = &Registry{
	byType:       make(map[string][]Plugin),
	bySystemName: make(map[string]Plugin),
}

// GetRegistry 获取全局注册表
//
// 中文说明:
// - 返回全局单例;
// - 业务代码通过此方法获取 Registry;
// - 例如: GetRegistry().GetBySystemName("Payment.Alipay")。
func GetRegistry() *Registry {
	return globalRegistry
}

// Register 注册插件到注册表
//
// 中文说明:
// - 将插件添加到 byType 和 bySystemName 映射;
// - 同一个 systemName 重复注册会覆盖;
// - 由 Manager.Register() 自动调用。
func (r *Registry) Register(p Plugin) {
	r.mu.Lock()
	defer r.mu.Unlock()

	systemName := p.Meta().SystemName
	pluginType := p.PluginType()

	// 添加到 systemName 映射
	r.bySystemName[systemName] = p

	// 添加到 type 映射
	// 先检查是否已存在,避免重复
	existing := r.byType[pluginType]
	for i, ep := range existing {
		if ep.Meta().SystemName == systemName {
			// 已存在,替换
			existing[i] = p
			r.byType[pluginType] = existing
			return
		}
	}
	// 不存在,添加
	r.byType[pluginType] = append(existing, p)
}

// Unregister 从注册表移除插件
func (r *Registry) Unregister(systemName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.bySystemName[systemName]
	if !ok {
		return
	}

	// 从 systemName 映射移除
	delete(r.bySystemName, systemName)

	// 从 type 映射移除
	pluginType := p.PluginType()
	existing := r.byType[pluginType]
	for i, ep := range existing {
		if ep.Meta().SystemName == systemName {
			r.byType[pluginType] = append(existing[:i], existing[i+1:]...)
			break
		}
	}
}

// GetByType 按类型获取插件列表
//
// 中文说明:
// - pluginType 例如: "payment", "shipping", "widget";
// - 返回该类型的所有已注册插件;
// - 按 DisplayOrder 排序。
func (r *Registry) GetByType(pluginType string) []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := r.byType[pluginType]

	// 按 DisplayOrder 排序
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Meta().DisplayOrder < plugins[j].Meta().DisplayOrder
	})

	return plugins
}

// GetBySystemName 按系统名获取插件
//
// 中文说明:
// - systemName 例如: "Payment.Alipay";
// - 返回单个插件实例;
// - 如果不存在返回 nil, false。
func (r *Registry) GetBySystemName(systemName string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.bySystemName[systemName]
	return p, ok
}

// GetPaymentMethod 获取支付插件(类型安全)
//
// 中文说明:
// - 返回 PaymentMethod 接口;
// - 如果不是支付类型,返回 nil, false。
func (r *Registry) GetPaymentMethod(systemName string) (PaymentMethod, bool) {
	p, ok := r.GetBySystemName(systemName)
	if !ok {
		return nil, false
	}

	pm, ok := p.(PaymentMethod)
	return pm, ok
}

// GetShippingMethod 获取配送插件(类型安全)
func (r *Registry) GetShippingMethod(systemName string) (ShippingMethod, bool) {
	p, ok := r.GetBySystemName(systemName)
	if !ok {
		return nil, false
	}

	sm, ok := p.(ShippingMethod)
	return sm, ok
}

// GetWidgetPlugin 获取小部件插件(类型安全)
func (r *Registry) GetWidgetPlugin(systemName string) (WidgetPlugin, bool) {
	p, ok := r.GetBySystemName(systemName)
	if !ok {
		return nil, false
	}

	wp, ok := p.(WidgetPlugin)
	return wp, ok
}

// ListPaymentMethods 列出所有支付插件
//
// 中文说明:
// - 获取所有 PaymentMethod 类型的插件;
// - 用于支付方式选择列表;
// - 按 DisplayOrder 排序。
func (r *Registry) ListPaymentMethods() []PaymentMethod {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := r.byType["payment"]
	result := make([]PaymentMethod, 0, len(plugins))

	for _, p := range plugins {
		if pm, ok := p.(PaymentMethod); ok {
			result = append(result, pm)
		}
	}

	// 按 DisplayOrder 排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Meta().DisplayOrder < result[j].Meta().DisplayOrder
	})

	return result
}

// ListShippingMethods 列出所有配送插件
func (r *Registry) ListShippingMethods() []ShippingMethod {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := r.byType["shipping"]
	result := make([]ShippingMethod, 0, len(plugins))

	for _, p := range plugins {
		if sm, ok := p.(ShippingMethod); ok {
			result = append(result, sm)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Meta().DisplayOrder < result[j].Meta().DisplayOrder
	})

	return result
}

// ListWidgetPlugins 列出所有小部件插件
func (r *Registry) ListWidgetPlugins() []WidgetPlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := r.byType["widget"]
	result := make([]WidgetPlugin, 0, len(plugins))

	for _, p := range plugins {
		if wp, ok := p.(WidgetPlugin); ok {
			result = append(result, wp)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Meta().DisplayOrder < result[j].Meta().DisplayOrder
	})

	return result
}

// Count 统计已注册插件数量
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.bySystemName)
}

// CountByType 按类型统计插件数量
func (r *Registry) CountByType(pluginType string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byType[pluginType])
}

// Clear 清空注册表
//
// 中文说明:
// - 用于测试;
// - 生产环境不应调用。
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byType = make(map[string][]Plugin)
	r.bySystemName = make(map[string]Plugin)
}