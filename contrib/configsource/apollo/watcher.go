// Package apollo provides Apollo configuration watcher implementation.
// This file implements the ConfigWatcher contract for change notification.
//
// 本包提供 Apollo 配置监听器实现。
// 本文件实现 ConfigWatcher 契约，用于变更通知。
package apollo

import (
	"context"
	"sync"
)

// apolloWatcher implements datacontract.ConfigWatcher.
// Dispatches config change callbacks to registered handlers.
//
// apolloWatcher 实现 datacontract.ConfigWatcher。
// 将配置变更回调分发给注册的处理器。
type apolloWatcher struct {
	cancel    context.CancelFunc
	source    *ConfigSource
	callbacks *sync.Map
	stopped   bool
	stopMu    sync.RWMutex
}

// OnChange registers a callback for key changes.
// Implements datacontract.ConfigWatcher.OnChange.
//
// OnChange 为 key 变更注册回调。
// 实现 datacontract.ConfigWatcher.OnChange。
func (w *apolloWatcher) OnChange(key string, callback func(value any)) {
	w.stopMu.RLock()
	stopped := w.stopped
	w.stopMu.RUnlock()
	if stopped {
		return
	}
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
//
// Stop 停止监听并释放相关资源。
// 实现 datacontract.ConfigWatcher.Stop。
func (w *apolloWatcher) Stop() error {
	w.stopMu.Lock()
	if w.stopped {
		w.stopMu.Unlock()
		return nil
	}
	w.stopped = true
	w.stopMu.Unlock()
	w.cancel()
	w.source.watchers.Delete(w)
	return nil
}

// dispatch sends current values to all registered callbacks.
//
// dispatch 将当前值发送给所有注册的回调。
func (w *apolloWatcher) dispatch() {
	w.stopMu.RLock()
	stopped := w.stopped
	w.stopMu.RUnlock()
	if stopped {
		return
	}
	w.callbacks.Range(func(key, value any) bool {
		w.stopMu.RLock()
		stopped := w.stopped
		w.stopMu.RUnlock()
		if stopped {
			return false
		}
		path, ok := key.(string)
		if !ok {
			return true
		}
		callback, ok := value.(func(value any))
		if !ok {
			return true
		}

		w.source.mu.RLock()
		current, exists := lookupNestedValue(w.source.cache, path)
		w.source.mu.RUnlock()
		if exists {
			callback(current)
		}
		return true
	})
}