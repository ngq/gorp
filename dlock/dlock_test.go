package dlock

import (
	"context"
	"io"
	"testing"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

type exportDLockStub struct{}

func (s *exportDLockStub) Lock(context.Context, string, time.Duration) error { return nil }
func (s *exportDLockStub) TryLock(context.Context, string, time.Duration) (bool, error) {
	return true, nil
}
func (s *exportDLockStub) Unlock(context.Context, string) error               { return nil }
func (s *exportDLockStub) Renew(context.Context, string, time.Duration) error { return nil }
func (s *exportDLockStub) IsLocked(context.Context, string) (bool, error)     { return true, nil }
func (s *exportDLockStub) WithLock(context.Context, string, time.Duration, func() error) error {
	return nil
}

type exportDLockContainerStub struct {
	lock datacontract.DistributedLock
}

func (s *exportDLockContainerStub) Bind(string, runtimecontract.Factory, bool)                      {}
func (s *exportDLockContainerStub) NamedBind(string, string, runtimecontract.Factory, bool)          {}
func (s *exportDLockContainerStub) IsBind(string) bool                                               { return true }
func (s *exportDLockContainerStub) IsBindNamed(string, string) bool                                  { return false }
func (s *exportDLockContainerStub) MustMake(key string) any                                          { v, _ := s.Make(key); return v }
func (s *exportDLockContainerStub) MustMakeNamed(string, string) any                                 { return nil }
func (s *exportDLockContainerStub) RegisterCloser(string, io.Closer)                                 {}
func (s *exportDLockContainerStub) Destroy() error                                                   { return nil }
func (s *exportDLockContainerStub) RegisteredProviders() []runtimecontract.ProviderInfo              { return nil }
func (s *exportDLockContainerStub) DebugPrint() string                                               { return "" }
func (s *exportDLockContainerStub) ProviderDAG() runtimecontract.ProviderDAG                          { return runtimecontract.ProviderDAG{} }
func (s *exportDLockContainerStub) MakeNamed(string, string) (any, error)                            { return nil, nil }
func (s *exportDLockContainerStub) RegisterProvider(runtimecontract.ServiceProvider) error {
	return nil
}
func (s *exportDLockContainerStub) RegisterProviders(...runtimecontract.ServiceProvider) error {
	return nil
}
func (s *exportDLockContainerStub) Make(key string) (any, error) {
	if key == datacontract.DistributedLockKey {
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
