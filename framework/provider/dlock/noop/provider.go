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

// noopLock implements datacontract.DistributedLock using per-key local locks.
// Entries are kept in the map for the process lifetime: deleting them on
// unlock (the previous behavior) allowed two goroutines to hold "the lock"
// simultaneously and let one goroutine cascade-release another's lock. Key
// cardinality is bounded by the application's lock keys, so retention is
// the safe trade-off.
//
// noopLock 使用按 key 的本地锁实现 datacontract.DistributedLock 接口。
type noopLock struct {
	locks sync.Map // locks stores per-key *lockEntry, retained forever.
}

// lockEntry is a local mutex that also honors context cancellation while
// blocked, which sync.Mutex cannot do.
//
// lockEntry 是支持阻塞期间响应 ctx 取消的本地互斥锁。
type lockEntry struct {
	mu      sync.Mutex
	locked  bool
	waiters []chan struct{}
}

// Lock acquires the entry, blocking until available, the context is done, or
// ownership is handed over by a previous holder.
func (e *lockEntry) Lock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	e.mu.Lock()
	if !e.locked {
		e.locked = true
		e.mu.Unlock()
		return nil
	}
	ch := make(chan struct{})
	e.waiters = append(e.waiters, ch)
	e.mu.Unlock()

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		// If the hand-over already happened we own the lock and must give it
		// back; otherwise just leave the waiter queue.
		select {
		case <-ch:
			e.Unlock()
		default:
			e.removeWaiter(ch)
		}
		return ctx.Err()
	}
}

func (e *lockEntry) removeWaiter(ch chan struct{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, w := range e.waiters {
		if w == ch {
			e.waiters = append(e.waiters[:i], e.waiters[i+1:]...)
			return
		}
	}
}

// TryLock acquires the entry without blocking.
func (e *lockEntry) TryLock() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.locked {
		return false
	}
	e.locked = true
	return true
}

// Unlock releases the entry, handing ownership to the longest-waiting waiter
// when one exists (the entry stays locked across the hand-over).
func (e *lockEntry) Unlock() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.locked {
		return
	}
	if len(e.waiters) > 0 {
		ch := e.waiters[0]
		e.waiters = e.waiters[1:]
		close(ch)
		return
	}
	e.locked = false
}

// IsLocked reports whether the entry is currently held.
func (e *lockEntry) IsLocked() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.locked
}

// Lock acquires a local mutex lock, honoring ctx cancellation while blocked.
// TTL is ignored: this is a single-process stand-in for a real distributed lock.
//
// Lock 获取本地互斥锁，阻塞期间响应 ctx 取消。TTL 不生效：
// 这是真实分布式锁的单进程替身。
func (l *noopLock) Lock(ctx context.Context, key string, ttl time.Duration) error {
	_ = ttl
	if err := ctx.Err(); err != nil {
		return err
	}
	return l.getOrCreateLock(key).Lock(ctx)
}

// TryLock attempts to acquire lock, returns immediately if locked by others.
//
// TryLock 尝试获取锁，如果已被锁定则立即返回。
func (l *noopLock) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	_ = ttl
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return l.getOrCreateLock(key).TryLock(), nil
}

// Unlock releases the local mutex lock. The map entry is intentionally kept:
// deleting it would let a concurrent Lock create a second live lock for the
// same key and break mutual exclusion.
//
// Unlock 释放本地互斥锁。map 条目有意保留：删除它会让并发的 Lock
// 为同一 key 创建第二个活跃锁，破坏互斥性。
func (l *noopLock) Unlock(ctx context.Context, key string) error {
	_ = ctx
	if entry, ok := l.locks.Load(key); ok {
		entry.(*lockEntry).Unlock()
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
//
// IsLocked 检查键是否当前被锁定。
func (l *noopLock) IsLocked(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if entry, ok := l.locks.Load(key); ok {
		return entry.(*lockEntry).IsLocked(), nil
	}
	return false, nil
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

// getOrCreateLock gets or creates the lock entry for the given key.
// LoadOrStore is mandatory: separate Load+Store lets two concurrent callers
// end up with different entries and both "hold" the lock.
//
// getOrCreateLock 获取或创建给定键的锁条目。
// 必须用 LoadOrStore：分离的 Load+Store 会让两个并发调用者
// 拿到不同条目并同时"持有"锁。
func (l *noopLock) getOrCreateLock(key string) *lockEntry {
	if entry, ok := l.locks.Load(key); ok {
		return entry.(*lockEntry)
	}
	actual, _ := l.locks.LoadOrStore(key, &lockEntry{})
	return actual.(*lockEntry)
}
