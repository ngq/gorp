package sentinel

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	sentinelcb "github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/isolation"
	internalnative "github.com/ngq/gorp/contrib/internal/native"
	"github.com/ngq/gorp/framework/contract"
)

var sentinelEntry = sentinel.Entry
var sentinelInitDefault = sentinel.InitDefault

// Provider 提供 Sentinel 熔断降级实现。
//
// 中文说明：
//   - 集成 Sentinel-golang；
//   - 支持熔断、限流、系统保护；
//   - 当前处于最小可验证治理闭环。
//   - 当前状态：部分可用
//   - 说明：已完成 P1 最小治理闭环，具备规则加载、状态记录与关键行为测试；
//     但规则来源与完整治理产品化仍未完成，当前不能按完整治理后端对外宣传。
type Provider struct{}

func NewProvider() *Provider      { return &Provider{} }
func (p *Provider) Name() string  { return "circuitbreaker.sentinel" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{contract.CircuitBreakerKey, contract.RateLimiterKey}
}

func (p *Provider) Register(c contract.Container) error {
	cfg, err := getCircuitBreakerConfig(c)
	if err != nil {
		return err
	}
	if cfg.Enabled {
		if err := initSentinel(cfg); err != nil {
			return fmt.Errorf("circuitbreaker.sentinel: init failed: %w", err)
		}
	}
	c.Bind(contract.CircuitBreakerKey, func(c contract.Container) (any, error) {
		return NewSentinelCircuitBreaker(cfg), nil
	}, true)
	c.Bind(contract.RateLimiterKey, func(c contract.Container) (any, error) {
		return NewSentinelRateLimiter(cfg), nil
	}, true)
	return nil
}
func (p *Provider) Boot(c contract.Container) error { return nil }

func getCircuitBreakerConfig(c contract.Container) (*contract.CircuitBreakerConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("circuitbreaker: invalid config service")
	}

	enabled := true
	if cfg.Get("circuit_breaker.enabled") != nil {
		enabled = cfg.GetBool("circuit_breaker.enabled")
	}

	cbCfg := &contract.CircuitBreakerConfig{
		Enabled:         enabled,
		Strategy:        "sentinel",
		ResourceConfigs: make(map[string]contract.ResourceConfig),
		DefaultConfig: contract.ResourceConfig{
			Threshold:             0.5,
			MinRequestCount:       10,
			MaxConcurrentRequests: 100,
			Timeout:               10 * time.Second,
			Interval:              1 * time.Second,
		},
	}
	return cbCfg, nil
}

func initSentinel(cfg *contract.CircuitBreakerConfig) error {
	if err := sentinelInitDefault(); err != nil {
		return err
	}

	isolationRules := make([]*isolation.Rule, 0, len(cfg.ResourceConfigs))
	breakerRules := make([]*sentinelcb.Rule, 0, len(cfg.ResourceConfigs))
	for resource, ruleCfg := range cfg.ResourceConfigs {
		if ruleCfg.MaxConcurrentRequests > 0 {
			isolationRules = append(isolationRules, &isolation.Rule{
				Resource:   resource,
				MetricType: isolation.Concurrency,
				Threshold:  uint32(ruleCfg.MaxConcurrentRequests),
			})
		}
		if breakerRule := buildCircuitBreakerRule(resource, ruleCfg, cfg.DefaultConfig); breakerRule != nil {
			breakerRules = append(breakerRules, breakerRule)
		}
	}

	if len(isolationRules) > 0 {
		if _, err := isolation.LoadRules(isolationRules); err != nil {
			return fmt.Errorf("load isolation rules failed: %w", err)
		}
	}
	if len(breakerRules) > 0 {
		if _, err := sentinelcb.LoadRules(breakerRules); err != nil {
			return fmt.Errorf("load circuitbreaker rules failed: %w", err)
		}
	}
	return nil
}

func buildCircuitBreakerRule(resource string, current contract.ResourceConfig, defaults contract.ResourceConfig) *sentinelcb.Rule {
	minRequest := current.MinRequestCount
	if minRequest <= 0 {
		minRequest = defaults.MinRequestCount
	}
	interval := current.Interval
	if interval <= 0 {
		interval = defaults.Interval
	}
	retryTimeout := current.RetryTimeoutMs
	if retryTimeout <= 0 && current.Timeout > 0 {
		retryTimeout = current.Timeout.Milliseconds()
	}
	if retryTimeout <= 0 && defaults.Timeout > 0 {
		retryTimeout = defaults.Timeout.Milliseconds()
	}
	threshold := current.Threshold
	if threshold <= 0 {
		threshold = defaults.Threshold
	}
	if minRequest <= 0 || retryTimeout <= 0 || threshold <= 0 {
		return nil
	}

	return &sentinelcb.Rule{
		Resource:         resource,
		Strategy:         sentinelcb.ErrorRatio,
		RetryTimeoutMs:   uint32(retryTimeout),
		MinRequestAmount: uint64(minRequest),
		StatIntervalMs:   uint32(interval.Milliseconds()),
		Threshold:        threshold,
	}
}

type resourceState struct {
	state        contract.CircuitBreakerState
	lastChanged  time.Time
	lastFailure  error
	successCount int64
	failureCount int64
}

type SentinelCircuitBreaker struct {
	cfg    *contract.CircuitBreakerConfig
	states sync.Map
}

