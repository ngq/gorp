// Package dlock 分布式锁封装
// 基于框架 contract.DistributedLock 能力
package dlock

import (
	"context"
	"fmt"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// Lock 分布式锁
type Lock struct {
	key    string
	locker contract.DistributedLock
}

// Release 释放锁
func (l *Lock) Release(ctx context.Context) error {
	return l.locker.Unlock(ctx, l.key)
}

// LockManager 锁管理器
type LockManager struct {
	locker contract.DistributedLock
}

// NewLockManager 创建锁管理器
func NewLockManager(locker contract.DistributedLock) *LockManager {
	return &LockManager{locker: locker}
}

// AcquireLock 获取锁
//
// 中文说明：
// - 获取指定 key 的分布式锁；
// - 支持超时时间；
// - 使用框架 DistributedLock 能力。
func AcquireLock(ctx context.Context, mgr *LockManager, key string, ttl time.Duration) (*Lock, error) {
	if err := mgr.locker.Lock(ctx, key, ttl); err != nil {
		return nil, err
	}
	return &Lock{key: key, locker: mgr.locker}, nil
}

// AcquireInventoryLock 获取库存锁
//
// 中文说明：
// - 为商品+仓库获取分布式锁；
// - 防止超卖；
// - 使用框架 DistributedLock 能力。
func AcquireInventoryLock(ctx context.Context, mgr *LockManager, productID, warehouseID uint64, ttl time.Duration) (*Lock, error) {
	key := inventoryLockKey(productID, warehouseID)
	return AcquireLock(ctx, mgr, key, ttl)
}

// inventoryLockKey 生成库存锁 key
func inventoryLockKey(productID, warehouseID uint64) string {
	return fmt.Sprintf("inventory:%d:%d", productID, warehouseID)
}