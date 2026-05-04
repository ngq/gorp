package resilience

import (
	"context"
	"time"
)

const RetryKey = "framework.retry"

type Retry interface {
	Do(ctx context.Context, fn func() error) error
	DoWithResult(ctx context.Context, fn func() (any, error)) (any, error)
	IsRetryable(err error) bool
}

type RetryPolicy struct {
	MaxAttempts        int
	InitialDelay       time.Duration
	MaxDelay           time.Duration
	Multiplier         float64
	RetryableErrors    []ErrorReason
	RetryableCodes     []int
	RetryableGRPCCodes []string
}

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

type RetryConfig struct {
	Enabled          bool
	Strategy         string
	DefaultPolicy    RetryPolicy
	ResourcePolicies map[string]RetryPolicy
}

func (c *RetryConfig) GetPolicy(resource string) RetryPolicy {
	if policy, ok := c.ResourcePolicies[resource]; ok {
		return policy
	}
	return c.DefaultPolicy
}

func (p *RetryPolicy) CalculateDelay(attempt int, jitter float64) time.Duration {
	delay := float64(p.InitialDelay)
	for i := 0; i < attempt; i++ {
		delay *= p.Multiplier
	}
	if delay > float64(p.MaxDelay) {
		delay = float64(p.MaxDelay)
	}
	delay += delay * jitter * 0.1
	return time.Duration(delay)
}
