package dlock

import (
	"context"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type exportDLockStub struct{}

func (s *exportDLockStub) Lock(context.Context, string, time.Duration) error                  { return nil }
func (s *exportDLockStub) TryLock(context.Context, string, time.Duration) (bool, error)       { return true, nil }
func (s *exportDLockStub) Unlock(context.Context, string) error                                { return nil }
func (s *exportDLockStub) Renew(context.Context, string, time.Duration) error                  { return nil }
func (s *exportDLockStub) IsLocked(context.Context, string) (bool, error)                      { return true, nil }
func (s *exportDLockStub) WithLock(context.Context, string, time.Duration, func() error) error { return nil }

type exportDLockContainerStub struct {
	lock contract.DistributedLock
}

func (s *exportDLockContainerStub) Bind(string, contract.Factory, bool)                {}
func (s *exportDLockContainerStub) IsBind(string) bool                                 { return true }
func (s *exportDLockContainerStub) MustMake(key string) any                            { v, _ := s.Make(key); return v }
func (s *exportDLockContainerStub) RegisterProvider(contract.ServiceProvider) error     { return nil }
func (s *exportDLockContainerStub) RegisterProviders(...contract.ServiceProvider) error { return nil }
func (s *exportDLockContainerStub) Make(key string) (any, error) {
	if key == contract.DistributedLockKey {
		return s.lock, nil
	}
	return nil, context.DeadlineExceeded
}

func TestExportedDLockHelpers(t *testing.T) {
	stub := &exportDLockStub{}
	containerStub := &exportDLockContainerStub{lock: stub}

	lockSvc, err := Make(containerStub)
	require.NoError(t, err)
	require.Same(t, stub, lockSvc)
	require.Same(t, stub, MustMake(containerStub))

	err = Lock(context.Background(), containerStub, "k", time.Second)
	require.NoError(t, err)
	ok, err := TryLock(context.Background(), containerStub, "k", time.Second)
	require.NoError(t, err)
	require.True(t, ok)
	err = Unlock(context.Background(), containerStub, "k")
	require.NoError(t, err)
	err = WithLock(context.Background(), containerStub, "k", time.Second, func() error { return nil })
	require.NoError(t, err)

	var _ = DistributedLockConfig{Type: "redis"}
}
