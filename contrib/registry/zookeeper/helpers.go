// Package zookeeper provides Zookeeper registry helper functions.
// This file contains utility functions for sorting, snapshot keys, encoding, and in-memory backend.
//
// 本包提供 Zookeeper 注册中心辅助函数。
// 本文件包含排序、快照 key、编码和内存后端的工具函数。
package zookeeper

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// serviceRecord 定义服务实例在 Zookeeper 中存储的数据结构。
//
// serviceRecord 定义服务实例在 Zookeeper 中存储的 JSON 数据结构。
type serviceRecord struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Metadata map[string]string `json:"metadata"`
	Healthy  bool              `json:"healthy"`
}

// encodeServiceRecord 将服务实例记录编码为 JSON 字节数组。
//
// encodeServiceRecord 将 serviceRecord 编码为 []byte。
func encodeServiceRecord(record serviceRecord) ([]byte, error) {
	return json.Marshal(record)
}

// decodeServiceRecord 从 JSON 字节数组解码服务实例记录。
//
// decodeServiceRecord 从 []byte 解码为 serviceRecord。
func decodeServiceRecord(data []byte) (serviceRecord, error) {
	var record serviceRecord
	err := json.Unmarshal(data, &record)
	return record, err
}

// inMemoryZKBackend implements zkBackend for testing purposes.
//
// inMemoryZKBackend 实现 zkBackend 用于测试。
type inMemoryZKBackend struct {
	mu       sync.RWMutex
	nodes    map[string][]byte
	ephemeral map[string]struct{}
	children map[string][]string
	watchers map[string][]func()
}

// newInMemoryZKBackend creates a new in-memory backend for testing.
//
// newInMemoryZKBackend 创建新的内存后端用于测试。
func newInMemoryZKBackend() *inMemoryZKBackend {
	return &inMemoryZKBackend{
		nodes:     make(map[string][]byte),
		ephemeral: make(map[string]struct{}),
		children:  make(map[string][]string),
		watchers:  make(map[string][]func()),
	}
}

// EnsurePath ensures the target path exists in memory backend.
//
// EnsurePath 确保内存后端中目标路径存在。
func (b *inMemoryZKBackend) EnsurePath(target string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 分割路径并逐层创建
	parts := strings.Split(strings.Trim(target, "/"), "/")
	current := ""
	for _, part := range parts {
		current += "/" + part
		if _, exists := b.nodes[current]; !exists {
			b.nodes[current] = nil
		}
	}
	return nil
}

// CreateEphemeral creates an ephemeral node in memory backend.
//
// CreateEphemeral 在内存后端创建临时节点。
func (b *inMemoryZKBackend) CreateEphemeral(target string, data []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 检查节点是否已存在
	if _, exists := b.nodes[target]; exists {
		return zk.ErrNodeExists
	}

	// 创建临时节点
	b.nodes[target] = data
	b.ephemeral[target] = struct{}{}

	// 更新父节点的子节点列表
	parent := path.Dir(target)
	if parent != "/" && parent != "" {
		childName := path.Base(target)
		b.children[parent] = append(b.children[parent], childName)
	}

	// 通知监听器
	b.notifyWatchersLocked(parent)
	return nil
}

// Delete deletes a node from memory backend.
//
// Delete 从内存后端删除节点。
func (b *inMemoryZKBackend) Delete(target string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 检查节点是否存在
	if _, exists := b.nodes[target]; !exists {
		return zk.ErrNoNode
	}

	// 删除节点
	delete(b.nodes, target)
	delete(b.ephemeral, target)

	// 更新父节点的子节点列表
	parent := path.Dir(target)
	if parent != "/" && parent != "" {
		childName := path.Base(target)
		children := b.children[parent]
		for i, child := range children {
			if child == childName {
				b.children[parent] = append(children[:i], children[i+1:]...)
				break
			}
		}
	}

	// 通知监听器
	b.notifyWatchersLocked(parent)
	return nil
}

// Children returns the list of child nodes under the target path.
//
// Children 返回目标路径下的子节点列表。
func (b *inMemoryZKBackend) Children(target string) ([]string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// 检查节点是否存在
	if _, exists := b.nodes[target]; !exists {
		return nil, zk.ErrNoNode
	}

	children := b.children[target]
	if children == nil {
		return []string{}, nil
	}
	return append([]string(nil), children...), nil
}

