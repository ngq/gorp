// Package polaris provides Polaris registry helper functions.
// This file contains utility functions for sorting, snapshot keys, and in-memory client.
//
// 本包提供 Polaris 注册中心辅助函数。
// 本文件包含排序、快照 key 和内存客户端的工具函数。
package polaris

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// inMemoryPolarisClient implements polarisRegistryClient for testing.
//
// inMemoryPolarisClient 实现 polarisRegistryClient 用于测试。
type inMemoryPolarisClient struct {
	mu       sync.RWMutex
	services map[string][]transportcontract.ServiceInstance
	watchers map[string][]chan []transportcontract.ServiceInstance
}

// Register registers a service instance in memory.
//
// Register 在内存中注册服务实例。
func (c *inMemoryPolarisClient) Register(ctx context.Context, cfg *PolarisConfig, name, addr string, meta map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.services == nil {
		c.services = make(map[string][]transportcontract.ServiceInstance)
	}
	fullMeta := make(map[string]string)
	for k, v := range cfg.ServiceMeta {
		fullMeta[k] = v
	}
	for k, v := range meta {
		fullMeta[k] = v
	}
	instance := transportcontract.ServiceInstance{
		ID:       generateServiceID(name, addr),
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
func (c *inMemoryPolarisClient) Deregister(ctx context.Context, cfg *PolarisConfig, name, addr string) error {
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

// Discover discovers service instances from memory.
//
// Discover 从内存发现服务实例。
func (c *inMemoryPolarisClient) Discover(ctx context.Context, cfg *PolarisConfig, name string) ([]transportcontract.ServiceInstance, error) {
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
func (c *inMemoryPolarisClient) Watch(ctx context.Context, cfg *PolarisConfig, name string, onUpdate func([]transportcontract.ServiceInstance)) error {
	ch := make(chan []transportcontract.ServiceInstance, 4)

	c.mu.Lock()
	if c.watchers == nil {
		c.watchers = make(map[string][]chan []transportcontract.ServiceInstance)
	}
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
//
// notifyWatchersLocked 通知监听器服务变更。
func (c *inMemoryPolarisClient) notifyWatchersLocked(name string) {
	watchers := append([]chan []transportcontract.ServiceInstance(nil), c.watchers[name]...)
	instances := append([]transportcontract.ServiceInstance(nil), c.services[name]...)
	for _, watcher := range watchers {
		select {
		case watcher <- instances:
		default:
		}
	}
}

// generateServiceID generates a service instance ID.
//
// generateServiceID 生成服务实例 ID。
func generateServiceID(name, addr string) string {
	return fmt.Sprintf("%s-%s", name, addr)
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
		return ""
	}
	var builder strings.Builder
	for _, inst := range instances {
		builder.WriteString(inst.ID)
		builder.WriteString("|")
		builder.WriteString(inst.Address)
		builder.WriteString("|")
		if inst.Healthy {
			builder.WriteString("1")
		} else {
			builder.WriteString("0")
		}
		builder.WriteString(";")
	}
	return builder.String()
}