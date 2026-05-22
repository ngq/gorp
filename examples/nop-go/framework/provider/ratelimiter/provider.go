// Package ratelimiter provides rate limiting capability for gorp framework.
// Based on golang.org/x/time/rate (official Go extension).
//
// 限流器 provider，基于 golang.org/x/time/rate（Go 官方扩展库）。
package ratelimiter

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider 是基于 golang.org/x/time/rate 的限流器 provider。
// 将 RateLimiter 契约实现注册到容器中，供 HTTP 中间件和 RPC 客户端使用。
type Provider struct{}

// NewProvider 创建限流器 provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 provider 唯一名称。
func (p *Provider) Name() string { return "ratelimiter.tokenbucket" }

// IsDefer 标记此 provider 延迟装载。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回该 provider 提供的容器 key 列表。
func (p *Provider) Provides() []string {
	return []string{resiliencecontract.RateLimiterKey}
}

// DependsOn 返回该 provider 依赖的 key。
func (p *Provider) DependsOn() []string { return nil }

// Register 将 RateLimiter 实例注册到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(resiliencecontract.RateLimiterKey, func(c runtimecontract.Container) (any, error) {
		cfg := loadRateLimiterConfig(c)
		return newTokenBucketRateLimiter(cfg), nil
	}, true)
	return nil
}

// Boot 启动期初始化（无额外操作）。
func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

// --- 令牌桶限流器实现 ---

// tokenBucketRateLimiter 是 RateLimiter 契约的令牌桶实现。
// 基于 golang.org/x/time/rate.Limiter，支持 Allow、Wait、Reserve 全套接口。
type tokenBucketRateLimiter struct {
	limiter   *rate.Limiter
	mu        sync.Mutex
	limiters  sync.Map // resource -> *rate.Limiter
	config    resiliencecontract.RateLimiterConfig
}

// newTokenBucketRateLimiter 根据配置创建令牌桶限流器。
func newTokenBucketRateLimiter(cfg resiliencecontract.RateLimiterConfig) *tokenBucketRateLimiter {
	// 默认配置：100 QPS，burst 200
	defaultQPS := cfg.DefaultConfig.QPS
	if defaultQPS <= 0 {
		defaultQPS = 100
	}
	defaultBurst := cfg.DefaultConfig.Burst
	if defaultBurst <= 0 {
		defaultBurst = 200
	}

	return &tokenBucketRateLimiter{
		limiter: rate.NewLimiter(rate.Limit(defaultQPS), defaultBurst),
		config:  cfg,
	}
}

// NewTokenBucketRateLimiter 导出的构造函数，供测试和外部使用。
func NewTokenBucketRateLimiter(cfg resiliencecontract.RateLimiterConfig) *tokenBucketRateLimiter {
	return newTokenBucketRateLimiter(cfg)
}

// Allow 尝试获取一个令牌，立即返回结果（非阻塞）。
// 如果令牌可用，返回 nil；否则返回 ErrRateLimited。
func (l *tokenBucketRateLimiter) Allow(ctx context.Context, resource string) error {
	limiter := l.getOrCreateLimiter(resource)
	if !limiter.Allow() {
		return ErrRateLimited
	}
	return nil
}

// AllowN 尝试获取 n 个令牌，立即返回结果（非阻塞）。
func (l *tokenBucketRateLimiter) AllowN(ctx context.Context, resource string, n int) error {
	limiter := l.getOrCreateLimiter(resource)
	if !limiter.AllowN(time.Now(), n) {
		return ErrRateLimited
	}
	return nil
}

// Reserve 预留令牌，返回 Reservation 信息。
// 可用于判断延迟时间、取消预留等高级场景。
func (l *tokenBucketRateLimiter) Reserve(ctx context.Context, resource string) resiliencecontract.Reservation {
	limiter := l.getOrCreateLimiter(resource)
	r := limiter.Reserve()
	return &tokenBucketReservation{r: r}
}

// Wait 阻塞等待直到获取令牌，或 ctx 被取消。
func (l *tokenBucketRateLimiter) Wait(ctx context.Context, resource string) error {
	limiter := l.getOrCreateLimiter(resource)
	return limiter.Wait(ctx)
}

// WaitTimeout 阻塞等待直到获取令牌，或超时。
func (l *tokenBucketRateLimiter) WaitTimeout(ctx context.Context, resource string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	limiter := l.getOrCreateLimiter(resource)
	return limiter.Wait(ctx)
}

