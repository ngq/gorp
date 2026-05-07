// Application scenarios:
// - Provide the monolith-friendly no-op retry capability.
// - Keep retry wiring valid even when production retry policies are disabled.
// - Offer one contract-complete fallback for bootstrap and tests.
//
// 适用场景：
// - 为单体或未启用治理的场景提供空实现 Retry 能力。
// - 在未开启生产级重试策略时，仍然保持容器装配与调用链完整。
// - 为 bootstrap 与测试提供一套契约完整的兜底实现。
package noop

import (
	"context"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers the no-op retry implementation.
//
// Provider 注册空实现 Retry provider。
type Provider struct{}

// NewProvider creates a no-op retry provider.
//
// NewProvider 创建空实现 Retry provider。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name.
//
// Name 返回 provider 名称。
func (p *Provider) Name() string { return "retry.noop" }

// IsDefer reports that the provider can be lazily loaded.
//
// IsDefer 返回该 provider 可延迟加载。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the retry capability key exposed by this provider.
//
// Provides 返回该 provider 暴露的 Retry 能力 key。
func (p *Provider) Provides() []string {
	return []string{resiliencecontract.RetryKey}
}

// Register binds the no-op retry service into the container.
//
// Register 将空实现 Retry 服务注册到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(resiliencecontract.RetryKey, func(c runtimecontract.Container) (any, error) {
		return &noopRetry{}, nil
	}, true)

	return nil
}

// Boot does not need extra runtime work for the no-op provider.
//
// Boot 对空实现 provider 不需要额外启动逻辑。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// noopRetry executes the operation once and never retries.
//
// noopRetry 只执行一次操作，不做任何重试。
type noopRetry struct{}

// Do runs the operation once.
//
// Do 执行一次操作并直接返回结果。
func (r *noopRetry) Do(ctx context.Context, fn func() error) error {
	return fn()
}

// DoForResource runs the operation once without resource-specific retry behavior.
//
// DoForResource 按资源执行一次操作，但不应用资源级重试行为。
func (r *noopRetry) DoForResource(ctx context.Context, resource string, fn func() error) error {
	return fn()
}

// DoWithResult runs the operation once and returns its result.
//
// DoWithResult 执行一次操作并返回其结果。
func (r *noopRetry) DoWithResult(ctx context.Context, fn func() (any, error)) (any, error) {
	return fn()
}

// IsRetryable always returns false for the no-op implementation.
//
// IsRetryable 对空实现始终返回 false。
func (r *noopRetry) IsRetryable(err error) bool {
	return false
}
