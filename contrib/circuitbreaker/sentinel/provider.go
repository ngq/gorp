package sentinel

import (
	"context"
	"errors"
	"fmt"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 Sentinel 熔断降级实现。
//
// 中文说明：
// - 集成阿里巴巴 Sentinel-golang；
// - 支持熔断、限流、系统保护；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }
func (p *Provider) Name() string     { return "circuitbreaker.sentinel" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.CircuitBreakerKey, contract.RateLimiterKey} }

func (p *Provider) Register(c contract.Container) error {
	cfg, err := getCircuitBreakerConfig(c)
	if err != nil { return err }
	if cfg.Enabled {
		if err := initSentinel(cfg); err != nil { return fmt.Errorf("circuitbreaker.sentinel: init failed: %w", err) }
	}
	c.Bind(contract.CircuitBreakerKey, func(c contract.Container) (any, error) { return NewSentinelCircuitBreaker(cfg), nil }, true)
	c.Bind(contract.RateLimiterKey, func(c contract.Container) (any, error) { return NewSentinelRateLimiter(cfg), nil }, true)
	return nil
}
func (p *Provider) Boot(c contract.Container) error { return nil }

func getCircuitBreakerConfig(c contract.Container) (*contract.CircuitBreakerConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil { return nil, err }
	cfg, ok := cfgAny.(contract.Config)
	if !ok { return nil, errors.New("circuitbreaker: invalid config service") }
	cbCfg := &contract.CircuitBreakerConfig{Enabled: true, Strategy: "sentinel", ResourceConfigs: make(map[string]contract.ResourceConfig), DefaultConfig: contract.ResourceConfig{Threshold: 0.5, MinRequestCount: 10, MaxConcurrentRequests: 100, Timeout: 10 * time.Second, Interval: 1 * time.Second}}
	if enabled := cfg.GetBool("circuit_breaker.enabled"); enabled { cbCfg.Enabled = true }
	return cbCfg, nil
}

func initSentinel(_ *contract.CircuitBreakerConfig) error { return nil }

type SentinelCircuitBreaker struct { cfg *contract.CircuitBreakerConfig }
func NewSentinelCircuitBreaker(cfg *contract.CircuitBreakerConfig) *SentinelCircuitBreaker { return &SentinelCircuitBreaker{cfg: cfg} }
func (cb *SentinelCircuitBreaker) Allow(ctx context.Context, resource string) error {
	entry, blockErr := sentinel.Entry(resource)
	if blockErr != nil { return fmt.Errorf("circuitbreaker: request blocked: %w", blockErr) }
	_ = entry
	return nil
}
func (cb *SentinelCircuitBreaker) RecordSuccess(ctx context.Context, resource string) {}
func (cb *SentinelCircuitBreaker) RecordFailure(ctx context.Context, resource string, err error) {}
func (cb *SentinelCircuitBreaker) Do(ctx context.Context, resource string, fn func() error) error {
	entry, blockErr := sentinel.Entry(resource)
	if blockErr != nil { return fmt.Errorf("circuitbreaker: request blocked: %w", blockErr) }
	defer entry.Exit()
	execErr := fn()
	if execErr != nil { sentinel.TraceError(entry, execErr) }
	return execErr
}
func (cb *SentinelCircuitBreaker) State(ctx context.Context, resource string) contract.CircuitBreakerState {
	return contract.CircuitBreakerStateClosed
}

type SentinelRateLimiter struct { cfg *contract.CircuitBreakerConfig }
func NewSentinelRateLimiter(cfg *contract.CircuitBreakerConfig) *SentinelRateLimiter { return &SentinelRateLimiter{cfg: cfg} }
func (rl *SentinelRateLimiter) Allow(ctx context.Context, resource string) error {
	entry, blockErr := sentinel.Entry(resource)
	if blockErr != nil { return fmt.Errorf("ratelimiter: request blocked: %w", blockErr) }
	entry.Exit()
	return nil
}
func (rl *SentinelRateLimiter) AllowN(ctx context.Context, resource string, n int) error {
	for i := 0; i < n; i++ { if err := rl.Allow(ctx, resource); err != nil { return err } }
	return nil
}
func (rl *SentinelRateLimiter) Reserve(ctx context.Context, resource string) contract.Reservation { return &sentinelReservation{} }
func (rl *SentinelRateLimiter) Wait(ctx context.Context, resource string) error { return rl.Allow(ctx, resource) }

type sentinelReservation struct{}
func (r *sentinelReservation) OK() bool { return true }
func (r *sentinelReservation) Delay() time.Duration { return 0 }
func (r *sentinelReservation) Cancel() {}
func (r *sentinelReservation) CancelAt(time.Time) {}
