// Package noop_test provides unit tests for the distributed lock noop provider.
//
// 适用场景：
// - 验证分布式锁 noop provider 的注册与空操作行为。
package noop

import (
	"context"
	"testing"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/assert"
)

func TestNoopLock(t *testing.T) {
	lock := &noopLock{}

	// 测试 Lock
	err := lock.Lock(context.Background(), "test-key", 10*time.Second)
	assert.NoError(t, err)

	// 测试 Unlock
	err = lock.Unlock(context.Background(), "test-key")
	assert.NoError(t, err)

	// 测试 TryLock
	ok, err := lock.TryLock(context.Background(), "test-key2", 10*time.Second)
	assert.NoError(t, err)
	assert.True(t, ok)

	// 测试 WithLock
	executed := false
	err = lock.WithLock(context.Background(), "test-key3", 10*time.Second, func() error {
		executed = true
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestProvider(t *testing.T) {
	p := NewProvider()

	assert.Equal(t, "dlock.noop", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{datacontract.DistributedLockKey}, p.Provides())
}
