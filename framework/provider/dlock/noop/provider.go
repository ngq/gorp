package noop

import (
	"context"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop 分布式锁实现。
//
// 中文说明：
// - 单体项目默认使用此 provider；
// - 使用本地互斥锁实现；
// - 不跨进程，仅用于本地并发控制。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "dlock.noop" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.DistributedLockKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.DistributedLockKey, func(c contract.Container) (any, error) {
		return &noopLock{}, nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// noopLock 是 DistributedLock 的本地实现。
//
// 中文说明：
// - 使用 sync.Mutex 实现本地锁；
// - 不支持跨进程锁定；
// - 用于单体项目本地并发控制。
type noopLock struct {
	mu    sync.Mutex
	locks sync.Map // map[string]*sync.Mutex
}

// Lock 获取锁（阻塞）。
func (l *noopLock) Lock(ctx context.Context, key string, ttl time.Duration) error {
	lock := l.getOrCreateLock(key)

	// 尝试获取锁
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	lock.Lock()
	return nil
}

// TryLock 尝试获取锁（非阻塞）。
func (l *noopLock) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	lock := l.getOrCreateLock(key)

	// 尝试获取锁（非阻塞）
	locked := make(chan struct{})
	go func() {
		lock.Lock()
		close(locked)
	}()

	select {
	case <-locked:
		return true, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

// Unlock 释放锁。
func (l *noopLock) Unlock(ctx context.Context, key string) error {
	if lock, ok := l.locks.Load(key); ok {
		lock.(*sync.Mutex).Unlock()
	}
	return nil
}

// Renew 续约锁（本地锁不支持续约）。
func (l *noopLock) Renew(ctx context.Context, key string, ttl time.Duration) error {
	// 本地锁不需要续约
	return nil
}

// IsLocked 检查锁是否被持有。
func (l *noopLock) IsLocked(ctx context.Context, key string) (bool, error) {
	// 尝试获取锁来判断是否被持有
	lock := l.getOrCreateLock(key)

	locked := make(chan bool)
	go func() {
		if lock.TryLock() {
			lock.Unlock()
			locked <- false
		} else {
			locked <- true
		}
	}()

	select {
	case result := <-locked:
		return result, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

// WithLock 获取锁并执行函数。
func (l *noopLock) WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	if err := l.Lock(ctx, key, ttl); err != nil {
		return err
	}
	defer l.Unlock(ctx, key)

	return fn()
}

// getOrCreateLock 获取或创建锁。
func (l *noopLock) getOrCreateLock(key string) *sync.Mutex {
	if lock, ok := l.locks.Load(key); ok {
		return lock.(*sync.Mutex)
	}

	lock := &sync.Mutex{}
	l.locks.Store(key, lock)
	return lock
}