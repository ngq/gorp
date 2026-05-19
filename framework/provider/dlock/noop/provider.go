// Package noop provides a no-op distributed lock for monolith scenarios.
// This lock uses local sync.Mutex instead of distributed lock.
// Note: Only suitable for single-process monolith, not for distributed systems.
//
// 空分布式锁实现包，用于单体应用场景。
// 此锁使用本地 sync.Mutex 代替分布式锁。
// 注意：仅适用于单进程单体应用，不适用于分布式系统。
package noop

import (
	"context"
	"sync"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers a no-op distributed lock contract.
//
// Provider 注册空分布式锁契约。
type Provider struct{}

// NewProvider creates a new no-op lock provider instance.
//
// NewProvider 创建新的空分布式锁 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "dlock.noop".
//
// Name 返回 Provider 名称 "dlock.noop"。
func (p *Provider) Name() string { return "dlock.noop" }

// IsDefer returns true, lock can be deferred until first use.
//
// IsDefer 返回 true，锁可延迟初始化直到首次使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the distributed lock contract key.
//
// Provides 返回分布式锁契约键。
func (p *Provider) Provides() []string { return []string{datacontract.DistributedLockKey} }

// DependsOn returns the keys this provider depends on.
// Noop dlock has no dependencies.
//
// DependsOn 返回该 provider 依赖的 key。
// Noop dlock 无依赖。
func (p *Provider) DependsOn() []string { return nil }

// Register binds the no-op lock to the container.
//
// Register 将空锁绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.DistributedLockKey, func(c runtimecontract.Container) (any, error) {
		return &noopLock{}, nil
	}, true)
	return nil
}

// Boot is a no-op for this provider.
//
// Boot 此 Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// noopLock implements datacontract.DistributedLock using local sync.Mutex.
// 注意：locks 中的 mutex 在 Unlock 后会从 sync.Map 中删除，避免内存膨胀。
// 但在高并发场景下，频繁创建/删除 mutex 可能影响性能，可考虑保留。
//
// noopLock 使用本地 sync.Mutex 实现 datacontract.DistributedLock 接口。
type noopLock struct {
	locks sync.Map // locks stores per-key mutexes.
	//
	// locks 存储每个键的互斥锁。
}

// Lock acquires a local mutex lock.
//
// Lock 获取本地互斥锁。
func (l *noopLock) Lock(ctx context.Context, key string, ttl time.Duration) error {
	_ = ttl
	lock := l.getOrCreateLock(key)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	lock.Lock()
	return nil
}

// TryLock attempts to acquire lock, returns immediately if locked by others.
// 使用 sync.Mutex.TryLock 避免 goroutine 泄漏。
//
// TryLock 尝试获取锁，如果已被锁定则立即返回。
func (l *noopLock) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	_ = ttl
	lock := l.getOrCreateLock(key)

	if lock.TryLock() {
		return true, nil
	}
	return false, nil
}

// Unlock releases the local mutex lock and removes it from the map to prevent memory bloat.
// Note: This may cause performance overhead in high-concurrency scenarios due to frequent mutex creation/deletion.
//
// Unlock 释放本地互斥锁并从 map 中删除，防止内存膨胀。
// 注意：高并发场景下频繁创建/删除 mutex 可能影响性能。
func (l *noopLock) Unlock(ctx context.Context, key string) error {
	_ = ctx
	if lock, ok := l.locks.Load(key); ok {
		l.locks.Delete(key) // 删除锁记录，防止内存膨胀
		lock.(*sync.Mutex).Unlock()
	}
	return nil
}

// Renew does nothing (TTL not supported in local lock).
//
// Renew 不执行任何操作（本地锁不支持 TTL）。
func (l *noopLock) Renew(ctx context.Context, key string, ttl time.Duration) error {
	_ = ctx
	_ = key
	_ = ttl
	return nil
}

// IsLocked checks if the key is currently locked.
// 使用 sync.Mutex.TryLock 避免 goroutine 泄漏。
//
// IsLocked 检查键是否当前被锁定。
func (l *noopLock) IsLocked(ctx context.Context, key string) (bool, error) {
	lock := l.getOrCreateLock(key)

	// TryLock 成功说明之前未被锁定，立即释放
	if lock.TryLock() {
		lock.Unlock()
		return false, nil
	}
	return true, nil
}

// WithLock acquires lock, executes function, then releases lock.
//
// WithLock 获取锁、执行函数、然后释放锁。
func (l *noopLock) WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	if err := l.Lock(ctx, key, ttl); err != nil {
		return err
	}
	defer l.Unlock(ctx, key)
	return fn()
}

// getOrCreateLock gets or creates a mutex for the given key.
//
// getOrCreateLock 获取或创建给定键的互斥锁。
func (l *noopLock) getOrCreateLock(key string) *sync.Mutex {
	if lock, ok := l.locks.Load(key); ok {
		return lock.(*sync.Mutex)
	}

	lock := &sync.Mutex{}
	l.locks.Store(key, lock)
	return lock
}
