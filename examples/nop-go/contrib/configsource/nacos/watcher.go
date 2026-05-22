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
	ctx       context.Context
	cancel    context.CancelFunc
	source    *ConfigSource
	callbacks *sync.Map
	stopped   bool
	stopMu    sync.Mutex
}

// OnChange registers a callback for key changes.
// Implements datacontract.ConfigWatcher.OnChange.
// No-ops when the watcher has already been stopped.
//
// OnChange 为 key 变更注册回调。
// 实现 datacontract.ConfigWatcher.OnChange。
// 监听器已停止时不再注册新回调。
func (w *nacosWatcher) OnChange(key string, callback func(value any)) {
	w.stopMu.Lock()
	if w.stopped {
		w.stopMu.Unlock()
		return
	}
	w.stopMu.Unlock()

	w.callbacks.Store(key, callback)

	w.source.mu.RLock()
	current, exists := lookupNestedValue(w.source.cache, key)
	w.source.mu.RUnlock()
	if exists {
		callback(current)
	}
}

// Stop stops the watcher and releases related resources.
// Implements datacontract.ConfigWatcher.Stop.
// Safe to call multiple times; subsequent calls are no-ops.
//
// Stop 停止监听并释放相关资源。
// 实现 datacontract.ConfigWatcher.Stop。
// 可安全重复调用，后续调用为空操作。
func (w *nacosWatcher) Stop() error {
	w.stopMu.Lock()
	defer w.stopMu.Unlock()
	if w.stopped {
		return nil
	}
	w.stopped = true
	w.cancel()
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