func NewSentinelCircuitBreaker(cfg *contract.CircuitBreakerConfig) *SentinelCircuitBreaker {
	return &SentinelCircuitBreaker{cfg: cfg}
}

func (cb *SentinelCircuitBreaker) Allow(ctx context.Context, resource string) error {
	entry, blockErr := sentinelEntry(resource)
	if blockErr != nil {
		cb.markOpen(resource, blockErr)
		return fmt.Errorf("circuitbreaker: request blocked: %w", blockErr)
	}
	entry.Exit()
	cb.markHalfOpenIfRecovered(resource)
	return nil
}

func (cb *SentinelCircuitBreaker) RecordSuccess(ctx context.Context, resource string) {
	state := cb.loadState(resource)
	state.successCount++
	state.lastFailure = nil
	state.state = contract.CircuitBreakerStateClosed
	state.lastChanged = time.Now()
}

func (cb *SentinelCircuitBreaker) RecordFailure(ctx context.Context, resource string, err error) {
	state := cb.loadState(resource)
	state.failureCount++
	state.lastFailure = err
	state.state = contract.CircuitBreakerStateOpen
	state.lastChanged = time.Now()
}

func (cb *SentinelCircuitBreaker) Do(ctx context.Context, resource string, fn func() error) error {
	entry, blockErr := sentinelEntry(resource)
	if blockErr != nil {
		cb.markOpen(resource, blockErr)
		return fmt.Errorf("circuitbreaker: request blocked: %w", blockErr)
	}
	defer entry.Exit()

	execErr := fn()
	if execErr != nil {
		sentinel.TraceError(entry, execErr)
		cb.RecordFailure(ctx, resource, execErr)
		return execErr
	}

	cb.RecordSuccess(ctx, resource)
	return nil
}

func (cb *SentinelCircuitBreaker) State(ctx context.Context, resource string) contract.CircuitBreakerState {
	state := cb.loadState(resource)
	if state.state == contract.CircuitBreakerStateOpen && cb.shouldHalfOpen(state) {
		state.state = contract.CircuitBreakerStateHalfOpen
	}
	return state.state
}

func (cb *SentinelCircuitBreaker) loadState(resource string) *resourceState {
	actual, _ := cb.states.LoadOrStore(resource, &resourceState{
		state:       contract.CircuitBreakerStateClosed,
		lastChanged: time.Now(),
	})
	return actual.(*resourceState)
}

func (cb *SentinelCircuitBreaker) shouldHalfOpen(state *resourceState) bool {
	timeout := cb.cfg.DefaultConfig.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return time.Since(state.lastChanged) >= timeout
}

func (cb *SentinelCircuitBreaker) markOpen(resource string, err error) {
	state := cb.loadState(resource)
	state.state = contract.CircuitBreakerStateOpen
	state.lastFailure = err
	state.lastChanged = time.Now()
}

func (cb *SentinelCircuitBreaker) markHalfOpenIfRecovered(resource string) {
	state := cb.loadState(resource)
	if state.state == contract.CircuitBreakerStateOpen && cb.shouldHalfOpen(state) {
		state.state = contract.CircuitBreakerStateHalfOpen
		state.lastChanged = time.Now()
	}
}

func (cb *SentinelCircuitBreaker) Underlying() any {
	return sentinel.GlobalSlotChain()
}

func (cb *SentinelCircuitBreaker) As(target any) bool {
	return internalnative.As(cb.Underlying(), target)
}

type SentinelRateLimiter struct {
	cfg *contract.CircuitBreakerConfig
}

func NewSentinelRateLimiter(cfg *contract.CircuitBreakerConfig) *SentinelRateLimiter {
	return &SentinelRateLimiter{cfg: cfg}
}

func (rl *SentinelRateLimiter) Allow(ctx context.Context, resource string) error {
	entry, blockErr := sentinelEntry(resource)
	if blockErr != nil {
		return fmt.Errorf("ratelimiter: request blocked: %w", blockErr)
	}
	entry.Exit()
	return nil
}

func (rl *SentinelRateLimiter) AllowN(ctx context.Context, resource string, n int) error {
	for i := 0; i < n; i++ {
		if err := rl.Allow(ctx, resource); err != nil {
			return err
		}
	}
	return nil
}

func (rl *SentinelRateLimiter) Reserve(ctx context.Context, resource string) contract.Reservation {
	if err := rl.Allow(ctx, resource); err != nil {
		return &sentinelReservation{ok: false}
	}
	return &sentinelReservation{ok: true}
}

func (rl *SentinelRateLimiter) Wait(ctx context.Context, resource string) error {
	return rl.WaitTimeout(ctx, resource, 0)
}

func (rl *SentinelRateLimiter) WaitTimeout(ctx context.Context, resource string, timeout time.Duration) error {
	if timeout <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return rl.Allow(ctx, resource)
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return rl.Allow(ctx, resource)
	}
}

func (rl *SentinelRateLimiter) Underlying() any {
	return sentinel.GlobalSlotChain()
}

func (rl *SentinelRateLimiter) As(target any) bool {
	return internalnative.As(rl.Underlying(), target)
}

type sentinelReservation struct{ ok bool }

func (r *sentinelReservation) OK() bool             { return r.ok }
func (r *sentinelReservation) Delay() time.Duration { return 0 }
func (r *sentinelReservation) Cancel()              {}
func (r *sentinelReservation) CancelAt(time.Time)   {}
