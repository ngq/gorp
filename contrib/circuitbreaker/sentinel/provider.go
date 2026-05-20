// Package sentinel provides Sentinel circuit breaker implementation for gorp.
//
// Sentinel 熔断降级 Provider，实现 resiliencecontract.CircuitBreaker 契约。
// 支持熔断、限流、系统保护规则。
//
// 使用示例：
//
//  cfg := &CircuitBreakerConfig{
//      Resource: "my-service",
//      Strategy: "error_ratio",
//      Threshold: 0.5,
//  }
//  cb, err := NewCircuitBreaker(cfg)
//  if err != nil {
//      panic(err)
//  }
//
//  if cb.Allow() {
//      // 执行请求
//  } else {
//      // 熔断降级
//  }
//
// 配置路径：circuitbreaker.sentinel.*
package sentinel

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	sentinelcb "github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/isolation"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
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

// DependsOn returns the keys this provider depends on.
// Sentinel circuit breaker depends on Config for rule configuration.
//
// DependsOn 返回该 provider 依赖的 key。
// Sentinel circuit breaker 依赖 Config 获取规则配置。
func (p *Provider) DependsOn() []string { return []string{datacontract.ConfigKey} }
func (p *Provider) Provides() []string {
	return []string{resiliencecontract.CircuitBreakerKey, resiliencecontract.RateLimiterKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	cfg, err := getCircuitBreakerConfig(c)
	if err != nil {
		return err
	}
	if cfg.Enabled {
		if err := initSentinel(cfg); err != nil {
			return fmt.Errorf("circuitbreaker.sentinel: init failed: %w", err)
		}
	}
	c.Bind(resiliencecontract.CircuitBreakerKey, func(c runtimecontract.Container) (any, error) {
		cb := NewSentinelCircuitBreaker(cfg)
		c.RegisterCloser(resiliencecontract.CircuitBreakerKey, cb)
		return cb, nil
	}, true)
	c.Bind(resiliencecontract.RateLimiterKey, func(c runtimecontract.Container) (any, error) {
		rl := NewSentinelRateLimiter(cfg)
		c.RegisterCloser(resiliencecontract.RateLimiterKey, rl)
		return rl, nil
	}, true)
	return nil
}
func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

func getCircuitBreakerConfig(c runtimecontract.Container) (*resiliencecontract.CircuitBreakerConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("circuitbreaker.sentinel: invalid config service")
	}

	enabled := true
	if cfg.Get("circuit_breaker.enabled") != nil {
		enabled = cfg.GetBool("circuit_breaker.enabled")
	}

	cbCfg := &resiliencecontract.CircuitBreakerConfig{
		Enabled:         enabled,
		Strategy:        "sentinel",
		ResourceConfigs: make(map[string]resiliencecontract.ResourceConfig),
		DefaultConfig: resiliencecontract.ResourceConfig{
			Threshold:             0.5,
			MinRequestCount:       10,
			MaxConcurrentRequests: 100,
			Timeout:               10 * time.Second,
			Interval:              1 * time.Second,
		},
	}

	// Load per-resource config from circuit_breaker.resources map in config.
	// Each key under circuit_breaker.resources is treated as a resource name,
	// with nested fields: threshold, min_request_count, timeout, interval, strategy.
	if resources := cfg.Get("circuit_breaker.resources"); resources != nil {
		if resMap, ok := resources.(map[string]any); ok {
			for name, val := range resMap {
				if sub, ok := val.(map[string]any); ok {
					rc := cbCfg.DefaultConfig // start from defaults
					if v, ok := sub["threshold"]; ok {
						if f, ok := v.(float64); ok {
							rc.Threshold = f
						}
					}
					if v, ok := sub["min_request_count"]; ok {
						if f, ok := v.(float64); ok {
							rc.MinRequestCount = int64(f)
						}
					}
					if v, ok := sub["timeout"]; ok {
						if s, ok := v.(string); ok {
							if d, err := time.ParseDuration(s); err == nil {
								rc.Timeout = d
							}
						}
					}
					if v, ok := sub["interval"]; ok {
						if s, ok := v.(string); ok {
							if d, err := time.ParseDuration(s); err == nil {
								rc.Interval = d
							}
						}
					}
					cbCfg.ResourceConfigs[name] = rc
				}
			}
		}
	}

	return cbCfg, nil
}

func initSentinel(cfg *resiliencecontract.CircuitBreakerConfig) error {
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

func buildCircuitBreakerRule(resource string, current resiliencecontract.ResourceConfig, defaults resiliencecontract.ResourceConfig) *sentinelcb.Rule {
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
		Strategy:         sentinelcb.ErrorRatio, // Default strategy; per-resource strategy not yet in ResourceConfig
		RetryTimeoutMs:   uint32(retryTimeout),
		MinRequestAmount: uint64(minRequest),
		StatIntervalMs:   uint32(interval.Milliseconds()),
		Threshold:        threshold,
	}
}

type resourceState struct {
	mu           sync.Mutex
	state        resiliencecontract.CircuitBreakerState
	lastChanged  time.Time
	lastFailure  error
	successCount atomic.Int64
	failureCount atomic.Int64
}

type SentinelCircuitBreaker struct {
	cfg    *resiliencecontract.CircuitBreakerConfig
	states sync.Map
}

func NewSentinelCircuitBreaker(cfg *resiliencecontract.CircuitBreakerConfig) *SentinelCircuitBreaker {
	return &SentinelCircuitBreaker{cfg: cfg}
}