// Get returns the data stored in the node at the target path.
//
// Get 返回目标路径节点中存储的数据。
func (b *inMemoryZKBackend) Get(target string) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// 检查节点是否存在
	if _, exists := b.nodes[target]; !exists {
		return nil, zk.ErrNoNode
	}

	data := b.nodes[target]
	if data == nil {
		return []byte{}, nil
	}
	return append([]byte(nil), data...), nil
}

// WatchChildren watches for changes to child nodes under the target path.
//
// WatchChildren 监听目标路径下子节点的变更。
func (b *inMemoryZKBackend) WatchChildren(ctx context.Context, target string, onUpdate func()) error {
	b.mu.Lock()
	// 检查节点是否存在
	if _, exists := b.nodes[target]; !exists {
		b.mu.Unlock()
		onUpdate()
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(200 * time.Millisecond):
			return b.WatchChildren(ctx, target, onUpdate)
		}
	}

	// 注册监听器
	b.watchers[target] = append(b.watchers[target], onUpdate)
	b.mu.Unlock()

	// 等待上下文取消
	<-ctx.Done()

	// 移除监听器
	b.mu.Lock()
	watchers := b.watchers[target]
	for i := range watchers {
		if &watchers[i] == &onUpdate {
			b.watchers[target] = append(watchers[:i], watchers[i+1:]...)
			break
		}
	}
	if len(b.watchers[target]) == 0 {
		delete(b.watchers, target)
	}
	b.mu.Unlock()

	return nil
}

// Close closes the memory backend.
//
// Close 关闭内存后端。
func (b *inMemoryZKBackend) Close() error {
	return nil
}

// Underlying returns nil for memory backend (no native client).
//
// Underlying 内存后端返回 nil（无原生客户端）。
func (b *inMemoryZKBackend) Underlying() any {
	return nil
}

// notifyWatchersLocked notifies all watchers for a path change (must hold lock).
//
// notifyWatchersLocked 通知路径的所有监听器变更（必须持有锁）。
func (b *inMemoryZKBackend) notifyWatchersLocked(path string) {
	watchers := append([]func(){}, b.watchers[path]...)
	for _, watcher := range watchers {
		watcher()
	}
}

// sanitizeNodeName sanitizes address for use as ZNode name.
//
// sanitizeNodeName 规范化地址用于 ZNode 名称。
func sanitizeNodeName(addr string) string {
	return strings.ReplaceAll(addr, "/", "_")
}

// generateInstanceID generates a service instance ID.
//
// generateInstanceID 生成服务实例 ID。
func generateInstanceID(name, addr string) string {
	return name + "-" + addr
}

// instanceKey generates a registry-internal key for tracking registered instances.
//
// instanceKey 生成注册中心内部追踪已注册实例的 key。
func instanceKey(name, addr string) string {
	return name + "|" + addr
}

// mergeMeta merges base and extra metadata.
//
// mergeMeta 合并基础和额外的 metadata。
func mergeMeta(base map[string]string, extra map[string]string) map[string]string {
	result := make(map[string]string, len(base)+len(extra))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range extra {
		result[k] = v
	}
	return result
}

// sortServiceInstances sorts service instances by ID and address.
//
// sortServiceInstances 按 ID 和地址排序服务实例。
func sortServiceInstances(instances []transportcontract.ServiceInstance) {
	sort.Slice(instances, func(i, j int) bool {
		if instances[i].ID != instances[j].ID {
			return instances[i].ID < instances[j].ID
		}
		return instances[i].Address < instances[j].Address
	})
}

// snapshotKey generates a snapshot key for instances.
//
// snapshotKey 为实例生成快照 key。
func snapshotKey(instances []transportcontract.ServiceInstance) string {
	if len(instances) == 0 {
		return "<empty>"
	}
	parts := make([]string, 0, len(instances))
	for _, instance := range instances {
		parts = append(parts, instance.ID+"|"+instance.Address)
	}
	sort.Strings(parts)
	return strings.Join(parts, ";")
}

// isRetryableWatchError checks if the error is retryable for watch.
//
// isRetryableWatchError 检查错误是否可重试监听。
func isRetryableWatchError(err error) bool {
	return errors.Is(err, zk.ErrConnectionClosed) || errors.Is(err, zk.ErrSessionExpired)
}