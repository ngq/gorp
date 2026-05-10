// Package nacos provides Nacos configuration watcher implementation.
// This file implements the ConfigWatcher contract for change notification.
//
// 本包提供 Nacos 配置监听器实现。
// 本文件实现 ConfigWatcher 契约，用于变更通知。
package nacos

import (
	"context"
	"sync"
)

// nacosWatcher implements datacontract.ConfigWatcher.
// Dispatches config change callbacks to registered handlers.
//
// nacosWatcher 实现 datacontract.ConfigWatcher。
// 将配置变更回调分发给注册的处理器。
// 该监听器通过底层 Nacos SDK 的 ListenConfig 机制接收配置变更，
// 并将变更分发给所有通过 OnChange 注册的回调函数。
type nacosWatcher struct {
	// ctx 是监听器的上下文，用于控制监听生命周期。
	ctx context.Context
	// cancel 用于取消监听上下文，停止监听。
	cancel context.CancelFunc
	// source 是关联的配置源，用于读取当前配置值。
	source *ConfigSource
	// callbacks 存储所有注册的回调函数，key 为配置路径。
	callbacks *sync.Map
}

// OnChange registers a callback for key changes.
// Implements datacontract.ConfigWatcher.OnChange.
//
// OnChange 为 key 变更注册回调。
// 实现 datacontract.ConfigWatcher.OnChange。
// 该方法注册一个回调函数，当指定的配置路径值发生变化时被调用。
// 注册时会立即读取当前值并触发回调，确保回调初始状态。
// 参数：
//   - key: 配置路径，支持点分隔的嵌套路径，如 "app.name"
//   - callback: 变更回调函数，接收新值作为参数
func (w *nacosWatcher) OnChange(key string, callback func(value any)) {
	// 注册回调到 callbacks map
	w.callbacks.Store(key, callback)

	// 立即读取当前值并触发回调
	// 这确保回调函数在注册时就能获得当前状态，而不是等待下次变更
	w.source.mu.RLock()
	current, exists := lookupNestedValue(w.source.cache, key)
	w.source.mu.RUnlock()
	if exists {
		callback(current)
	}
}

// Stop stops the watcher and releases related resources.
// Implements datacontract.ConfigWatcher.Stop.
//
// Stop 停止监听并释放相关资源。
// 实现 datacontract.ConfigWatcher.Stop。
// 该方法取消监听上下文，停止后台监听 goroutine，并从配置源中移除监听器记录。
func (w *nacosWatcher) Stop() error {
	// 取消监听上下文，停止后台 goroutine
	w.cancel()
	// 从配置源的监听器列表中移除
	w.source.watchers.Delete(w)
	return nil
}

// dispatch sends current values to all registered callbacks.
//
// dispatch 将当前值发送给所有注册的回调。
// 该方法在配置变更时被调用，遍历所有注册的回调并触发通知。
// dispatch 是内部方法，由配置源的 Watch 回调触发。
func (w *nacosWatcher) dispatch() {
	// 遍历所有注册的回调
	w.callbacks.Range(func(key, value any) bool {
		// 检查 key 是否为字符串类型
		path, ok := key.(string)
		if !ok {
			return true // 继续遍历
		}
		// 检查回调函数类型
		callback, ok := value.(func(value any))
		if !ok {
			return true // 继续遍历
		}

		// 读取当前配置值
		w.source.mu.RLock()
		current, exists := lookupNestedValue(w.source.cache, path)
		w.source.mu.RUnlock()

		// 如果值存在，触发回调
		if exists {
			callback(current)
		}
		return true // 继续遍历所有回调
	})
}