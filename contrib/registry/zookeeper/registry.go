// Package zookeeper provides Zookeeper service registry implementation.
// This file implements the ServiceRegistry contract with Zookeeper SDK integration.
//
// 本包提供 Zookeeper 服务注册实现。
// 本文件实现 ServiceRegistry 契约，集成 Zookeeper SDK。
package zookeeper

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sync"
	"time"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// ErrNoServers indicates Zookeeper servers are required.
//
// ErrNoServers 表示 Zookeeper 服务器地址必需。
var ErrNoServers = errors.New("registry.zookeeper: no servers configured")

// ErrServiceNotFound indicates Zookeeper service not found.
//
// ErrServiceNotFound 表示 Zookeeper 服务未找到。
var ErrServiceNotFound = errors.New("registry.zookeeper: service not found")

// ErrRegistryClosed indicates Zookeeper registry closed.
//
// ErrRegistryClosed 表示 Zookeeper 注册中心已关闭。
var ErrRegistryClosed = errors.New("registry.zookeeper: registry closed")

// ErrAlreadyRegistered indicates Zookeeper instance already registered.
//
// ErrAlreadyRegistered 表示 Zookeeper 实例已注册。
var ErrAlreadyRegistered = errors.New("registry.zookeeper: instance already registered")

// Registry implements transportcontract.ServiceRegistry with Zookeeper SDK.
// Supports service registration, discovery, and watch with caching.
//
// Registry 使用 Zookeeper SDK 实现 transportcontract.ServiceRegistry。
// 支持服务注册、发现和监听，带缓存功能。
type Registry struct {
	config  *ZookeeperConfig
	backend zkBackend

	mu                  sync.RWMutex
	endpointCache       map[string][]transportcontract.ServiceInstance
	watchSnapshots      map[string]string
	registeredInstances map[string]string
	closeMu             sync.Mutex
	closed              bool
	watchCancels        []context.CancelFunc
}

// NewRegistry creates a new Zookeeper registry with default backend.
//
// NewRegistry 使用默认后端创建新的 Zookeeper 注册中心。
func NewRegistry(cfg *ZookeeperConfig) (*Registry, error) {
	backend, err := newZKBackend(cfg)
	if err != nil {
		return nil, err
	}
	return NewRegistryWithBackend(cfg, backend)
}

// NewRegistryWithBackend creates a new Zookeeper registry with custom backend.
//
// NewRegistryWithBackend 使用自定义后端创建新的 Zookeeper 注册中心。
func NewRegistryWithBackend(cfg *ZookeeperConfig, backend zkBackend) (*Registry, error) {
	if len(cfg.Servers) == 0 {
		return nil, ErrNoServers
	}
	if backend == nil {
		return nil, errors.New("registry.zookeeper: backend is required")
	}
	return &Registry{
		config:              cfg,
		backend:             backend,
		endpointCache:       make(map[string][]transportcontract.ServiceInstance),
		watchSnapshots:      make(map[string]string),
		registeredInstances: make(map[string]string),
	}, nil
}

// Register registers a service instance to Zookeeper.
// Implements transportcontract.ServiceRegistry.Register.
//
// Register 将服务实例注册到 Zookeeper。
// 实现 transportcontract.ServiceRegistry.Register。
func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrRegistryClosed
	}

	key := instanceKey(name, addr)
	if _, exists := r.registeredInstances[key]; exists {
		return ErrAlreadyRegistered
	}

	// 确保 Zookeeper 服务路径存在
	servicePath := path.Join(r.config.BasePath, name)
	if err := r.backend.EnsurePath(servicePath); err != nil {
		return fmt.Errorf("registry.zookeeper: ensure service path failed: %w", err)
	}

	// 构造服务实例记录
	record := serviceRecord{
		ID:       generateInstanceID(name, addr),
		Name:     name,
		Address:  addr,
		Metadata: mergeMeta(r.config.ServiceMeta, meta),
		Healthy:  true,
	}
	payload, err := encodeServiceRecord(record)
	if err != nil {
		return fmt.Errorf("registry.zookeeper: encode instance data failed: %w", err)
	}

	// 创建临时节点（Ephemeral ZNode）
	instancePath := path.Join(servicePath, sanitizeNodeName(addr))
	if err := r.backend.CreateEphemeral(instancePath, payload); err != nil {
		return fmt.Errorf("registry.zookeeper: create ephemeral node failed: %w", err)
	}

	// 记录已注册实例路径，清理缓存
	r.registeredInstances[key] = instancePath
	delete(r.endpointCache, name)
	delete(r.watchSnapshots, name)
	return nil
}

