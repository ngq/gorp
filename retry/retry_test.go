package retry

import (
	"context"
	"errors"
	"io"
	"testing"

	frameworkcontainer "github.com/ngq/gorp/framework/container"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

// retryStub 是重试服务的测试桩实现，用于验证 facade 函数的正确调用转发。
type retryStub struct{}

func (s *retryStub) Do(_ context.Context, fn func() error) error { return fn() }
func (s *retryStub) DoForResource(_ context.Context, _ string, fn func() error) error {
	return fn()
}
func (s *retryStub) DoWithResult(_ context.Context, fn func() (any, error)) (any, error) {
	return fn()
}
func (s *retryStub) IsRetryable(err error) bool { return err != nil }

// containerStub 是最小化的容器桩，仅支持按 RetryKey 返回重试服务实例。
// 其他 Make 调用返回 ErrDefaultContainerNotSet，确保测试不会意外命中其他绑定。
type containerStub struct {
	retry resiliencecontract.Retry
}

func (s *containerStub) Make(key string) (any, error) {
	if key == resiliencecontract.RetryKey {
		return s.retry, nil
	}
	return nil, frameworkcontainer.ErrDefaultContainerNotSet
}

// 以下为满足 runtimecontract.Container 接口的空实现

func (s *containerStub) Bind(string, runtimecontract.Factory, bool)                {}
func (s *containerStub) NamedBind(string, string, runtimecontract.Factory, bool)   {}
func (s *containerStub) IsBind(string) bool                                        { return true }
func (s *containerStub) IsBindNamed(string, string) bool                           { return false }
func (s *containerStub) MustMake(key string) any                                   { v, _ := s.Make(key); return v }
func (s *containerStub) MustMakeNamed(string, string) any                          { return nil }
func (s *containerStub) MakeNamed(string, string) (any, error)                     { return nil, nil }
func (s *containerStub) RegisterProvider(runtimecontract.ServiceProvider) error     { return nil }
func (s *containerStub) RegisterProviders(...runtimecontract.ServiceProvider) error { return nil }
func (s *containerStub) RegisterCloser(string, io.Closer)                          {}
func (s *containerStub) Destroy() error                                            { return nil }
func (s *containerStub) RegisteredProviders() []runtimecontract.ProviderInfo        { return nil }
func (s *containerStub) DebugPrint() string                                        { return "" }
func (s *containerStub) ProviderDAG() runtimecontract.ProviderDAG {
	return runtimecontract.ProviderDAG{}
}

// setupStubContainer 创建容器桩并注入为全局默认容器，测试结束后自动清理。
func setupStubContainer(t *testing.T) *retryStub {
	t.Helper()
	stub := &retryStub{}
	frameworkcontainer.SetDefault(&containerStub{retry: stub})
	t.Cleanup(func() {
		frameworkcontainer.SetDefault(nil)
	})
	return stub
}

func TestGetService(t *testing.T) {
	stub := setupStubContainer(t)
	ctx := context.Background()

	svc, err := GetService(ctx)
	require.NoError(t, err)
	require.Same(t, stub, svc)
}

func TestMustGetService(t *testing.T) {
	stub := setupStubContainer(t)
	ctx := context.Background()

	svc := MustGetService(ctx)
	require.Same(t, stub, svc)
}

func TestMustGetService_Panics(t *testing.T) {
	// 不设置默认容器，MustGetService 应该 panic
	ctx := context.Background()
	require.Panics(t, func() {
		MustGetService(ctx)
	})
}

func TestDo(t *testing.T) {
	setupStubContainer(t)
	ctx := context.Background()

	called := false
	err := Do(ctx, func() error {
		called = true
		return nil
	})
	require.NoError(t, err)
	require.True(t, called)
}

func TestDo_WithError(t *testing.T) {
	setupStubContainer(t)
	ctx := context.Background()

	expectedErr := errors.New("boom")
	err := Do(ctx, func() error {
		return expectedErr
	})
	require.ErrorIs(t, err, expectedErr)
}

func TestDoWithResult(t *testing.T) {
	setupStubContainer(t)
	ctx := context.Background()

	result, err := DoWithResult(ctx, func() (any, error) {
		return 42, nil
	})
	require.NoError(t, err)
	require.Equal(t, 42, result)
}

func TestDoWithResult_WithError(t *testing.T) {
	setupStubContainer(t)
	ctx := context.Background()

	expectedErr := errors.New("fail")
	result, err := DoWithResult(ctx, func() (any, error) {
		return nil, expectedErr
	})
	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, result)
}

func TestIsRetryable_True(t *testing.T) {
	setupStubContainer(t)
	ctx := context.Background()

	ok := IsRetryable(ctx, errors.New("transient"))
	require.True(t, ok)
}

func TestIsRetryable_False(t *testing.T) {
	setupStubContainer(t)
	ctx := context.Background()

	ok := IsRetryable(ctx, nil)
	require.False(t, ok)
}

func TestIsRetryable_ServiceNotBound(t *testing.T) {
	// 不设置默认容器，重试服务不可用时应返回 false
	ctx := context.Background()
	ok := IsRetryable(ctx, errors.New("boom"))
	require.False(t, ok)
}

func TestDefaultRetryPolicy(t *testing.T) {
	require.Equal(t, resiliencecontract.DefaultRetryPolicy(), DefaultRetryPolicy())
}
