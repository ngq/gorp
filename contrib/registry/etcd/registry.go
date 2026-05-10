// Package etcd provides etcd service registry implementation.
// This file implements the ServiceRegistry contract with etcd SDK integration.
//
// 本包提供 etcd 服务注册实现。
// 本文件实现 ServiceRegistry 契约，集成 etcd SDK。
package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// ErrRegistryClosed indicates etcd registry closed.
//
// ErrRegistryClosed 表示 etcd 注册中心已关闭。
var ErrRegistryClosed = errors.New("discovery.etcd: registry closed")

// Registry implements transportcontract.ServiceRegistry with etcd SDK.
// Supports service registration, discovery with lease TTL keepalive.
//
// Registry 使用 etcd SDK 实现 transportcontract.ServiceRegistry。
// 支持服务注册、发现，带租约 TTL keepalive 功能。
type Registry struct {
	cfg    *DiscoveryConfig
	client etcdRegistryClient

	registered sync.Map
	mu         sync.Mutex
	closed     bool
}

// registeredService tracks a registered service instance.
//
// registeredService 跟踪已注册的服务实例。
type registeredService struct {
	serviceID  string
	name       string
	addr       string
	meta       map[string]string
	serviceKey string
	leaseID    clientv3.LeaseID
	stopCh     chan struct{}
}

// NewRegistry creates a new etcd registry with default live client.
//
// NewRegistry 使用默认真实客户端创建新的 etcd 注册中心。
func NewRegistry(cfg *DiscoveryConfig) (*Registry, error) {
	client, err := newLiveEtcdRegistryClient(cfg)
	if err != nil {
		return nil, err
	}
	return NewRegistryWithClient(cfg, client), nil
}

// NewRegistryWithClient creates a new etcd registry with custom client.
//
// NewRegistryWithClient 使用自定义客户端创建新的 etcd 注册中心。
func NewRegistryWithClient(cfg *DiscoveryConfig, client etcdRegistryClient) *Registry {
	return &Registry{
		cfg:    cfg,
		client: client,
	}
}

// Register registers a service instance to etcd.
// Implements transportcontract.ServiceRegistry.Register.
//
// Register 将服务实例注册到 etcd。
// 实现 transportcontract.ServiceRegistry.Register。
func (r *Registry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrRegistryClosed
	}

	host, port := parseAddr(addr)
	if port == 0 {
		port = r.cfg.ServicePort
	}
	serviceID := generateServiceID(name, host, port)

	if _, ok := r.registered.Load(serviceID); ok {
		return nil
	}

	return r.registerLocked(ctx, serviceID, name, addr, meta)
}

// registerLocked performs the actual registration under lock.
//
// registerLocked 在锁保护下执行实际注册逻辑。
func (r *Registry) registerLocked(ctx context.Context, serviceID, name, addr string, meta map[string]string) error {
	leaseID, err := r.client.Grant(ctx, r.cfg.LeaseTTL)
	if err != nil {
		return fmt.Errorf("discovery.etcd: create lease failed: %w", err)
	}

	host, port := parseAddr(addr)
	if port == 0 {
		port = r.cfg.ServicePort
	}

	// 构建服务信息 JSON 数据
	serviceInfo := map[string]any{
		"name":          name,
		"address":       host,
		"port":          port,
		"meta":          meta,
		"healthy":       true,
		"registered_at": timeNow().Unix(),
	}
	serviceData, _ := json.Marshal(serviceInfo)

	serviceKey := path.Join(r.cfg.ServicePath, name, serviceID)
	if err := r.client.Put(ctx, serviceKey, string(serviceData), leaseID); err != nil {
		return fmt.Errorf("discovery.etcd: put service failed: %w", err)
	}

	// 启动 keepalive 循环
	stopCh := make(chan struct{})
	keepAliveCh, err := r.client.KeepAlive(context.Background(), leaseID)
	if err != nil {
		return fmt.Errorf("discovery.etcd: start keepalive failed: %w", err)
	}

	go r.keepAliveLoop(serviceID, keepAliveCh, stopCh)

	r.registered.Store(serviceID, &registeredService{
		serviceID:  serviceID,
		name:       name,
		addr:       addr,
		meta:       cloneStringMap(meta),
		serviceKey: serviceKey,
		leaseID:    leaseID,
		stopCh:     stopCh,
	})

	return nil
}

// Deregister deregisters a service instance from etcd.
// Implements transportcontract.ServiceRegistry.Deregister.
//
// Deregister 将服务实例从 etcd 注销。
// 实现 transportcontract.ServiceRegistry.Deregister。
func (r *Registry) Deregister(ctx context.Context, name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrRegistryClosed
	}

	host, port := parseAddr(addr)
	if port == 0 {
		port = r.cfg.ServicePort
	}
	serviceID := generateServiceID(name, host, port)

	if cached, ok := r.registered.Load(serviceID); ok {
		reg := cached.(*registeredService)
		close(reg.stopCh)
		_ = r.client.Revoke(ctx, reg.leaseID)
		r.registered.Delete(serviceID)
	}

	return nil
}

