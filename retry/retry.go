// Package retry provides a thin facade for the framework's retry capability.
// All public functions accept context.Context instead of runtimecontract.Container,
// resolving the container internally via frameworkcontainer.Resolve(ctx).
// This keeps business code free of container plumbing — just pass the context
// you already have and the retry service is looked up automatically.
//
// retry 包提供框架重试能力的薄门面。
// 所有公开函数接受 context.Context 而非 runtimecontract.Container，
// 内部通过 frameworkcontainer.Resolve(ctx) 解析容器。
// 业务代码无需关心容器细节，传入已有的 context 即可自动查找重试服务。
package retry

import (
	"context"
	"fmt"

	frameworkcontainer "github.com/ngq/gorp/framework/container"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// Retry 是重试契约的顶层别名，方便业务代码直接引用。
type Retry = resiliencecontract.Retry

// RetryPolicy 是重试策略契约的顶层别名，方便业务代码直接引用。
type RetryPolicy = resiliencecontract.RetryPolicy

// RetryConfig 是重试配置契约的顶层别名，方便业务代码直接引用。
type RetryConfig = resiliencecontract.RetryConfig

// DefaultRetryPolicy 返回框架内置的默认重试策略。
func DefaultRetryPolicy() resiliencecontract.RetryPolicy {
	return resiliencecontract.DefaultRetryPolicy()
}

// resolveRetry 从 ctx 解析容器并获取重试服务实例。
// 优先从 ctx 提取容器，提取不到则回退到全局默认容器。
// 如果容器未设置或重试服务未绑定，返回错误。
//
// resolveRetry resolves the retry service from the container extracted via ctx.
// It prefers the container embedded in ctx, falling back to the global default.
// Returns an error if the container is not set or the retry service is not bound.
func resolveRetry(ctx context.Context) (resiliencecontract.Retry, error) {
	cont := frameworkcontainer.Resolve(ctx)
	return frameworkcontainer.MakeWith[resiliencecontract.Retry](cont, resiliencecontract.RetryKey)
}

// GetService 从容器获取统一重试服务。
// 内部通过 ctx 解析容器，再按 RetryKey 查找重试服务实例。
// 如果容器未初始化或重试服务未绑定，返回错误。
//
// GetService returns the unified retry service from the container resolved via ctx.
// Returns an error if the container is not initialized or the retry service is not bound.
func GetService(ctx context.Context) (resiliencecontract.Retry, error) {
	return resolveRetry(ctx)
}

// MustGetService 从容器获取统一重试服务，失败时 panic。
// 适用于启动阶段确认服务已绑定的场景，运行时建议使用 GetService。
//
// MustGetService returns the unified retry service from the container resolved via ctx,
// panicking on failure. Use this in bootstrap code where the service is guaranteed
// to be bound; prefer GetService in runtime paths.
func MustGetService(ctx context.Context) resiliencecontract.Retry {
	svc, err := resolveRetry(ctx)
	if err != nil {
		panic(fmt.Sprintf("retry: failed to resolve retry service: %v", err))
	}
	return svc
}

// Do 使用容器中的重试服务执行函数。
// 内部自动从 ctx 解析容器并获取重试服务，然后调用 retrySvc.Do(ctx, fn)。
// 如果重试服务未绑定，直接返回解析错误。
//
// Do executes a function with the retry service resolved from ctx.
// It automatically resolves the container and retry service, then calls
// retrySvc.Do(ctx, fn). If the retry service is not bound, returns the
// resolution error directly.
//
// 示例:
//
//	err := retry.Do(ctx, func() error {
//	    return callRemote(ctx)
//	})
func Do(ctx context.Context, fn func() error) error {
	retrySvc, err := resolveRetry(ctx)
	if err != nil {
		return err
	}
	return retrySvc.Do(ctx, fn)
}

// DoWithResult 使用容器中的重试服务执行带返回值的函数。
// 内部自动从 ctx 解析容器并获取重试服务，然后调用 retrySvc.DoWithResult(ctx, fn)。
// 如果重试服务未绑定，直接返回解析错误。
//
// DoWithResult executes a function with result using the retry service resolved from ctx.
// It automatically resolves the container and retry service, then calls
// retrySvc.DoWithResult(ctx, fn). If the retry service is not bound, returns the
// resolution error directly.
func DoWithResult(ctx context.Context, fn func() (any, error)) (any, error) {
	retrySvc, err := resolveRetry(ctx)
	if err != nil {
		return nil, err
	}
	return retrySvc.DoWithResult(ctx, fn)
}

// IsRetryable 判断错误是否可重试。
// 内部自动从 ctx 解析容器并获取重试服务，然后调用 retrySvc.IsRetryable(err)。
// 如果重试服务未绑定，返回 false。
//
// IsRetryable reports whether the given error is retryable.
// It resolves the retry service from ctx and calls retrySvc.IsRetryable(err).
// If the retry service is not bound, returns false.
func IsRetryable(ctx context.Context, err error) bool {
	retrySvc, resolveErr := resolveRetry(ctx)
	if resolveErr != nil {
		return false
	}
	return retrySvc.IsRetryable(err)
}
