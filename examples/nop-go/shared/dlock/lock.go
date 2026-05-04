// Package dlock 鍒嗗竷寮忛攣灏佽
// 鍩轰簬妗嗘灦 datacontract.DistributedLock 鑳藉姏
package dlock

import (
	"context"
	"fmt"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// Lock 鍒嗗竷寮忛攣
type Lock struct {
	key    string
	locker datacontract.DistributedLock
}

// Release 閲婃斁閿?
func (l *Lock) Release(ctx context.Context) error {
	return l.locker.Unlock(ctx, l.key)
}

// LockManager 閿佺鐞嗗櫒
type LockManager struct {
	locker datacontract.DistributedLock
}

// NewLockManager 鍒涘缓閿佺鐞嗗櫒
func NewLockManager(locker datacontract.DistributedLock) *LockManager {
	return &LockManager{locker: locker}
}

// AcquireLock 鑾峰彇閿?
//
// 涓枃璇存槑锛?
// - 鑾峰彇鎸囧畾 key 鐨勫垎甯冨紡閿侊紱
// - 鏀寔瓒呮椂鏃堕棿锛?
// - 浣跨敤妗嗘灦 DistributedLock 鑳藉姏銆?
func AcquireLock(ctx context.Context, mgr *LockManager, key string, ttl time.Duration) (*Lock, error) {
	if err := mgr.locker.Lock(ctx, key, ttl); err != nil {
		return nil, err
	}
	return &Lock{key: key, locker: mgr.locker}, nil
}

// AcquireInventoryLock 鑾峰彇搴撳瓨閿?
//
// 涓枃璇存槑锛?
// - 涓哄晢鍝?浠撳簱鑾峰彇鍒嗗竷寮忛攣锛?
// - 闃叉瓒呭崠锛?
// - 浣跨敤妗嗘灦 DistributedLock 鑳藉姏銆?
func AcquireInventoryLock(ctx context.Context, mgr *LockManager, productID, warehouseID uint64, ttl time.Duration) (*Lock, error) {
	key := inventoryLockKey(productID, warehouseID)
	return AcquireLock(ctx, mgr, key, ttl)
}

// inventoryLockKey 鐢熸垚搴撳瓨閿?key
func inventoryLockKey(productID, warehouseID uint64) string {
	return fmt.Sprintf("inventory:%d:%d", productID, warehouseID)
}
