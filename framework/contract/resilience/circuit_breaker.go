package resilience

import (
	"context"
	"time"
)

const (
	CircuitBreakerKey = "framework.circuit_breaker"
	RateLimiterKey    = "framework.rate_limiter"
)

type CircuitBreaker interface {
	Allow(ctx context.Context, resource string) error
	RecordSuccess(ctx context.Context, resource string)
	RecordFailure(ctx context.Context, resource string, err error)
	Do(ctx context.Context, resource string, fn func() error) error
	State(ctx context.Context, resource string) CircuitBreakerState
}

type CircuitBreakerState int

const (
	CircuitBreakerStateClosed CircuitBreakerState = iota
	CircuitBreakerStateOpen
	CircuitBreakerStateHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitBreakerStateClosed:
		return "closed"
	case CircuitBreakerStateOpen:
		return "open"
	case CircuitBreakerStateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

type CircuitBreakerConfig struct {
	Enabled         bool
	Strategy        string
	ResourceConfigs map[string]ResourceConfig
	DefaultConfig   ResourceConfig
}

type ResourceConfig struct {
	Threshold             float64
	MinRequestCount       int64
	MaxConcurrentRequests int64
	Timeout               time.Duration
	RetryTimeoutMs        int64
	Interval              time.Duration
}
