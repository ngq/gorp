// Package eureka provides Eureka registry helper functions.
// This file contains utility functions, error definitions, and in-memory client for testing.
//
// 本包提供 Eureka 注册中心辅助函数。
// 本文件包含工具函数、错误定义和内存客户端（用于测试）。
package eureka

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// ============================================================================
// 错误定义
// ============================================================================

// ErrNoServerURL indicates Eureka server URL is required.
//
// ErrNoServerURL 表示 Eureka 服务端地址必需。
var ErrNoServerURL = errors.New("eureka: server_url is required")

// ErrServiceNotFound indicates Eureka service not found.
//
// ErrServiceNotFound 表示 Eureka 服务未找到。
var ErrServiceNotFound = errors.New("eureka: service not found")

// ErrRegistryClosed indicates Eureka registry closed.
//
// ErrRegistryClosed 表示 Eureka 注册中心已关闭。
var ErrRegistryClosed = errors.New("eureka: registry closed")

// ErrAlreadyRegistered indicates Eureka instance already registered.
//
// ErrAlreadyRegistered 表示 Eureka 实例已注册。
var ErrAlreadyRegistered = errors.New("eureka: instance already registered")

// ============================================================================
// 辅助函数
// ============================================================================

// instanceKey generates a unique key for an instance (name|addr).
//
// instanceKey 为实例生成唯一 key（name|addr）。
// 用于内部缓存和去重判断。
func instanceKey(name, addr string) string {
	return name + "|" + addr
}

// instanceID generates Eureka instance ID (NAME:addr).
//
// instanceID 生成 Eureka 实例 ID（NAME:addr）。
// Eureka 实例 ID 格式为大写应用名:地址。
func instanceID(name, addr string) string {
	return strings.ToUpper(name) + ":" + addr
}

// hostFromAddr extracts host from address (host:port).
//
// hostFromAddr 从地址中提取 host（host:port）。
func hostFromAddr(addr string) string {
	parts := strings.Split(addr, ":")
	if len(parts) == 0 {
		return addr
	}
	return parts[0]
}