// Deregister deregisters a service instance from Zookeeper.
// Implements transportcontract.ServiceRegistry.Deregister.
//
// Deregister 将服务实例从 Zookeeper 注销。
// 实现 transportcontract.ServiceRegistry.Deregister。
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrRegistryClosed
	}

	key := instanceKey(name, addr)
	instancePath, ok := r.registeredInstances[key]
	if !ok {
		// 未在本地记录中找到，尝试构造路径删除
		instancePath = path.Join(r.config.BasePath, name, sanitizeNodeName(addr))
	}

	if err := r.backend.Delete(instancePath); err != nil {
		if errors.Is(err, errZKNoNode) {
			return ErrServiceNotFound
		}
		return fmt.Errorf("registry.zookeeper: delete instance failed: %w", err)
	}

	// 清理本地记录和缓存
	delete(r.registeredInstances, key)
	delete(r.endpointCache, name)
	delete(r.watchSnapshots, name)
	return nil
}

// Discover discovers service instances from Zookeeper.
// Implements transportcontract.ServiceRegistry.Discover.
//
// Discover 从 Zookeeper 发现服务实例。
// 实现 transportcontract.ServiceRegistry.Discover。
func (r *Registry) Discover(ctx context.Context, name string) ([]transportcontract.ServiceInstance, error) {
	r.mu.RLock()
	// 检查缓存是否可用
	if instances, ok := r.endpointCache[name]; ok && len(instances) > 0 {
		cached := append([]transportcontract.ServiceInstance(nil), instances...)
		r.mu.RUnlock()
		return cached, nil
	}
	closed := r.closed
	r.mu.RUnlock()
	if closed {
		return nil, ErrRegistryClosed
	}

	// 从 Zookeeper 获取服务实例列表
	servicePath := path.Join(r.config.BasePath, name)
	children, err := r.backend.Children(servicePath)
	if err != nil {
		if errors.Is(err, errZKNoNode) {
			return nil, ErrServiceNotFound
		}
		return nil, fmt.Errorf("registry.zookeeper: list instances failed: %w", err)
	}
	if len(children) == 0 {
		return nil, ErrServiceNotFound
	}

	// 解析每个实例的数据
	result := make([]transportcontract.ServiceInstance, 0, len(children))
	for _, child := range children {
		data, getErr := r.backend.Get(path.Join(servicePath, child))
		if getErr != nil {
			return nil, fmt.Errorf("registry.zookeeper: read instance failed: %w", getErr)
		}

		record, err := decodeServiceRecord(data)
		if err != nil {
			return nil, fmt.Errorf("registry.zookeeper: decode instance failed: %w", err)
		}
		result = append(result, transportcontract.ServiceInstance{
			ID:       record.ID,
			Name:     record.Name,
			Address:  record.Address,
			Metadata: record.Metadata,
			Healthy:  record.Healthy,
		})
	}

	// 排序并缓存结果
	sortServiceInstances(result)
	r.mu.Lock()
	r.endpointCache[name] = append([]transportcontract.ServiceInstance(nil), result...)
	r.mu.Unlock()
	return result, nil
}

