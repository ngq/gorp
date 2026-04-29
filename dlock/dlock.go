package dlock

import (
	"context"
	"time"

	contribredis "github.com/ngq/gorp/contrib/dlock/redis"
	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
)

type DistributedLock = contract.DistributedLock
type DistributedLockConfig = contract.DistributedLockConfig

// Make 从容器获取分布式锁服务。
func Make(c contract.Container) (contract.DistributedLock, error) {
	return container.MakeDistributedLock(c)
}

// MustMake 从容器获取分布式锁服务，失败 panic。
func MustMake(c contract.Container) contract.DistributedLock {
	return container.MustMakeDistributedLock(c)
}

// NewRedisLock 创建 Redis 分布式锁实现。
func NewRedisLock(cfg *contract.DistributedLockConfig) (contract.DistributedLock, error) {
	return contribredis.NewLock(cfg)
}

// Lock 使用容器中的分布式锁获取锁。
func Lock(ctx context.Context, c contract.Container, key string, ttl time.Duration) error {
	lockSvc, err := Make(c)
	if err != nil {
		return err
	}
	return lockSvc.Lock(ctx, key, ttl)
}

// TryLock 使用容器中的分布式锁尝试获取锁。
func TryLock(ctx context.Context, c contract.Container, key string, ttl time.Duration) (bool, error) {
	lockSvc, err := Make(c)
	if err != nil {
		return false, err
	}
	return lockSvc.TryLock(ctx, key, ttl)
}

// Unlock 使用容器中的分布式锁释放锁。
func Unlock(ctx context.Context, c contract.Container, key string) error {
	lockSvc, err := Make(c)
	if err != nil {
		return err
	}
	return lockSvc.Unlock(ctx, key)
}

// WithLock 使用容器中的分布式锁包裹执行函数。
func WithLock(ctx context.Context, c contract.Container, key string, ttl time.Duration, fn func() error) error {
	lockSvc, err := Make(c)
	if err != nil {
		return err
	}
	return lockSvc.WithLock(ctx, key, ttl, fn)
}