// Discover discovers service instances from etcd.
// Implements transportcontract.ServiceRegistry.Discover.
//
// Discover 从 etcd 发现服务实例。
// 实现 transportcontract.ServiceRegistry.Discover。
func (r *Registry) Discover(ctx context.Context, name string) ([]transportcontract.ServiceInstance, error) {
	if r.isClosed() {
		return nil, ErrRegistryClosed
	}

	servicePrefix := path.Join(r.cfg.ServicePath, name) + "/"
	resp, err := r.client.Get(ctx, servicePrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("discovery.etcd: get services failed: %w", err)
	}

	instances := make([]transportcontract.ServiceInstance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var info map[string]any
		if err := json.Unmarshal(kv.Value, &info); err != nil {
			continue
		}

		serviceName := getString(info, "name")
		address := getString(info, "address")
		port := getInt(info, "port")
		healthy := getBool(info, "healthy")
		meta := getMap(info, "meta")
		fullAddr := fmt.Sprintf("%s:%d", address, port)
		serviceID := strings.TrimPrefix(string(kv.Key), servicePrefix)

		instances = append(instances, transportcontract.ServiceInstance{
			ID:       serviceID,
			Name:     serviceName,
			Address:  fullAddr,
			Metadata: meta,
			Healthy:  healthy,
		})
	}

	// 应用负载均衡策略
	if len(instances) > 1 {
		instances = r.applyLoadBalance(instances)
	}

	return instances, nil
}

// Close releases resources held by the registry.
// Implements transportcontract.ServiceRegistry.Close.
//
// Close 释放注册中心持有的资源。
// 实现 transportcontract.ServiceRegistry.Close。
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}
	r.closed = true

	// 遍历所有已注册服务，停止 keepalive 并撤销租约
	r.registered.Range(func(key, value any) bool {
		reg := value.(*registeredService)
		close(reg.stopCh)
		_ = r.client.Revoke(context.Background(), reg.leaseID)
		return true
	})

	return r.client.Close()
}

// Underlying returns the current native client object used by this registry.
//
// Underlying 返回此注册中心使用的当前原生客户端对象。
func (r *Registry) Underlying() any {
	if provider, ok := r.client.(etcdRegistryNativeClient); ok {
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

// keepAliveLoop runs the lease keepalive loop.
//
// keepAliveLoop 运行租约 keepalive 循环。
func (r *Registry) keepAliveLoop(serviceID string, keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse, stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case resp, ok := <-keepAliveCh:
			if !ok {
				// keepalive 通道关闭，尝试重新注册
				r.tryReRegister(serviceID, stopCh)
				return
			}
			if resp != nil {
				continue
			}
		}
	}
}

// tryReRegister attempts to re-register a service when keepalive fails.
//
// tryReRegister 当 keepalive 失败时尝试重新注册服务。
func (r *Registry) tryReRegister(serviceID string, stopCh chan struct{}) {
	select {
	case <-stopCh:
		return
	default:
	}
	if r.isClosed() {
		return
	}

	cached, ok := r.registered.Load(serviceID)
	if !ok {
		return
	}
	reg := cached.(*registeredService)
	if reg.stopCh != stopCh {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*timeSecond())
	defer cancel()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}

	cached, ok = r.registered.Load(serviceID)
	if !ok {
		return
	}
	reg = cached.(*registeredService)
	if reg.stopCh != stopCh {
		return
	}

	// 删除旧注册记录，尝试重新注册
	r.registered.Delete(serviceID)
	if err := r.registerLocked(ctx, serviceID, reg.name, reg.addr, reg.meta); err != nil {
		r.registered.Store(serviceID, reg)
	}
}

// isClosed checks if registry is closed.
//
// isClosed 检查注册中心是否已关闭。
func (r *Registry) isClosed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.closed
}

// applyLoadBalance applies load balance strategy to instances.
//
// applyLoadBalance 对实例应用负载均衡策略。
func (r *Registry) applyLoadBalance(instances []transportcontract.ServiceInstance) []transportcontract.ServiceInstance {
	switch r.cfg.LoadBalance {
	case "random":
		randShuffle(len(instances), func(i, j int) {
			instances[i], instances[j] = instances[j], instances[i]
		})
	}
	return instances
}

// etcdRegistryClient defines the internal client interface for registry operations.
//
// etcdRegistryClient 定义注册操作的内部客户端接口。
type etcdRegistryClient interface {
	Grant(ctx context.Context, ttl int64) (clientv3.LeaseID, error)
	Put(ctx context.Context, key, value string, leaseID clientv3.LeaseID) error
	KeepAlive(ctx context.Context, leaseID clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error)
	Revoke(ctx context.Context, leaseID clientv3.LeaseID) error
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Close() error
}

// etcdRegistryNativeClient defines the interface for accessing underlying client.
//
// etcdRegistryNativeClient 定义访问底层客户端的接口。
type etcdRegistryNativeClient interface {
	Underlying() any
}