// Watch watches service instance changes from Zookeeper.
// Implements transportcontract.ServiceRegistry.Watch.
//
// Watch 监听 Zookeeper 服务实例变更。
// 实现 transportcontract.ServiceRegistry.Watch。
func (r *Registry) Watch(ctx context.Context, name string) (<-chan []transportcontract.ServiceInstance, error) {
	r.closeMu.Lock()
	defer r.closeMu.Unlock()

	if r.closed {
		return nil, ErrRegistryClosed
	}

	// 创建 watch 上下文和取消函数
	watchCtx, cancel := context.WithCancel(ctx)
	r.watchCancels = append(r.watchCancels, cancel)

	ch := make(chan []transportcontract.ServiceInstance, 10)
	servicePath := path.Join(r.config.BasePath, name)
	var workers sync.WaitGroup

	// emit 函数用于发送实例变更通知，带快照去重逻辑
	emit := func(instances []transportcontract.ServiceInstance) bool {
		snapshot := snapshotKey(instances)

		r.mu.Lock()
		last := r.watchSnapshots[name]
		if last == snapshot {
			r.mu.Unlock()
			return true
		}
		r.watchSnapshots[name] = snapshot
		r.mu.Unlock()

		// 发送实例列表到通道
		select {
		case ch <- append([]transportcontract.ServiceInstance(nil), instances...):
			return true
		case <-watchCtx.Done():
			return false
		default:
			return true
		}
	}

	// 启动 watch worker，监听 Zookeeper 子节点变更
	workers.Add(1)
	go func() {
		defer workers.Done()
		for {
			err := r.backend.WatchChildren(watchCtx, servicePath, func() {
				// 清理缓存并重新发现
				r.mu.Lock()
				delete(r.endpointCache, name)
				r.mu.Unlock()

				instances, err := r.Discover(watchCtx, name)
				if err != nil {
					if errors.Is(err, ErrServiceNotFound) {
						r.mu.Lock()
						delete(r.endpointCache, name)
						r.mu.Unlock()
						emit(nil)
					}
					return
				}
				emit(instances)
			})
			if err == nil || watchCtx.Err() != nil {
				return
			}
			if !isRetryableWatchError(err) {
				return
			}
			// 重试等待
			select {
			case <-watchCtx.Done():
				return
			case <-time.After(r.config.WatchRetryInterval):
			}
		}
	}()

	// 启动初始发现 worker
	workers.Add(1)
	go func() {
		defer workers.Done()
		instances, err := r.Discover(watchCtx, name)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				emit(nil)
			}
			return
		}
		emit(instances)
	}()

	// 等待所有 worker 完成后关闭通道
	go func() {
		workers.Wait()
		close(ch)
	}()

	return ch, nil
}

// Close releases resources held by the registry.
// Implements transportcontract.ServiceRegistry.Close.
//
// Close 释放注册中心持有的资源。
// 实现 transportcontract.ServiceRegistry.Close。
func (r *Registry) Close() error {
	r.closeMu.Lock()
	defer r.closeMu.Unlock()

	if r.closed {
		return nil
	}
	r.closed = true

	// 取消所有 watch 上下文
	for _, cancel := range r.watchCancels {
		cancel()
	}
	r.watchCancels = nil

	return r.backend.Close()
}

// Underlying returns the current native backend object used by this registry.
//
// Underlying 返回此注册中心使用的当前原生后端对象。
func (r *Registry) Underlying() any {
	if provider, ok := r.backend.(zkBackendNativeProvider); ok {
		if native := provider.Underlying(); native != nil {
			return native
		}
	}
	return r.backend
}

// As projects the current native backend into the requested target when possible.
//
// As 在可能时将当前原生后端投射到请求的目标。
func (r *Registry) As(target any) bool {
	return internalnative.As(r.Underlying(), target)
}

// zkBackend defines the internal backend interface for Zookeeper operations.
//
// zkBackend 定义 Zookeeper 操作的内部后端接口。
type zkBackend interface {
	EnsurePath(path string) error
	CreateEphemeral(path string, data []byte) error
	Delete(path string) error
	Children(path string) ([]string, error)
	Get(path string) ([]byte, error)
	WatchChildren(ctx context.Context, path string, onUpdate func()) error
	Close() error
}

// zkBackendNativeProvider defines the interface for accessing underlying backend.
//
// zkBackendNativeProvider 定义访问底层后端的接口。
type zkBackendNativeProvider interface {
	Underlying() any
}
