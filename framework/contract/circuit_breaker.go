package contract

import (
	"context"
	"time"
)

const (
	// CircuitBreakerKey 是熔断器在容器中的绑定 key。
	//
	// 中文说明：
	// - 用于服务熔断和降级；
	// - 保护服务免受级联故障影响；
	// - noop 实现空操作，单体项目零依赖。
	CircuitBreakerKey = "framework.circuit_breaker"

	// RateLimiterKey 是限流器在容器中的绑定 key。
	//
	// 中文说明：
	// - 用于限流控制；
	// - 保护服务免受过载影响。
	RateLimiterKey = "framework.rate_limiter"
)

// CircuitBreaker 熔断器接口。
//
// 中文说明：
// - 实现熔断器模式；
// - 支持三种状态：关闭、打开、半开；
// - 自动故障检测和恢复。
type CircuitBreaker interface {
	// Allow 检查是否允许请求。
	//
	// 中文说明：
	// - 返回 nil 表示允许；
	// - 返回错误表示熔断器打开，拒绝请求。
	Allow(ctx context.Context, resource string) error

	// RecordSuccess 记录成功请求。
	RecordSuccess(ctx context.Context, resource string)

	// RecordFailure 记录失败请求。
	RecordFailure(ctx context.Context, resource string, err error)

	// Do 执行受熔断器保护的函数。
	//
	// 中文说明：
	// - 自动检查熔断状态；
	// - 自动记录成功/失败。
	Do(ctx context.Context, resource string, fn func() error) error

	// State 获取熔断器状态。
	State(ctx context.Context, resource string) CircuitBreakerState
}

// CircuitBreakerState 熔断器状态。
type CircuitBreakerState int

const (
	CircuitBreakerStateClosed CircuitBreakerState = iota // 关闭（正常）
	CircuitBreakerStateOpen                               // 打开（熔断）
	CircuitBreakerStateHalfOpen                           // 半开（尝试恢复）
)

// String 返回状态字符串。
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

// RateLimiter 限流器接口。
//
// 中文说明：
// - 实现限流控制；
// - 支持令牌桶、漏桶等算法。
type RateLimiter interface {
	// Allow 检查是否允许请求。
	//
	// 中文说明：
	// - 返回 nil 表示允许；
	// - 返回错误表示被限流。
	Allow(ctx context.Context, resource string) error

	// AllowN 检查是否允许 N 个请求。
	AllowN(ctx context.Context, resource string, n int) error

	// Reserve 预留令牌。
	Reserve(ctx context.Context, resource string) Reservation

	// Wait 等待直到获取令牌。
	Wait(ctx context.Context, resource string) error

	// WaitTimeout 等待直到获取令牌或超时。
	WaitTimeout(ctx context.Context, resource string, timeout time.Duration) error
}

// Reservation 令牌预留。
type Reservation interface {
	// OK 是否成功预留。
	OK() bool

	// Delay 等待时间。
	Delay() time.Duration

	// Cancel 取消预留。
	Cancel()

	// CancelAt 在指定时间取消预留。
	CancelAt(t time.Time)
}

// CircuitBreakerConfig 熔断器配置。
type CircuitBreakerConfig struct {
	// Enabled 是否启用
	Enabled bool

	// Strategy 策略：noop/sentinel/hystrix
	Strategy string

	// 资源配置
	ResourceConfigs map[string]ResourceConfig

	// 默认配置
	DefaultConfig ResourceConfig
}

// ResourceConfig 资源熔断配置。
type ResourceConfig struct {
	// Threshold 熔断阈值（错误率或错误数）
	Threshold float64

	// MinRequestCount 最小请求数（触发熔断的最小请求数）
	MinRequestCount int64

	// MaxConcurrentRequests 最大并发请求数
	MaxConcurrentRequests int64

	// Timeout 熔断超时时间（熔断器打开后的等待时间）
	Timeout time.Duration

	// RetryTimeoutMs 重试超时（毫秒）
	RetryTimeoutMs int64

	// Interval 统计时间窗口
	Interval time.Duration
}

// RateLimiterConfig 限流器配置。
type RateLimiterConfig struct {
	// Enabled 是否启用
	Enabled bool

	// Strategy 策略：noop/sentinel/token-bucket/leaky-bucket
	Strategy string

	// 资源配置
	ResourceConfigs map[string]RateResourceConfig

	// 默认配置
	DefaultConfig RateResourceConfig
}

// RateResourceConfig 资源限流配置。
type RateResourceConfig struct {
	// QPS 每秒请求数
	QPS float64

	// Burst 突发流量
	Burst int

	// MaxWait 最大等待时间
	MaxWait time.Duration
}