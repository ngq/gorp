package dlock

import (
	"context"
	"io"
	"testing"
	"time"

	frameworkcontainer "github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

// exportDLockStub 实现分布式锁契约，用于测试。
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

// exportDLockContainerStub 实现容器契约，用于注入测试用分布式锁服务。
type exportDLockContainerStub struct {
	lock datacontract.DistributedLock
}

func (s *exportDLockContainerStub) Bind(string, runtimecontract.Factory, bool)              {}
func (s *exportDLockContainerStub) NamedBind(string, string, runtimecontract.Factory, bool) {}
func (s *exportDLockContainerStub) IsBind(string) bool                                      { return true }
func (s *exportDLockContainerStub) IsBindNamed(string, string) bool                         { return false }
func (s *exportDLockContainerStub) MustMake(key string) any                                 { v, _ := s.Make(key); return v }
func (s *exportDLockContainerStub) MustMakeNamed(string, string) any                        { return nil }
func (s *exportDLockContainerStub) MakeNamed(string, string) (any, error)                   { return nil, nil }
func (s *exportDLockContainerStub) RegisterProvider(runtimecontract.ServiceProvider) error   { return nil }
func (s *exportDLockContainerStub) RegisterProviders(...runtimecontract.ServiceProvider) error {
	return nil
}

func (s *exportDLockContainerStub) Make(key string) (any, error) {
	if key == datacontract.DistributedLockKey {
		return s.lock, nil
	}
	return nil, frameworkcontainer.ErrDefaultContainerNotSet
}

func (s *exportDLockContainerStub) RegisterCloser(string, io.Closer)                    {}
func (s *exportDLockContainerStub) Destroy() error                                      { return nil }
func (s *exportDLockContainerStub) RegisteredProviders() []runtimecontract.ProviderInfo  { return nil }
func (s *exportDLockContainerStub) DebugPrint() string                                  { return "" }
func (s *exportDLockContainerStub) ProviderDAG() runtimecontract.ProviderDAG {
	return runtimecontract.ProviderDAG{}
}

// setupTestContainer 注入测试容器到 frameworkcontainer 全局默认，
// 并在测试结束后清理，避免污染其他测试。
func setupTestContainer(t *testing.T, lock datacontract.DistributedLock) {
	t.Helper()
	stub := &exportDLockContainerStub{lock: lock}
	frameworkcontainer.SetDefault(stub)
	t.Cleanup(func() {
		frameworkcontainer.SetDefault(nil)
	})
}

func TestGetService(t *testing.T) {
	stub := &exportDLockStub{}
	setupTestContainer(t, stub)

	ctx := context.Background()

	// GetService 正常返回
	lockSvc, err := GetService(ctx)
	require.NoError(t, err)
	require.Same(t, stub, lockSvc)

	// MustGetService 正常返回
	require.Same(t, stub, MustGetService(ctx))
}

func TestLockUnlock(t *testing.T) {
	stub := &exportDLockStub{}
	setupTestContainer(t, stub)

	ctx := context.Background()

	err := Lock(ctx, "k", time.Second)
	require.NoError(t, err)

	ok, err := TryLock(ctx, "k", time.Second)
	require.NoError(t, err)
	require.True(t, ok)

	err = Unlock(ctx, "k")
	require.NoError(t, err)
}

func TestWithLock(t *testing.T) {
	stub := &exportDLockStub{}
	setupTestContainer(t, stub)

	ctx := context.Background()

	err := WithLock(ctx, "k", time.Second, func() error { return nil })
	require.NoError(t, err)
}

func TestGetServiceNotRegistered(t *testing.T) {
	// 不设置默认容器，resolveDLock 应返回错误
	frameworkcontainer.SetDefault(nil)
	t.Cleanup(func() {
		frameworkcontainer.SetDefault(nil)
	})

	ctx := context.Background()

	_, err := GetService(ctx)
	require.Error(t, err)
}

func TestDistributedLockConfigAlias(t *testing.T) {
	// 验证 DistributedLockConfig 类型别名可用
	var _ = DistributedLockConfig{Type: "redis"}
}
