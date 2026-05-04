package retry

import (
	"context"

	"github.com/ngq/gorp/framework/container"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Retry = resiliencecontract.Retry
type RetryPolicy = resiliencecontract.RetryPolicy
type RetryConfig = resiliencecontract.RetryConfig

// DefaultRetryPolicy 返回默认重试策略。
func DefaultRetryPolicy() resiliencecontract.RetryPolicy {
	return resiliencecontract.DefaultRetryPolicy()
}

// Make 从容器获取统一重试服务。
func Make(c runtimecontract.Container) (resiliencecontract.Retry, error) {
	return container.MakeRetry(c)
}

// MustMake 从容器获取统一重试服务，失败 panic。
func MustMake(c runtimecontract.Container) resiliencecontract.Retry {
	return container.MustMakeRetry(c)
}

// Do 使用容器中的重试服务执行函数。
func Do(ctx context.Context, c runtimecontract.Container, fn func() error) error {
	retrySvc, err := Make(c)
	if err != nil {
		return err
	}
	return retrySvc.Do(ctx, fn)
}

// DoWithResult 使用容器中的重试服务执行带返回值的函数。
func DoWithResult(ctx context.Context, c runtimecontract.Container, fn func() (any, error)) (any, error) {
	retrySvc, err := Make(c)
	if err != nil {
		return nil, err
	}
	return retrySvc.DoWithResult(ctx, fn)
}

// IsRetryable 判断错误是否可重试。
func IsRetryable(c runtimecontract.Container, err error) (bool, error) {
	retrySvc, makeErr := Make(c)
	if makeErr != nil {
		return false, makeErr
	}
	return retrySvc.IsRetryable(err), nil
}
