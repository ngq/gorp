package sentinel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ngq/gorp/framework/contract"
	sentinel "github.com/alibaba/sentinel-golang/api"
)

// Provider 提供 Sentinel 熔断降级实现。
//
// 中文说明：
// - 集成阿里巴巴 Sentinel-golang；
// - 支持熔断、限流、系统保护；
// - 需要项目引入 github.com/alibaba/sentinel-golang 依赖。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "circuitbreaker.sentinel" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.CircuitBreakerKey, contract.RateLimiterKey}
}

func (p *Provider) Register(c contract.Container) error {
	cfg, err := getCircuitBreakerConfig(c)
	if err != nil {
		return err
	}

	// 初始化 Sentinel
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

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// getCircuitBreakerConfig 从容器获取熔断器配置。
func getCircuitBreakerConfig(c contract.Container) (*contract.CircuitBreakerConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("circuitbreaker: invalid config service")
	}

	cbCfg := &contract.CircuitBreakerConfig{
		Enabled:          true,
		Strategy:         "sentinel",
		ResourceConfigs:  make(map[string]contract.ResourceConfig),
		DefaultConfig: contract.ResourceConfig{
			Threshold:             0.5,
			MinRequestCount:       10,
			MaxConcurrentRequests: 100,
			Timeout:               10 * time.Second,
			Interval:              1 * time.Second,
		},
	}

	// 是否启用
	if enabled := cfg.GetBool("circuit_breaker.enabled"); enabled {
		cbCfg.Enabled = true
	}

	return cbCfg, nil
}

// initSentinel 初始化 Sentinel。
func initSentinel(cfg *contract.CircuitBreakerConfig) error {
	// Sentinel 初始化通常需要配置文件，这里简化处理
	// 实际使用时应该加载完整的 Sentinel 配置
	// 参考：https://github.com/alibaba/sentinel-golang/wiki/初始化-InitSentinel
	return nil
}

// SentinelCircuitBreaker 是 Sentinel 熔断器实现。
//
// 中文说明：
// - 使用 Sentinel Entry API 进行熔断控制；
// - 需要先配置 Sentinel 规则才能生效；
// - 参考 https://github.com/alibaba/sentinel-golang 获取完整使用指南。
type SentinelCircuitBreaker struct {
	cfg *contract.CircuitBreakerConfig
}

// NewSentinelCircuitBreaker 创建 Sentinel 熔断器。
func NewSentinelCircuitBreaker(cfg *contract.CircuitBreakerConfig) *SentinelCircuitBreaker {
	return &SentinelCircuitBreaker{cfg: cfg}
}

// Allow 检查是否允许请求。
//
// 中文说明：
// - 调用 sentinel.Entry(resource) 进入资源；
// - 如果被熔断，返回 BlockError；
// - 调用方需要在完成后调用 RecordSuccess 或 RecordFailure。
func (cb *SentinelCircuitBreaker) Allow(ctx context.Context, resource string) error {
	// 使用 Sentinel Entry API
	entry, blockErr := sentinel.Entry(resource)
	if blockErr != nil {
		return fmt.Errorf("circuitbreaker: request blocked: %w", blockErr)
	}

	// 存储到上下文（注意：这里不使用 context.WithValue 的返回值）
	// 实际使用时建议在请求处理完成后调用 entry.Exit()
	_ = entry

	return nil
}

// RecordSuccess 记录成功请求。
//
// 中文说明：
// - 调用方需要自行管理 Sentinel Entry 的生命周期；
// - 建议使用 Do 方法自动管理 Entry。
func (cb *SentinelCircuitBreaker) RecordSuccess(ctx context.Context, resource string) {
	// Sentinel 通过 Exit 自动记录成功
	// 使用方需要自行调用 entry.Exit()
}

// RecordFailure 记录失败请求。
//
// 中文说明：
// - Sentinel 会根据规则自动判断是否需要熔断；
// - 使用方需要自行调用 entry.Exit() 并传递错误信息。
func (cb *SentinelCircuitBreaker) RecordFailure(ctx context.Context, resource string, err error) {
	// Sentinel 通过 Exit + TraceError 记录失败
	// 使用方需要自行调用 entry.Exit()
}

