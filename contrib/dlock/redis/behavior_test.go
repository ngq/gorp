package redis

import (
	"context"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestGenerateTokenProducesDistinctValues(t *testing.T) {
	a := generateToken()
	b := generateToken()
	require.NotEmpty(t, a)
	require.NotEmpty(t, b)
	require.NotEqual(t, a, b)
}

func TestLockUnlockRenewRequireHeldLock(t *testing.T) {
	lock := &Lock{cfg: &contract.DistributedLockConfig{KeyPrefix: "lock:"}}
	_, err := lock.IsLocked(context.Background(), "demo")
	require.Error(t, err)

	err = lock.Unlock(context.Background(), "demo")
	require.Error(t, err)
	require.Contains(t, err.Error(), "lock not held")

	err = lock.Renew(context.Background(), "demo", time.Second)
	require.Error(t, err)
	require.Contains(t, err.Error(), "lock not held")
}

func TestWithLockReturnsAcquireError(t *testing.T) {
	lock := &Lock{cfg: &contract.DistributedLockConfig{KeyPrefix: "lock:"}}
	err := lock.WithLock(context.Background(), "demo", time.Second, func() error { return nil })
	require.Error(t, err)
}

func TestStopWatchdogOnUnknownKeyIsNoop(t *testing.T) {
	lock := &Lock{}
	lock.stopWatchdog("missing")
}
