// Application scenarios:
// - Define the retry contract shared by middleware, providers, and service code.
// - Standardize retry execution, retryability checks, policy lookup, and delay calculation.
// - Provide one reusable policy/config model for HTTP, gRPC, and generic operation retries.
//
// 适用场景：
// - 定义中间件、provider 和 service 代码共享的重试契约。
// - 统一重试执行、可重试判断、策略查找和退避延迟计算语义。
// - 为 HTTP、gRPC 和通用操作重试提供复用型策略/配置模型。
package resilience

import (
	"context"
	"time"
)

// RetryKey is the container key for the retry capability.
//
// RetryKey 是重试能力的容器键。
const RetryKey = "framework.retry"

// Retry defines the retry execution contract.
//
// Retry 定义重试执行契约。
type Retry interface {
	Do(ctx context.Context, fn func() error) error
	DoForResource(ctx context.Context, resource string, fn func() error) error
	DoWithResult(ctx context.Context, fn func() (any, error)) (any, error)
	IsRetryable(err error) bool
}

// RetryPolicy describes one retry strategy.
//
// RetryPolicy 描述一条重试策略。
type RetryPolicy struct {
	MaxAttempts        int
	InitialDelay       time.Duration
	MaxDelay           time.Duration
	Multiplier         float64
	RetryableErrors    []ErrorReason
	RetryableCodes     []int
	RetryableGRPCCodes []string
}

// DefaultRetryPolicy returns the framework default retry policy.
//
// DefaultRetryPolicy 返回框架默认重试策略。
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

// RetryConfig describes retry-related runtime configuration.
//
// RetryConfig 描述重试相关运行时配置。
type RetryConfig struct {
	Enabled          bool
	Strategy         string
	DefaultPolicy    RetryPolicy
	ResourcePolicies map[string]RetryPolicy
}

// GetPolicy returns the policy for one resource, falling back to the default policy.
//
// GetPolicy 返回指定资源的策略，未命中时回退到默认策略。
func (c *RetryConfig) GetPolicy(resource string) RetryPolicy {
	if policy, ok := c.ResourcePolicies[resource]; ok {
		return policy
	}
	return c.DefaultPolicy
}

// CalculateDelay calculates the delay for one retry attempt with jitter applied.
//
// CalculateDelay 计算某次重试的退避延迟，并附加抖动。
func (p *RetryPolicy) CalculateDelay(attempt int, jitter float64) time.Duration {
	delay := float64(p.InitialDelay)
	for i := 0; i < attempt; i++ {
		delay *= p.Multiplier
	}
	if delay > float64(p.MaxDelay) {
		// Cap the exponential growth at MaxDelay so policies stay bounded.
		// 将指数增长限制在 MaxDelay 以内，避免策略失控膨胀。
		delay = float64(p.MaxDelay)
	}
	delay += delay * jitter * 0.1
	return time.Duration(delay)
}