// getOrCreateLimiter 获取或创建资源对应的限流器。
// 不同资源可以有独立的 QPS 和 Burst 配置。
func (l *tokenBucketRateLimiter) getOrCreateLimiter(resource string) *rate.Limiter {
	// 快速路径：已有限流器直接返回
	if v, ok := l.limiters.Load(resource); ok {
		return v.(*rate.Limiter)
	}

	// 确定该资源的 QPS 和 Burst
	qps := l.config.DefaultConfig.QPS
	burst := l.config.DefaultConfig.Burst
	if cfg, ok := l.config.ResourceConfigs[resource]; ok {
		if cfg.QPS > 0 {
			qps = cfg.QPS
		}
		if cfg.Burst > 0 {
			burst = cfg.Burst
		}
	}

	// 慢路径：创建新限流器
	limiter := rate.NewLimiter(rate.Limit(qps), burst)
	actual, _ := l.limiters.LoadOrStore(resource, limiter)
	return actual.(*rate.Limiter)
}

// --- Reservation 实现 ---

// tokenBucketReservation 包装 rate.Reservation，实现契约接口。
type tokenBucketReservation struct {
	r *rate.Reservation
}

// OK 返回预留是否成功。
func (r *tokenBucketReservation) OK() bool {
	return r.r.OK()
}

// Delay 返回需要等待的时间。
func (r *tokenBucketReservation) Delay() time.Duration {
	return r.r.Delay()
}

// Cancel 取消预留。
func (r *tokenBucketReservation) Cancel() {
	r.r.Cancel()
}

// CancelAt 在指定时间取消预留。
func (r *tokenBucketReservation) CancelAt(t time.Time) {
	r.r.CancelAt(t)
}

// --- 错误定义 ---

// ErrRateLimited 表示请求被限流。
var ErrRateLimited = resiliencecontract.ServiceUnavailable("rate limit exceeded")

// --- 配置读取辅助 ---

// loadRateLimiterConfig 从容器中读取 rate_limiter 配置。
func loadRateLimiterConfig(c runtimecontract.Container) resiliencecontract.RateLimiterConfig {
	cfg := resiliencecontract.RateLimiterConfig{}

	if c == nil || !c.IsBind("framework.config") {
		return cfg
	}

	configAny, err := c.Make("framework.config")
	if err != nil {
		return cfg
	}

	type configGetter interface {
		GetBool(key string) bool
		GetString(key string) string
		GetFloat64(key string) float64
		GetInt(key string) int
		Get(key string) any
	}

	getter, ok := configAny.(configGetter)
	if !ok {
		return cfg
	}

	// 全局开关
	cfg.Enabled = getter.GetBool("rate_limiter.enabled")

	// 策略类型
	cfg.Strategy = getter.GetString("rate_limiter.strategy")

	// 默认配置
	cfg.DefaultConfig.QPS = getter.GetFloat64("rate_limiter.default.qps")
	cfg.DefaultConfig.Burst = getter.GetInt("rate_limiter.default.burst")
	if maxWaitStr := getter.GetString("rate_limiter.default.max_wait"); maxWaitStr != "" {
		cfg.DefaultConfig.MaxWait, _ = time.ParseDuration(maxWaitStr)
	}

	// 按资源粒度的配置
	cfg.ResourceConfigs = loadResourceConfigs(getter)

	return cfg
}

// loadResourceConfigs 从配置中读取按资源粒度的限流配置。
func loadResourceConfigs(getter interface {
	Get(key string) any
	GetString(key string) string
	GetFloat64(key string) float64
	GetInt(key string) int
}) map[string]resiliencecontract.RateResourceConfig {
	configs := make(map[string]resiliencecontract.RateResourceConfig)

	raw := getter.Get("rate_limiter.resources")
	if raw == nil {
		return configs
	}

	// 支持 map[string]map[string]any 格式
	if m, ok := raw.(map[string]any); ok {
		for resource, cfgRaw := range m {
			if cm, ok := cfgRaw.(map[string]any); ok {
				cfg := resiliencecontract.RateResourceConfig{}
				if v, ok := cm["qps"].(float64); ok && v > 0 {
					cfg.QPS = v
				}
				if v, ok := cm["burst"].(int); ok && v > 0 {
					cfg.Burst = v
				}
				if v, ok := cm["burst"].(float64); ok && v > 0 {
					cfg.Burst = int(v)
				}
				if v, ok := cm["max_wait"].(string); ok {
					cfg.MaxWait, _ = time.ParseDuration(v)
				}
				configs[resource] = cfg
			}
		}
	}

	return configs
}