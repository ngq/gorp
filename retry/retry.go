package retry

import (
	"context"

	"github.com/ngq/gorp/framework/container"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Retry is the top-level alias of the retry contract.
// Retry 是重试契约的顶层别名。
type Retry = resiliencecontract.Retry

// RetryPolicy is the top-level alias of the retry policy contract.
// RetryPolicy 是重试策略契约的顶层别名。
type RetryPolicy = resiliencecontract.RetryPolicy

// RetryConfig is the top-level alias of the retry config contract.
// RetryConfig 是重试配置契约的顶层别名。
type RetryConfig = resiliencecontract.RetryConfig

// DefaultRetryPolicy returns the built-in default retry policy.
// DefaultRetryPolicy 返回默认重试策略。
func DefaultRetryPolicy() resiliencecontract.RetryPolicy {
	return resiliencecontract.DefaultRetryPolicy()
}

// Get returns the unified retry service from the container.
// Get 从容器获取统一重试服务。
func Get(c runtimecontract.Container) (resiliencecontract.Retry, error) {
	return container.MakeRetry(c)
}

// GetOrPanic returns the unified retry service from the container and panics on failure.
// GetOrPanic 从容器获取统一重试服务，失败 panic。
func GetOrPanic(c runtimecontract.Container) resiliencecontract.Retry {
	return container.MustMakeRetry(c)
}

// Do executes a function with the retry service from the container.
// Do 使用容器中的重试服务执行函数。
//
// Example:
//
//	err := retry.Do(ctx, c, func() error {
//	    return callRemote(ctx)
//	})
func Do(ctx context.Context, c runtimecontract.Container, fn func() error) error {
	retrySvc, err := Get(c)
	if err != nil {
		return err
	}
	return retrySvc.Do(ctx, fn)
}

// DoWithResult executes a function with result using the retry service from the container.
// DoWithResult 使用容器中的重试服务执行带返回值的函数。
func DoWithResult(ctx context.Context, c runtimecontract.Container, fn func() (any, error)) (any, error) {
	retrySvc, err := Get(c)
	if err != nil {
		return nil, err
	}
	return retrySvc.DoWithResult(ctx, fn)
}

// IsRetryable reports whether the given error is retryable.
// IsRetryable 判断错误是否可重试。
func IsRetryable(c runtimecontract.Container, err error) (bool, error) {
	retrySvc, makeErr := Get(c)
	if makeErr != nil {
		return false, makeErr
	}
	return retrySvc.IsRetryable(err), nil
}