func (cb *SentinelCircuitBreaker) Allow(ctx context.Context, resource string) error {
	entry, blockErr := sentinelEntry(resource)
	if blockErr != nil {
		cb.markOpen(resource, blockErr)
		return fmt.Errorf("circuitbreaker.sentinel: request blocked: %w", blockErr)
	}
	entry.Exit()
	cb.markHalfOpenIfRecovered(resource)
	return nil
}

func (cb *SentinelCircuitBreaker) RecordSuccess(ctx context.Context, resource string) {
	state := cb.loadState(resource)
	state.successCount.Add(1)
	state.mu.Lock()
	state.lastFailure = nil
	state.state = resiliencecontract.CircuitBreakerStateClosed
	state.lastChanged = time.Now()
	state.mu.Unlock()
}

func (cb *SentinelCircuitBreaker) RecordFailure(ctx context.Context, resource string, err error) {
	state := cb.loadState(resource)
	state.failureCount.Add(1)
	state.mu.Lock()
	state.lastFailure = err
	state.state = resiliencecontract.CircuitBreakerStateOpen
	state.lastChanged = time.Now()
	state.mu.Unlock()
}

func (cb *SentinelCircuitBreaker) Do(ctx context.Context, resource string, fn func() error) error {
	entry, blockErr := sentinelEntry(resource)
	if blockErr != nil {
		cb.markOpen(resource, blockErr)
		return fmt.Errorf("circuitbreaker.sentinel: request blocked: %w", blockErr)
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

func (cb *SentinelCircuitBreaker) State(ctx context.Context, resource string) resiliencecontract.CircuitBreakerState {
	state := cb.loadState(resource)
	state.mu.Lock()
	result := state.state
	if result == resiliencecontract.CircuitBreakerStateOpen && cb.shouldHalfOpen(state) {
		state.state = resiliencecontract.CircuitBreakerStateHalfOpen
		result = resiliencecontract.CircuitBreakerStateHalfOpen
		state.lastChanged = time.Now()
	}
	state.mu.Unlock()
	return result
}

func (cb *SentinelCircuitBreaker) loadState(resource string) *resourceState {
	actual, _ := cb.states.LoadOrStore(resource, &resourceState{
		state:       resiliencecontract.CircuitBreakerStateClosed,
		lastChanged: time.Now(),
	})
	return actual.(*resourceState)
}

// shouldHalfOpen must be called with state.mu held.
func (cb *SentinelCircuitBreaker) shouldHalfOpen(state *resourceState) bool {
	timeout := cb.cfg.DefaultConfig.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return time.Since(state.lastChanged) >= timeout
}

func (cb *SentinelCircuitBreaker) markOpen(resource string, err error) {
	state := cb.loadState(resource)
	state.mu.Lock()
	state.state = resiliencecontract.CircuitBreakerStateOpen
	state.lastFailure = err
	state.lastChanged = time.Now()
	state.mu.Unlock()
}

func (cb *SentinelCircuitBreaker) markHalfOpenIfRecovered(resource string) {
	state := cb.loadState(resource)
	state.mu.Lock()
	if state.state == resiliencecontract.CircuitBreakerStateOpen && cb.shouldHalfOpen(state) {
		state.state = resiliencecontract.CircuitBreakerStateHalfOpen
		state.lastChanged = time.Now()
	}
	state.mu.Unlock()
}

func (cb *SentinelCircuitBreaker) Underlying() any {
	return sentinel.GlobalSlotChain()
}

// Close releases resources held by the circuit breaker.
// It clears all internal state entries so they can be garbage collected.
func (cb *SentinelCircuitBreaker) Close() error {
	cb.states.Range(func(key, _ any) bool {
		cb.states.Delete(key)
		return true
	})
	return nil
}

func (cb *SentinelCircuitBreaker) As(target any) bool {
	return As(cb.Underlying(), target)
}

type SentinelRateLimiter struct {
	cfg *resiliencecontract.CircuitBreakerConfig
}

func NewSentinelRateLimiter(cfg *resiliencecontract.CircuitBreakerConfig) *SentinelRateLimiter {
	return &SentinelRateLimiter{cfg: cfg}
}

func (rl *SentinelRateLimiter) Allow(ctx context.Context, resource string) error {
	entry, blockErr := sentinelEntry(resource)
	if blockErr != nil {
		return fmt.Errorf("circuitbreaker.sentinel: request blocked: %w", blockErr)
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

func (rl *SentinelRateLimiter) Reserve(ctx context.Context, resource string) resiliencecontract.Reservation {
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
	return As(rl.Underlying(), target)
}

// Close is a no-op for the rate limiter.
// The underlying Sentinel engine is a global singleton managed by the Provider lifecycle.
func (rl *SentinelRateLimiter) Close() error {
	return nil
}

type sentinelReservation struct{ ok bool }

func (r *sentinelReservation) OK() bool             { return r.ok }
func (r *sentinelReservation) Delay() time.Duration { return 0 }
func (r *sentinelReservation) Cancel()              {}
func (r *sentinelReservation) CancelAt(time.Time)   {}

// mapSentinelStrategy converts a string strategy name to the corresponding
// sentinel Rule Strategy constant. Falls back to ErrorRatio for unknown values.
//
// mapSentinelStrategy 将字符串策略名映射为 sentinel Rule Strategy 常量。
// 未知值回退为 ErrorRatio。
func mapSentinelStrategy(strategy string) sentinelcb.Strategy {
	switch strings.ToLower(strings.TrimSpace(strategy)) {
	case "error_count", "errorcount":
		return sentinelcb.ErrorCount
	case "slow_request_ratio", "slowrequestratio", "slow_request":
		return sentinelcb.SlowRequestRatio
	case "error_ratio", "errorratio", "":
		return sentinelcb.ErrorRatio
	default:
		return sentinelcb.ErrorRatio
	}
}