// portFromAddr extracts port from address, with fallback.
//
// portFromAddr 从地址中提取 port，提供 fallback 值。
func portFromAddr(addr string, fallback int) int {
	parts := strings.Split(addr, ":")
	if len(parts) < 2 {
		return fallback
	}
	var port int
	_, _ = fmt.Sscanf(parts[len(parts)-1], "%d", &port)
	if port == 0 {
		return fallback
	}
	return port
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

// cloneStringMap creates a copy of a string map.
//
// cloneStringMap 创建 string map 的副本。
func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	out := make(map[string]string, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

// snapshotKey generates a snapshot key for instances (for deduplication).
//
// snapshotKey 为实例生成快照 key（用于去重）。
// 格式：ID|Address|Healthy，按 ID 排序拼接。
func snapshotKey(instances []transportcontract.ServiceInstance) string {
	if len(instances) == 0 {
		return "<empty>"
	}
	parts := make([]string, 0, len(instances))
	for _, instance := range instances {
		parts = append(parts, instance.ID+"|"+instance.Address+"|"+fmt.Sprintf("%t", instance.Healthy))
	}
	sort.Strings(parts)
	return strings.Join(parts, ";")
}

// watchRetryInterval returns the retry interval for watch errors.
//
// watchRetryInterval 返回 watch 错误的重试间隔。
// 优先使用 WatchInterval，否则使用 HeartbeatRetryBackoff，最后默认 1 秒。
func watchRetryInterval(cfg *EurekaConfig) time.Duration {
	if cfg.WatchInterval > 0 {
		return cfg.WatchInterval
	}
	if cfg.HeartbeatRetryBackoff > 0 {
		return cfg.HeartbeatRetryBackoff
	}
	return time.Second
}

// isRetryableWatchError checks if a watch error should be retried.
//
// isRetryableWatchError 检查 watch 错误是否应该重试。
// ErrRegistryClosed 和 context.Canceled 不重试。
func isRetryableWatchError(err error) bool {
	if err == nil {
		return false
	}
	return !errors.Is(err, ErrRegistryClosed) && !errors.Is(err, context.Canceled)
}

// ============================================================================
// 内存客户端（用于测试）
// ============================================================================

// inMemoryEurekaClient implements eurekaClient for testing.
//
// inMemoryEurekaClient 实现 eurekaClient 用于测试。
// 不依赖真实 Eureka 服务端，所有数据存储在内存中。
type inMemoryEurekaClient struct {
	mu       sync.RWMutex
	services map[string][]transportcontract.ServiceInstance
	watchers map[string][]chan []transportcontract.ServiceInstance
}

// newInMemoryEurekaClient creates a new in-memory Eureka client.
//
// newInMemoryEurekaClient 创建新的内存 Eureka 客户端。
func newInMemoryEurekaClient() eurekaClient {
	return &inMemoryEurekaClient{
		services: make(map[string][]transportcontract.ServiceInstance),
		watchers: make(map[string][]chan []transportcontract.ServiceInstance),
	}
}

// Register registers a service instance in memory.
//
// Register 在内存中注册服务实例。
func (c *inMemoryEurekaClient) Register(ctx context.Context, cfg *EurekaConfig, name, addr string, meta map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullMeta := make(map[string]string)
	for k, v := range cfg.ServiceMeta {
		fullMeta[k] = v
	}
	for k, v := range meta {
		fullMeta[k] = v
	}

	instance := transportcontract.ServiceInstance{
		ID:       instanceID(name, addr),
		Name:     name,
		Address:  addr,
		Metadata: fullMeta,
		Healthy:  true,
	}

	c.services[name] = append(c.services[name], instance)
	c.notifyWatchersLocked(name)
	return nil
}

// Deregister deregisters a service instance from memory.
//
// Deregister 从内存注销服务实例。
func (c *inMemoryEurekaClient) Deregister(ctx context.Context, cfg *EurekaConfig, name, addr string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	instances := c.services[name]
	for i, inst := range instances {
		if inst.Address == addr {
			c.services[name] = append(instances[:i], instances[i+1:]...)
			c.notifyWatchersLocked(name)
			return nil
		}
	}
	return ErrServiceNotFound
}

// Heartbeat does nothing for in-memory client (always succeeds).
//
// Heartbeat 内存客户端心跳无实际操作（总是成功）。
func (c *inMemoryEurekaClient) Heartbeat(ctx context.Context, cfg *EurekaConfig, name, addr string) error {
	return nil
}

// Discover discovers service instances from memory.
//
// Discover 从内存发现服务实例。
func (c *inMemoryEurekaClient) Discover(ctx context.Context, cfg *EurekaConfig, name string) ([]transportcontract.ServiceInstance, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	instances := c.services[name]
	if len(instances) == 0 {
		return nil, ErrServiceNotFound
	}
	result := make([]transportcontract.ServiceInstance, len(instances))
	copy(result, instances)
	return result, nil
}

// Watch watches service instance changes from memory.
//
// Watch 从内存监听服务实例变更。
func (c *inMemoryEurekaClient) Watch(ctx context.Context, cfg *EurekaConfig, name string, onUpdate func([]transportcontract.ServiceInstance)) error {
	ch := make(chan []transportcontract.ServiceInstance, 4)

	c.mu.Lock()
	c.watchers[name] = append(c.watchers[name], ch)
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		watchers := c.watchers[name]
		for i, watcher := range watchers {
			if watcher == ch {
				c.watchers[name] = append(watchers[:i], watchers[i+1:]...)
				break
			}
		}
		if len(c.watchers[name]) == 0 {
			delete(c.watchers, name)
		}
		c.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case instances := <-ch:
			onUpdate(instances)
		}
	}
}

// notifyWatchersLocked notifies watchers for a service change.
// Must be called with c.mu held.
//
// notifyWatchersLocked 通知监听器服务变更。
// 必须在持有 c.mu 时调用。
func (c *inMemoryEurekaClient) notifyWatchersLocked(name string) {
	watchers := append([]chan []transportcontract.ServiceInstance(nil), c.watchers[name]...)
	instances := append([]transportcontract.ServiceInstance(nil), c.services[name]...)
	for _, watcher := range watchers {
		select {
		case watcher <- instances:
		default:
		}
	}
}