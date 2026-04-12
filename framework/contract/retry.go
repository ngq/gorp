package contract

import (
	"context"
	"time"
)

// RetryKey 是 Retry 服务在容器中的绑定 key。
const RetryKey = "framework.retry"

// Retry 重试策略接口。
//
// 中文说明：
// - 提供统一的错误重试机制；
// - 支持指数退避策略；
// - 支持可重试错误判断；
// - 可与 CircuitBreaker 组合使用。
type Retry interface {
	// Do 执行带重试的函数。
	//
	// 中文说明：
	// - 根据 RetryPolicy 进行重试；
	// - 使用指数退避延迟；
	// - 最多重试 MaxAttempts 次。
	Do(ctx context.Context, fn func() error) error

	// DoWithResult 执行带重试和返回值的函数。
	//
	// 中文说明：
	// - 支持返回值的重试执行；
	// - 成功时返回结果和 nil；
	// - 失败时返回 nil 和最后一个错误。
	DoWithResult(ctx context.Context, fn func() (any, error)) (any, error)

	// IsRetryable 判断错误是否可重试。
	//
	// 中文说明：
	// - 检查错误类型和错误码；
	// - 默认 502/503/504 可重试；
	// - gRPC Unavailable/DeadlineExceeded 可重试。
	IsRetryable(err error) bool
}

// RetryPolicy 重试策略。
//
// 中文说明：
// - 定义重试次数、延迟、退避策略等；
// - 支持指数退避算法。
type RetryPolicy struct {
	// MaxAttempts 最大重试次数（含首次调用，默认 3）
	MaxAttempts int

	// InitialDelay 初始延迟（默认 100ms）
	InitialDelay time.Duration

	// MaxDelay 最大延迟（默认 1s）
	MaxDelay time.Duration

	// Multiplier 延迟乘数（指数退避，默认 2.0）
	Multiplier float64

	// RetryableErrors 可重试的错误原因列表
	// 如：ErrorReasonInternal、ErrorReasonTimeout
	RetryableErrors []ErrorReason

	// RetryableCodes 可重试的 HTTP 状态码列表
	// 默认：502, 503, 504
	RetryableCodes []int

	// RetryableGRPCCodes 可重试的 gRPC 状态码列表
	// 如：UNAVAILABLE, DEADLINE_EXCEEDED, RESOURCE_EXHAUSTED
	RetryableGRPCCodes []string
}

// DefaultRetryPolicy 返回默认重试策略。
//
// 中文说明：
// - 最大重试 3 次；
// - 初始延迟 100ms；
// - 最大延迟 1s；
// - 乘数 2.0；
// - 可重试状态码：502, 503, 504。
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:        3,
		InitialDelay:       100 * time.Millisecond,
		MaxDelay:           1 * time.Second,
		Multiplier:         2.0,
		RetryableCodes:     []int{502, 503, 504},
		RetryableGRPCCodes: []string{"UNAVAILABLE", "DEADLINE_EXCEEDED", "RESOURCE_EXHAUSTED"},
	}
}

// RetryConfig 重试配置。
//
// 中文说明：
// - 定义 Retry 服务的启用状态和策略；
// - 支持按资源名称配置不同策略。
type RetryConfig struct {
	// Enabled 是否启用重试
	Enabled bool

	// Strategy 策略：exponential（指数退避）/ fixed（固定延迟）
	Strategy string

	// DefaultPolicy 默认重试策略
	DefaultPolicy RetryPolicy

	// ResourcePolicies 按资源名称的策略（可选）
	// key 为资源名称，value 为该资源的策略
	ResourcePolicies map[string]RetryPolicy
}

// GetPolicy 获取指定资源的重试策略。
//
// 中文说明：
// - 如果资源有特定策略则返回；
// - 否则返回默认策略。
func (c *RetryConfig) GetPolicy(resource string) RetryPolicy {
	if policy, ok := c.ResourcePolicies[resource]; ok {
		return policy
	}
	return c.DefaultPolicy
}

// CalculateDelay 计算下一次重试延迟。
//
// 中文说明：
// - 使用指数退避算法；
// - 添加随机抖动避免惊群效应；
// - 延迟不超过 MaxDelay。
func (p *RetryPolicy) CalculateDelay(attempt int, jitter float64) time.Duration {
	// 指数退避：delay = InitialDelay * Multiplier^attempt
	delay := float64(p.InitialDelay)
	for i := 0; i < attempt; i++ {
		delay *= p.Multiplier
	}

	// 不超过最大延迟
	if delay > float64(p.MaxDelay) {
		delay = float64(p.MaxDelay)
	}

	// 添加抖动（0-10%）
	delay += delay * jitter * 0.1

	return time.Duration(delay)
}