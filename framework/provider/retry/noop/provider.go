package noop

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop Retry 实现。
//
// 中文说明：
// - 单体模式下使用，零依赖；
// - 所有操作直接执行，不进行重试；
// - IsRetryable 总是返回 false。
type Provider struct{}

// NewProvider 创建 noop Retry Provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 Provider 名称。
func (p *Provider) Name() string { return "retry.noop" }

// IsDefer 返回是否延迟加载。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回提供的服务 key。
func (p *Provider) Provides() []string {
	return []string{contract.RetryKey}
}

// Register 注册 noop Retry 服务。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.RetryKey, func(c contract.Container) (any, error) {
		return &noopRetry{}, nil
	}, true)

	return nil
}

// Boot 启动 Provider。
func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// noopRetry noop Retry 实现。
//
// 中文说明：
// - Do 直接执行函数，不重试；
// - IsRetryable 总是返回 false。
type noopRetry struct{}

// Do 直接执行函数，不重试。
//
// 中文说明：
// - 单体模式下不需要重试；
// - 直接执行函数并返回结果。
func (r *noopRetry) Do(ctx context.Context, fn func() error) error {
	return fn()
}

// DoWithResult 直接执行函数，不重试。
func (r *noopRetry) DoWithResult(ctx context.Context, fn func() (any, error)) (any, error) {
	return fn()
}

// IsRetryable 总是返回 false。
//
// 中文说明：
// - noop 模式下不进行重试；
// - 所有错误都视为不可重试。
func (r *noopRetry) IsRetryable(err error) bool {
	return false
}