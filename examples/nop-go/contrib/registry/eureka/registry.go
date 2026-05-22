// Package eureka provides Eureka service registry implementation.
// This file implements the ServiceRegistry contract with Eureka REST API integration.
//
// 本包提供 Eureka 服务注册实现。
// 本文件实现 ServiceRegistry 契约，集成 Eureka REST API。
package eureka

import (
	"context"
	"errors"
	"sync"
	"time"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Registry 实现 transportcontract.ServiceRegistry，使用 Eureka 作为后端。
//
// 主要职责：
//   - 实现服务注册、注销、发现、监听契约方法
//   - 维护本地缓存和心跳续租
//   - 提供 Underlying/As 下探机制访问原生客户端
//
// 线程安全：
//   - 使用 sync.RWMutex 保护内部状态
//   - Watch 使用独立 cancel context 管理 goroutine 生命周期
type Registry struct {
	config *EurekaConfig
	client eurekaClient

	// 状态管理
	mu            sync.RWMutex
	registered    map[string]map[string]string  // 已注册实例的 metadata 缓存
	renewals      map[string]context.CancelFunc // 心跳续租 cancel 函数
	endpointCache map[string][]transportcontract.ServiceInstance
	watchCache    map[string]string
	closeMu       sync.Mutex
	closed        bool
	watchCancels  []context.CancelFunc
}

// NewRegistry creates a new Eureka registry with default HTTP client.
//
// NewRegistry 使用默认 HTTP 客户端创建新的 Eureka 注册中心。
func NewRegistry(cfg *EurekaConfig) (*Registry, error) {
	return NewRegistryWithClient(cfg, newHTTPEurekaClient())
}

// NewRegistryWithClient creates a new Eureka registry with custom client.
//
// NewRegistryWithClient 使用自定义客户端创建新的 Eureka 注册中心。
// 主要用于测试场景，可注入 fake client。
func NewRegistryWithClient(cfg *EurekaConfig, client eurekaClient) (*Registry, error) {
	if cfg.ServerURL == "" {
		return nil, ErrNoServerURL
	}
	if client == nil {
		return nil, errors.New("registry.eureka: client is required")
	}

	return &Registry{
		config:        cfg,
		client:        client,
		registered:    make(map[string]map[string]string),
		renewals:      make(map[string]context.CancelFunc),
		endpointCache: make(map[string][]transportcontract.ServiceInstance),
		watchCache:    make(map[string]string),
	}, nil
}

// Register registers a service instance to Eureka.
// Implements transportcontract.ServiceRegistry.Register.
//
// Register 将服务实例注册到 Eureka。
// 实现 transportcontract.ServiceRegistry.Register。
//
// 行为说明：
//   - 检查是否已注册，避免重复注册
//   - 调用底层 client 注册实例
//   - 缓存注册信息用于心跳续租恢复
//   - 启动心跳 goroutine（如配置了心跳间隔）
func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrRegistryClosed
	}
	key := instanceKey(name, addr)
	if _, exists := r.registered[key]; exists {
		return ErrAlreadyRegistered
	}
	if err := r.client.Register(ctx, r.config, name, addr, meta); err != nil {
		return err
	}
	r.registered[key] = cloneStringMap(meta)
	delete(r.endpointCache, name)
	delete(r.watchCache, name)
	r.startHeartbeatLocked(name, addr)
	return nil
}

// Deregister deregisters a service instance from Eureka.
// Implements transportcontract.ServiceRegistry.Deregister.
//
// Deregister 将服务实例从 Eureka 注销。
// 实现 transportcontract.ServiceRegistry.Deregister。
//
// 行为说明：
//   - 停止该实例的心跳 goroutine
//   - 调用底层 client 注销实例
//   - 清理本地缓存
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrRegistryClosed
	}
	if err := r.client.Deregister(ctx, r.config, name, addr); err != nil {
		return err
	}
	key := instanceKey(name, addr)
	if cancel, ok := r.renewals[key]; ok {
		cancel()
		delete(r.renewals, key)
	}
	delete(r.registered, key)
	delete(r.endpointCache, name)
	delete(r.watchCache, name)
	return nil
}

// Discover discovers service instances from Eureka.
// Implements transportcontract.ServiceRegistry.Discover.
//
// Discover 从 Eureka 发现服务实例。
// 实现 transportcontract.ServiceRegistry.Discover。
//
// 行为说明：
//   - 优先从本地缓存读取
//   - 缓存未命中则调用底层 client 发现
//   - 返回实例列表（已排序）
func (r *Registry) Discover(ctx context.Context, name string) ([]transportcontract.ServiceInstance, error) {
	r.mu.RLock()
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

	instances, err := r.client.Discover(ctx, r.config, name)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.endpointCache[name] = append([]transportcontract.ServiceInstance(nil), instances...)
	r.mu.Unlock()
	return instances, nil
}