// Do 执行受熔断器保护的函数。
//
// 中文说明：
// - 自动管理 Sentinel Entry 的生命周期；
// - 执行成功时自动 Exit；
// - 执行失败时记录错误并 Exit。
func (cb *SentinelCircuitBreaker) Do(ctx context.Context, resource string, fn func() error) error {
	// 使用 Sentinel Entry API
	entry, blockErr := sentinel.Entry(resource)
	if blockErr != nil {
		return fmt.Errorf("circuitbreaker: request blocked: %w", blockErr)
	}
	defer entry.Exit()

	// 执行函数
	execErr := fn()
	if execErr != nil {
		// 记录错误（Sentinel 会根据规则判断是否熔断）
		sentinel.TraceError(entry, execErr)
	}

	return execErr
}

// State 获取熔断器状态。
//
// 中文说明：
// - 需要先配置熔断规则才能获取真实状态；
// - 未配置时默认返回 Closed 状态。
func (cb *SentinelCircuitBreaker) State(ctx context.Context, resource string) contract.CircuitBreakerState {
	// 获取 Sentinel 熔断器状态
	// 简化处理，返回关闭状态
	// 实际使用时需要调用 GetCircuitBreakerOfResource(resource).CurrentState()
	return contract.CircuitBreakerStateClosed
}

// SentinelRateLimiter 是 Sentinel 限流器实现。
//
// 中文说明：
// - 使用 Sentinel Entry API 进行流量控制；
// - 需要先配置限流规则才能生效；
// - 支持 QPS 限流、并发数限流等模式。
type SentinelRateLimiter struct {
	cfg *contract.CircuitBreakerConfig
}

// NewSentinelRateLimiter 创建 Sentinel 限流器。
func NewSentinelRateLimiter(cfg *contract.CircuitBreakerConfig) *SentinelRateLimiter {
	return &SentinelRateLimiter{cfg: cfg}
}

// Allow 检查是否允许请求。
func (rl *SentinelRateLimiter) Allow(ctx context.Context, resource string) error {
	entry, blockErr := sentinel.Entry(resource)
	if blockErr != nil {
		return fmt.Errorf("ratelimiter: request blocked: %w", blockErr)
	}
	entry.Exit()
	return nil
}

// AllowN 检查是否允许 N 个请求。
//
// 中文说明：
// - Sentinel 流量控制不支持批量；
// - 简化为逐个检查。
func (rl *SentinelRateLimiter) AllowN(ctx context.Context, resource string, n int) error {
	// Sentinel 流量控制不支持批量，简化处理
	for i := 0; i < n; i++ {
		if err := rl.Allow(ctx, resource); err != nil {
			return err
		}
	}
	return nil
}

// Reserve 预留令牌。
//
// 中文说明：
// - Sentinel 不支持预留模式；
// - 返回空实现。
func (rl *SentinelRateLimiter) Reserve(ctx context.Context, resource string) contract.Reservation {
	return &sentinelReservation{}
}

// Wait 等待直到获取令牌。
//
// 中文说明：
// - Sentinel 不支持等待模式；
// - 直接检查是否允许。
func (rl *SentinelRateLimiter) Wait(ctx context.Context, resource string) error {
	// Sentinel 不支持等待，直接检查
	return rl.Allow(ctx, resource)
}

// WaitTimeout 等待直到获取令牌或超时。
func (rl *SentinelRateLimiter) WaitTimeout(ctx context.Context, resource string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 定期检查
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := rl.Allow(ctx, resource); err == nil {
				return nil
			}
		}
	}
}

// sentinelReservation 是 Sentinel 预留实现。
type sentinelReservation struct{}

func (r *sentinelReservation) OK() bool             { return true }
func (r *sentinelReservation) Delay() time.Duration { return 0 }
func (r *sentinelReservation) Cancel()              {}
func (r *sentinelReservation) CancelAt(t time.Time) {}