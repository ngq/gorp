package noop

import (
	"context"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop 熔断器实现。
//
// 中文说明：
// - 单体项目默认使用此 provider；
// - 所有请求都允许通过；
// - 不进行实际的熔断控制。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "circuitbreaker.noop" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.CircuitBreakerKey, contract.RateLimiterKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.CircuitBreakerKey, func(c contract.Container) (any, error) {
		return &noopCircuitBreaker{}, nil
	}, true)

	c.Bind(contract.RateLimiterKey, func(c contract.Container) (any, error) {
		return &noopRateLimiter{}, nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// noopCircuitBreaker 是 CircuitBreaker 的空实现。
type noopCircuitBreaker struct{}

// Allow 检查是否允许请求（总是允许）。
func (cb *noopCircuitBreaker) Allow(ctx context.Context, resource string) error {
	return nil
}

// RecordSuccess 记录成功请求（空操作）。
func (cb *noopCircuitBreaker) RecordSuccess(ctx context.Context, resource string) {}

// RecordFailure 记录失败请求（空操作）。
func (cb *noopCircuitBreaker) RecordFailure(ctx context.Context, resource string, err error) {}

// Do 执行受熔断器保护的函数（直接执行）。
func (cb *noopCircuitBreaker) Do(ctx context.Context, resource string, fn func() error) error {
	return fn()
}

// State 获取熔断器状态（总是返回关闭状态）。
func (cb *noopCircuitBreaker) State(ctx context.Context, resource string) contract.CircuitBreakerState {
	return contract.CircuitBreakerStateClosed
}

// noopRateLimiter 是 RateLimiter 的空实现。
type noopRateLimiter struct{}

// Allow 检查是否允许请求（总是允许）。
func (rl *noopRateLimiter) Allow(ctx context.Context, resource string) error {
	return nil
}

// AllowN 检查是否允许 N 个请求（总是允许）。
func (rl *noopRateLimiter) AllowN(ctx context.Context, resource string, n int) error {
	return nil
}

// Reserve 预留令牌（立即成功）。
func (rl *noopRateLimiter) Reserve(ctx context.Context, resource string) contract.Reservation {
	return &noopReservation{}
}

// Wait 等待直到获取令牌（立即返回）。
func (rl *noopRateLimiter) Wait(ctx context.Context, resource string) error {
	return nil
}

// WaitTimeout 等待直到获取令牌或超时（立即返回）。
func (rl *noopRateLimiter) WaitTimeout(ctx context.Context, resource string, timeout time.Duration) error {
	return nil
}

// noopReservation 是 Reservation 的空实现。
type noopReservation struct{}

func (r *noopReservation) OK() bool                { return true }
func (r *noopReservation) Delay() time.Duration     { return 0 }
func (r *noopReservation) Cancel()                  {}
func (r *noopReservation) CancelAt(t time.Time)     {}