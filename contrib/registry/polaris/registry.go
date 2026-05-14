// Package polaris provides Polaris service registry implementation.
// This file implements the ServiceRegistry contract with Polaris SDK integration.
//
// 本包提供 Polaris 服务注册实现。
// 本文件实现 ServiceRegistry 契约，集成 Polaris SDK。
package polaris

import (
	"context"
	"errors"
	"sync"
	"time"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// ErrNoAddress indicates Polaris address is required.
//
// ErrNoAddress 表示 Polaris 地址必需。
var ErrNoAddress = errors.New("registry.polaris: address is required")

// ErrServiceNotFound indicates Polaris service not found.
//
// ErrServiceNotFound 表示 Polaris 服务未找到。
var ErrServiceNotFound = errors.New("registry.polaris: service not found")

// ErrRegistryClosed indicates Polaris registry closed.
//
// ErrRegistryClosed 表示 Polaris 注册中心已关闭。
var ErrRegistryClosed = errors.New("registry.polaris: registry closed")

// ErrAlreadyRegistered indicates Polaris instance already registered.
//
// ErrAlreadyRegistered 表示 Polaris 实例已注册。
var ErrAlreadyRegistered = errors.New("registry.polaris: instance already registered")

// Registry implements transportcontract.ServiceRegistry with Polaris SDK.
// Supports service registration, discovery, and watch with caching.
//
// Registry 使用 Polaris SDK 实现 transportcontract.ServiceRegistry。
// 支持服务注册、发现和监听，带缓存功能。
type Registry struct {
	config *PolarisConfig
	client polarisRegistryClient

	mu            sync.RWMutex
	endpointCache map[string][]transportcontract.ServiceInstance
	registered    map[string]struct{}
	closeMu       sync.Mutex
	closed        bool
	watchCancels  []context.CancelFunc
}

// NewRegistry creates a new Polaris registry with default official client.
//
// NewRegistry 使用默认官方客户端创建新的 Polaris 注册中心。
func NewRegistry(cfg *PolarisConfig) (*Registry, error) {
	return NewRegistryWithClient(cfg, newOfficialPolarisRegistryClient())
}

// NewRegistryWithClient creates a new Polaris registry with custom client.
//
// NewRegistryWithClient 使用自定义客户端创建新的 Polaris 注册中心。
func NewRegistryWithClient(cfg *PolarisConfig, client polarisRegistryClient) (*Registry, error) {
	if cfg.Address == "" {
		return nil, ErrNoAddress
	}
	if client == nil {
		return nil, errors.New("registry.polaris: registry client is required")
	}
	return &Registry{
		config:        cfg,
		client:        client,
		endpointCache: make(map[string][]transportcontract.ServiceInstance),
		registered:    make(map[string]struct{}),
	}, nil
}

// Register registers a service instance to Polaris.
// Implements transportcontract.ServiceRegistry.Register.
//
// Register 将服务实例注册到 Polaris。
// 实现 transportcontract.ServiceRegistry.Register。
func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrRegistryClosed
	}
	key := name + "|" + addr
	if _, exists := r.registered[key]; exists {
		return ErrAlreadyRegistered
	}
	if err := r.client.Register(ctx, r.config, name, addr, meta); err != nil {
		return err
	}
	r.registered[key] = struct{}{}
	delete(r.endpointCache, name)
	return nil
}

// Deregister deregisters a service instance from Polaris.
// Implements transportcontract.ServiceRegistry.Deregister.
//
// Deregister 将服务实例从 Polaris 注销。
// 实现 transportcontract.ServiceRegistry.Deregister。
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrRegistryClosed
	}
	if err := r.client.Deregister(ctx, r.config, name, addr); err != nil {
		return err
	}
	delete(r.registered, name+"|"+addr)
	delete(r.endpointCache, name)
	return nil
}