// Watch watches service instance changes from Eureka.
// Implements transportcontract.ServiceRegistry.Watch.
//
// Watch 监听 Eureka 服务实例变更。
// 实现 transportcontract.ServiceRegistry.Watch。
//
// 行为说明：
//   - 返回 channel，变更时推送新实例列表
//   - 使用快照 key 去重，避免重复推送
//   - 首次立即发起 Discover 获取初始状态
//   - client.Watch 失败时自动重试
func (r *Registry) Watch(ctx context.Context, name string) (<-chan []transportcontract.ServiceInstance, error) {
	r.closeMu.Lock()
	defer r.closeMu.Unlock()

	if r.closed {
		return nil, ErrRegistryClosed
	}

	watchCtx, cancel := context.WithCancel(ctx)
	r.watchCancels = append(r.watchCancels, cancel)

	ch := make(chan []transportcontract.ServiceInstance, 10)
	var workers sync.WaitGroup

	// emit 用于推送实例变更，带去重逻辑
	emit := func(instances []transportcontract.ServiceInstance) bool {
		key := snapshotKey(instances)

		r.mu.Lock()
		last := r.watchCache[name]
		if last == key {
			r.mu.Unlock()
			return true
		}
		r.watchCache[name] = key
		if len(instances) == 0 {
			delete(r.endpointCache, name)
		} else {
			r.endpointCache[name] = append([]transportcontract.ServiceInstance(nil), instances...)
		}
		r.mu.Unlock()

		select {
		case ch <- append([]transportcontract.ServiceInstance(nil), instances...):
			return true
		case <-watchCtx.Done():
			return false
		default:
			return true
		}
	}

	// 主监听 goroutine：循环调用 client.Watch，失败重试
	workers.Add(1)
	go func() {
		defer workers.Done()
		for {
			err := r.client.Watch(watchCtx, r.config, name, func(instances []transportcontract.ServiceInstance) {
				emit(instances)
			})
			if err == nil || watchCtx.Err() != nil {
				return
			}
			if !isRetryableWatchError(err) {
				return
			}
			select {
			case <-watchCtx.Done():
				return
			case <-time.After(watchRetryInterval(r.config)):
			}
		}
	}()

	// 初始查询 goroutine：立即发起一次 Discover
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

	// 等待所有 worker 完成后关闭 channel
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

	// 停止所有心跳续租（renewals 受 r.mu 保护，需加锁）
	r.mu.Lock()
	for key, cancel := range r.renewals {
		cancel()
		delete(r.renewals, key)
	}
	r.mu.Unlock()

	// 停止所有 watch
	for _, cancel := range r.watchCancels {
		cancel()
	}
	r.watchCancels = nil
	return nil
}

// Underlying returns the current native client object used by this registry.
//
// Underlying 返回此注册中心使用的当前原生客户端对象。
// 用于需要访问底层 Eureka HTTP client 的场景。
func (r *Registry) Underlying() any {
	return r.client
}

// As projects the current native client into the requested target when possible.
//
// As 在可能时将当前原生客户端投射到请求的目标。
// 支持投射到 HTTPClientProvider 接口获取 http.Client。
func (r *Registry) As(target any) bool {
	return As(r.client, target)
}

// startHeartbeatLocked starts the heartbeat goroutine for an instance.
// Must be called with r.mu held.
//
// startHeartbeatLocked 为实例启动心跳 goroutine。
// 必须在持有 r.mu 时调用。
func (r *Registry) startHeartbeatLocked(name, addr string) {
	if r.config.HeartbeatInterval <= 0 {
		return
	}
	key := instanceKey(name, addr)
	if cancel, ok := r.renewals[key]; ok {
		cancel()
	}
	renewCtx, cancel := context.WithCancel(context.Background())
	r.renewals[key] = cancel

	go func() {
		RunHeartbeatLoop(renewCtx, HeartbeatLoopConfig{
			Interval:     r.config.HeartbeatInterval,
			RetryBackoff: r.config.HeartbeatRetryBackoff,
			Heartbeat: func(ctx context.Context) error {
				return r.client.Heartbeat(ctx, r.config, name, addr)
			},
			Recover: func(ctx context.Context) error {
				return r.recoverRegistration(ctx, name, addr)
			},
			ShouldRecover: func(err error) bool {
				return errors.Is(err, ErrServiceNotFound)
			},
		})
	}()
}

// recoverRegistration recovers a registration after heartbeat failure.
//
// recoverRegistration 在心跳失败后恢复注册。
// 从本地缓存获取 metadata 重新注册。
func (r *Registry) recoverRegistration(ctx context.Context, name, addr string) error {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return ErrRegistryClosed
	}
	meta := cloneStringMap(r.registered[instanceKey(name, addr)])
	r.mu.RUnlock()
	return r.client.Register(ctx, r.config, name, addr, meta)
}
