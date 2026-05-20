package dlock

import (
	"context"
	"time"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// DistributedLock is the top-level alias of the distributed lock contract.
// DistributedLock 是分布式锁契约的顶层别名。
type DistributedLock = datacontract.DistributedLock

// DistributedLockConfig is the top-level alias of the distributed lock config contract.
// DistributedLockConfig 是分布式锁配置契约的顶层别名。
type DistributedLockConfig = datacontract.DistributedLockConfig

// Get returns the distributed lock service from the container.
// Get 从容器获取分布式锁服务。
func Get(c runtimecontract.Container) (datacontract.DistributedLock, error) {
	return container.GetDistributedLock(c)
}

// GetOrPanic returns the distributed lock service from the container and panics on failure.
// GetOrPanic 从容器获取分布式锁服务，失败 panic。
func GetOrPanic(c runtimecontract.Container) datacontract.DistributedLock {
	return container.GetDistributedLockOrPanic(c)
}

// Lock acquires a lock using the distributed lock service from the container.
// Lock 使用容器中的分布式锁获取锁。
func Lock(ctx context.Context, c runtimecontract.Container, key string, ttl time.Duration) error {
	lockSvc, err := Get(c)
	if err != nil {
		return err
	}
	return lockSvc.Lock(ctx, key, ttl)
}

// TryLock tries to acquire a lock using the distributed lock service from the container.
// TryLock 使用容器中的分布式锁尝试获取锁。
func TryLock(ctx context.Context, c runtimecontract.Container, key string, ttl time.Duration) (bool, error) {
	lockSvc, err := Get(c)
	if err != nil {
		return false, err
	}
	return lockSvc.TryLock(ctx, key, ttl)
}

// Unlock releases a lock using the distributed lock service from the container.
// Unlock 使用容器中的分布式锁释放锁。
func Unlock(ctx context.Context, c runtimecontract.Container, key string) error {
	lockSvc, err := Get(c)
	if err != nil {
		return err
	}
	return lockSvc.Unlock(ctx, key)
}

// WithLock executes a function while holding a distributed lock.
// WithLock 使用容器中的分布式锁包裹执行函数。
//
// Example:
//
//	err := dlock.WithLock(ctx, c, "order:42", 5*time.Second, func() error {
//	    return processOrder(ctx, 42)
//	})
func WithLock(ctx context.Context, c runtimecontract.Container, key string, ttl time.Duration, fn func() error) error {
	lockSvc, err := Get(c)
	if err != nil {
		return err
	}
	return lockSvc.WithLock(ctx, key, ttl, fn)
}