// Discover discovers service instances from Polaris.
// Implements transportcontract.ServiceRegistry.Discover.
//
// Discover 从 Polaris 发现服务实例。
// 实现 transportcontract.ServiceRegistry.Discover。
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
	sortServiceInstances(instances)

	r.mu.Lock()
	r.endpointCache[name] = append([]transportcontract.ServiceInstance(nil), instances...)
	r.mu.Unlock()
	return instances, nil
}

// Watch watches service instance changes from Polaris.
// Implements transportcontract.ServiceRegistry.Watch.
//
// Watch 监听 Polaris 服务实例变更。
// 实现 transportcontract.ServiceRegistry.Watch。
func (r *Registry) Watch(ctx context.Context, name string) (<-chan []transportcontract.ServiceInstance, error) {
	r.closeMu.Lock()
	defer r.closeMu.Unlock()

	if r.closed {
		return nil, ErrRegistryClosed
	}

	watchCtx, cancel := context.WithCancel(ctx)
	r.watchCancels = append(r.watchCancels, cancel)

	ch := make(chan []transportcontract.ServiceInstance, 10)
	var emitMu sync.Mutex
	lastSnapshot := ""
	emit := func(instances []transportcontract.ServiceInstance) {
		sorted := append([]transportcontract.ServiceInstance(nil), instances...)
		sortServiceInstances(sorted)
		current := snapshotKey(sorted)

		emitMu.Lock()
		if current == lastSnapshot {
			emitMu.Unlock()
			return
		}
		lastSnapshot = current
		emitMu.Unlock()

		r.mu.Lock()
		if len(sorted) == 0 {
			delete(r.endpointCache, name)
		} else {
			r.endpointCache[name] = append([]transportcontract.ServiceInstance(nil), sorted...)
		}
		r.mu.Unlock()

		select {
		case ch <- append([]transportcontract.ServiceInstance(nil), sorted...):
		case <-watchCtx.Done():
		default:
		}
	}

	go func() {
		defer close(ch)
		for {
			err := r.client.Watch(watchCtx, r.config, name, func(instances []transportcontract.ServiceInstance) {
				emit(instances)
			})
			if err == nil || watchCtx.Err() != nil {
				return
			}
			select {
			case <-watchCtx.Done():
				return
			case <-time.After(r.config.WatchRetryInterval):
			}
		}
	}()

	go func() {
		instances, err := r.Discover(watchCtx, name)
		if err != nil {
			if errors.Is(err, ErrServiceNotFound) {
				emit([]transportcontract.ServiceInstance{})
			}
			return
		}
		emit(instances)
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
	for _, cancel := range r.watchCancels {
		cancel()
	}
	r.watchCancels = nil
	if closer, ok := r.client.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

// Underlying returns the current native client object used by this registry.
//
// Underlying 返回此注册中心使用的当前原生客户端对象。
func (r *Registry) Underlying() any {
	if provider, ok := r.client.(polarisRegistryNativeClient); ok {
		if native := provider.Underlying(); native != nil {
			return native
		}
	}
	return r.client
}

// As projects the current native client into the requested target when possible.
//
// As 在可能时将当前原生客户端投射到请求的目标。
func (r *Registry) As(target any) bool {
	return internalnative.As(r.Underlying(), target)
}

// polarisRegistryClient defines the internal client interface for registry operations.
//
// polarisRegistryClient 定义注册操作的内部客户端接口。
type polarisRegistryClient interface {
	Register(ctx context.Context, cfg *PolarisConfig, name, addr string, meta map[string]string) error
	Deregister(ctx context.Context, cfg *PolarisConfig, name, addr string) error
	Discover(ctx context.Context, cfg *PolarisConfig, name string) ([]transportcontract.ServiceInstance, error)
	Watch(ctx context.Context, cfg *PolarisConfig, name string, onUpdate func([]transportcontract.ServiceInstance)) error
}

// polarisRegistryNativeClient defines the interface for accessing underlying client.
//
// polarisRegistryNativeClient 定义访问底层客户端的接口。
type polarisRegistryNativeClient interface {
	Underlying() any
}
