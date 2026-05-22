// Application scenarios:
// - Define the distributed lock contract shared by lock providers and business code.
// - Standardize lock acquisition, renewal, release, and guarded execution semantics.
// - Provide a common config model for choosing lock backends and retry behavior.
//
// 适用场景：
// - 定义分布式锁 provider 和业务代码共同依赖的锁契约。
// - 统一加锁、续租、解锁和受保护执行语义。
// - 为锁后端选择和重试行为提供通用配置模型。
package data

import (
	"context"
	"time"
)

// DistributedLockKey is the container key for the distributed lock capability.
//
// DistributedLockKey 是分布式锁能力的容器键。
const DistributedLockKey = "framework.distributed_lock"

// DistributedLock defines the distributed lock operations exposed by the framework.
//
// DistributedLock 定义框架对外暴露的分布式锁操作。
type DistributedLock interface {
	// Lock blocks until the lock is acquired or the context is canceled.
	//
	// Lock 会阻塞直到成功获取锁或 context 被取消。
	Lock(ctx context.Context, key string, ttl time.Duration) error

	// TryLock attempts to acquire the lock immediately.
	//
	// TryLock 尝试立即获取锁。
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// Unlock releases the lock identified by key.
	//
	// Unlock 释放指定 key 对应的锁。
	Unlock(ctx context.Context, key string) error

	// Renew extends the TTL of the lock.
	//
	// Renew 延长锁的 TTL。
	Renew(ctx context.Context, key string, ttl time.Duration) error

	// IsLocked reports whether the target key is currently locked.
	//
	// IsLocked 返回目标 key 当前是否处于加锁状态。
	IsLocked(ctx context.Context, key string) (bool, error)

	// WithLock executes the callback while holding the lock.
	// Guarantees that the lock is released after fn returns (even on panic),
	// using deferred unlock semantics.
	//
	// WithLock 在持有锁期间执行回调。
	// 保证在 fn 返回后（即使 panic）自动释放锁，使用 defer unlock 语义。
	WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error
}

// DistributedLockConfig describes distributed lock backend settings.
//
// DistributedLockConfig 描述分布式锁后端配置。
type DistributedLockConfig struct {
	Type string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	EtcdEndpoints []string
	EtcdUsername  string
	EtcdPassword  string

	DefaultTTL    time.Duration
	RetryInterval time.Duration
	MaxRetry      int
	KeyPrefix     string
}
