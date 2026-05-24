package dlock

import (
	"context"
	"time"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// DistributedLock 是分布式锁契约的顶层别名。
// DistributedLock is the top-level alias of the distributed lock contract.
type DistributedLock = datacontract.DistributedLock

// DistributedLockConfig 是分布式锁配置契约的顶层别名。
// DistributedLockConfig is the top-level alias of the distributed lock config contract.
type DistributedLockConfig = datacontract.DistributedLockConfig

// resolveDLock 从 context 解析容器并获取分布式锁服务实例。
// 内部通过 frameworkcontainer.Resolve(ctx) 获取容器，
// 再通过 container.MakeWith 解析 datacontract.DistributedLockKey 对应的服务。
//
// resolveDLock resolves the Container from ctx and retrieves the distributed lock
// service instance using frameworkcontainer.Resolve + container.MakeWith.
func resolveDLock(ctx context.Context) (datacontract.DistributedLock, error) {
	cont := container.Resolve(ctx)
	return container.MakeWith[datacontract.DistributedLock](cont, datacontract.DistributedLockKey)
}

// GetService 从容器中获取分布式锁服务实例。
// 通过 context 解析容器，再从容器中解析 DistributedLockKey 对应的服务。
// 如果容器未注册分布式锁服务，返回错误。
//
// GetService retrieves the distributed lock service from the container resolved via ctx.
// Returns an error if the service is not registered.
func GetService(ctx context.Context) (datacontract.DistributedLock, error) {
	return resolveDLock(ctx)
}

// MustGetService 从容器中获取分布式锁服务实例，失败时 panic。
// 适用于确定已注册分布式锁服务的场景，简化错误处理。
//
// MustGetService retrieves the distributed lock service from the container resolved via ctx.
// Panics if the service is not registered.
func MustGetService(ctx context.Context) datacontract.DistributedLock {
	svc, err := resolveDLock(ctx)
	if err != nil {
		panic(err)
	}
	return svc
}

// Lock 使用分布式锁服务获取锁。
// 通过 context 解析容器获取分布式锁服务，然后调用其 Lock 方法。
// key 为锁的标识，ttl 为锁的存活时间。
//
// Lock acquires a lock using the distributed lock service resolved from ctx.
// key is the lock identifier, ttl is the time-to-live of the lock.
func Lock(ctx context.Context, key string, ttl time.Duration) error {
	lockSvc, err := resolveDLock(ctx)
	if err != nil {
		return err
	}
	return lockSvc.Lock(ctx, key, ttl)
}

// TryLock 使用分布式锁服务尝试获取锁。
// 通过 context 解析容器获取分布式锁服务，然后调用其 TryLock 方法。
// 返回是否成功获取锁及可能的错误。
//
// TryLock tries to acquire a lock using the distributed lock service resolved from ctx.
// Returns whether the lock was acquired and any error encountered.
func TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	lockSvc, err := resolveDLock(ctx)
	if err != nil {
		return false, err
	}
	return lockSvc.TryLock(ctx, key, ttl)
}

// Unlock 使用分布式锁服务释放锁。
// 通过 context 解析容器获取分布式锁服务，然后调用其 Unlock 方法。
//
// Unlock releases a lock using the distributed lock service resolved from ctx.
func Unlock(ctx context.Context, key string) error {
	lockSvc, err := resolveDLock(ctx)
	if err != nil {
		return err
	}
	return lockSvc.Unlock(ctx, key)
}

// WithLock 使用分布式锁包裹执行函数，执行完毕后自动释放锁。
// 通过 context 解析容器获取分布式锁服务，获取锁后执行 fn，执行完毕后释放锁。
//
// 示例:
//
//	err := dlock.WithLock(ctx, "order:42", 5*time.Second, func() error {
//	    return processOrder(ctx, 42)
//	})
//
// WithLock executes fn while holding a distributed lock resolved from ctx.
// The lock is released after fn completes.
//
// Example:
//
//	err := dlock.WithLock(ctx, "order:42", 5*time.Second, func() error {
//	    return processOrder(ctx, 42)
//	})
func WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	lockSvc, err := resolveDLock(ctx)
	if err != nil {
		return err
	}
	return lockSvc.WithLock(ctx, key, ttl, fn)
}
