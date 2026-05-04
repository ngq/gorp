package resilience

import (
	"context"
	"time"
)

type RateLimiter interface {
	Allow(ctx context.Context, resource string) error
	AllowN(ctx context.Context, resource string, n int) error
	Reserve(ctx context.Context, resource string) Reservation
	Wait(ctx context.Context, resource string) error
	WaitTimeout(ctx context.Context, resource string, timeout time.Duration) error
}

type Reservation interface {
	OK() bool
	Delay() time.Duration
	Cancel()
	CancelAt(t time.Time)
}

type RateLimiterConfig struct {
	Enabled         bool
	Strategy        string
	ResourceConfigs map[string]RateResourceConfig
	DefaultConfig   RateResourceConfig
}

type RateResourceConfig struct {
	QPS     float64
	Burst   int
	MaxWait time.Duration
}
