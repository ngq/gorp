// Application scenarios:
// - Protect external APIs from high-frequency bursts.
// - Apply route-level or resource-level request quotas.
// - Combine with identity or IP extraction to build access policies.
//
// 适用场景：
// - 保护对外 API，避免高频突发流量冲击。
// - 对路由级或资源级请求配额进行控制。
// - 结合身份或 IP 提取逻辑构建访问策略。
package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// RateLimiter is the native Gin limiter contract used by RateLimitMiddleware.
//
// RateLimiter 是 RateLimitMiddleware 使用的原生 Gin 限流器契约。
type RateLimiter interface {
	Allow(key string) bool
}

// RateLimit applies a transport-level rate limiter and returns a unified 429 response when exceeded.
//
// RateLimit 应用 transport 层限流器，并在超限时返回统一的 429 响应。
func RateLimit(limiter resiliencecontract.RateLimiter, resource string) transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			if limiter == nil {
				if next != nil {
					next(c)
				}
				return
			}
			target := resource
			if target == "" {
				target = c.RoutePath()
			}
			if target == "" && c.Request() != nil && c.Request().URL != nil {
				target = c.Request().Method + " " + c.Request().URL.Path
			}
			if err := limiter.Allow(c, target); err != nil {
				if gc, ok := unwrapGinContext(c); ok {
					writeGinResponseHeaders(gc)
					resp := Response{
						Code:    CodeTooManyRequests,
						Message: "rate limit exceeded",
						Data:    nil,
					}
					gc.JSON(http.StatusTooManyRequests, resp)
					gc.Abort()
					return
				}
				c.JSON(http.StatusTooManyRequests, map[string]any{
					"code":    CodeTooManyRequests,
					"message": "rate limit exceeded",
				})
				return
			}
			if next != nil {
				next(c)
			}
		}
	}
}

// RateLimitMiddleware is the native Gin form of the rate-limiting middleware.
//
// RateLimitMiddleware 是限流中间件的原生 Gin 形态。
func RateLimitMiddleware(limiter RateLimiter, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if !limiter.Allow(key) {
			writeGinResponseHeaders(c)
			resp := Response{
				Code:    CodeTooManyRequests,
				Message: "rate limit exceeded",
				Data:    nil,
			}
			c.JSON(http.StatusTooManyRequests, resp)
			c.Abort()
			return
		}
		c.Next()
	}
}

// IPKeyFunc extracts a client-IP-based rate-limit key from the request.
// X-Forwarded-For and X-Real-IP headers are ONLY trusted when the request
// originates from a trusted proxy (see SetTrustedProxies in ip_filter.go).
// Otherwise, RemoteAddr is used directly to prevent spoofing.
//
// IPKeyFunc 从请求中提取基于客户端 IP 的限流 key。
// X-Forwarded-For 和 X-Real-IP 头仅在请求来自可信代理时才会被信任
// （见 ip_filter.go 中的 SetTrustedProxies）。
// 否则直接使用 RemoteAddr，以防止伪造绕过限流。
func IPKeyFunc(c *gin.Context) string {
	remoteAddr := c.Request.RemoteAddr

	if isFromTrustedProxy(remoteAddr) {
		if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
			// Take only the first IP from the comma-separated list
			parts := strings.SplitN(xff, ",", 2)
			return strings.TrimSpace(parts[0])
		}
		if xri := c.GetHeader("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}

	return c.ClientIP()
}

// TokenBucketLimiter is an in-memory token-bucket limiter for lightweight HTTP protection.
//
// TokenBucketLimiter 是一个用于轻量 HTTP 保护的内存令牌桶限流器。
type TokenBucketLimiter struct {
	rate     float64
	burst    int
	tokens   float64
	lastTime time.Time
	mu       sync.Mutex
}

// NewTokenBucketLimiter creates an in-memory token-bucket limiter.
//
// NewTokenBucketLimiter 创建一个内存令牌桶限流器。
func NewTokenBucketLimiter(rate float64, burst int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:     rate,
		burst:    burst,
		tokens:   float64(burst),
		lastTime: time.Now(),
	}
}

// Allow consumes one token when available and reports whether the request may proceed.
//
// Allow 在令牌可用时消耗一个令牌，并返回请求是否允许继续。
func (l *TokenBucketLimiter) Allow(_ string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds()
	l.lastTime = now
	l.tokens += elapsed * l.rate
	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}
	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

// SlidingWindowLimiter is an in-memory sliding-window limiter.
//
// SlidingWindowLimiter 是一个内存滑动窗口限流器。
type SlidingWindowLimiter struct {
	limit  int
	window time.Duration
	counts map[string]*windowRecord
	mu     sync.Mutex
}

type windowRecord struct {
	timestamps []time.Time
}

// NewSlidingWindowLimiter creates an in-memory sliding-window limiter.
//
// NewSlidingWindowLimiter 创建一个内存滑动窗口限流器。
func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		limit:  limit,
		window: window,
		counts: make(map[string]*windowRecord),
	}
}

// Allow checks whether the key is still within the sliding-window quota.
//
// Allow 判断给定 key 是否仍在滑动窗口配额内。
func (l *SlidingWindowLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	threshold := now.Add(-l.window)
	record, exists := l.counts[key]
	if !exists {
		l.counts[key] = &windowRecord{timestamps: []time.Time{now}}
		return true
	}
	validIdx := len(record.timestamps)
	for i, t := range record.timestamps {
		if t.After(threshold) {
			validIdx = i
			break
		}
	}
	record.timestamps = record.timestamps[validIdx:]
	if len(record.timestamps) >= l.limit {
		return false
	}
	record.timestamps = append(record.timestamps, now)
	return true
}

// FixedWindowLimiter is an in-memory fixed-window limiter.
//
// FixedWindowLimiter 是一个内存固定窗口限流器。
type FixedWindowLimiter struct {
	limit   int
	window  time.Duration
	counts  map[string]int
	windows map[string]time.Time
	mu      sync.Mutex
}

// NewFixedWindowLimiter creates an in-memory fixed-window limiter.
//
// NewFixedWindowLimiter 创建一个内存固定窗口限流器。
func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		limit:   limit,
		window:  window,
		counts:  make(map[string]int),
		windows: make(map[string]time.Time),
	}
}

// Allow checks whether the key is still within the fixed-window quota.
//
// Allow 判断给定 key 是否仍在固定窗口配额内。
func (l *FixedWindowLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart, exists := l.windows[key]
	if !exists || now.Sub(windowStart) >= l.window {
		l.windows[key] = now
		l.counts[key] = 1
		return true
	}
	if l.counts[key] >= l.limit {
		return false
	}
	l.counts[key]++
	return true
}
