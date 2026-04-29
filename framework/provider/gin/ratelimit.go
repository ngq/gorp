package gin

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 限流器接口。
//
// 中文说明：
// - 定义通用的限流器接口，支持不同的限流算法实现；
// - Allow 方法返回 true 表示允许请求通过，false 表示被限流；
// - 具体实现可以是令牌桶、漏桶、滑动窗口等算法。
type RateLimiter interface {
	Allow(key string) bool
}

// RateLimitMiddleware 创建限流中间件。
//
// 中文说明：
// - 基于 RateLimiter 接口创建 Gin 限流中间件；
// - keyFunc 用于从请求中提取限流 key（如 IP、用户 ID 等）；
// - 被限流的请求返回 429 Too Many Requests。
func RateLimitMiddleware(limiter RateLimiter, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if !limiter.Allow(key) {
			writeResponseHeaders(c)
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

// IPKeyFunc 从请求中提取 IP 作为限流 key。
//
// 中文说明：
// - 默认的限流 key 提取函数；
// - 优先使用 X-Forwarded-For，其次 X-Real-IP，最后使用 RemoteAddr。
func IPKeyFunc(c *gin.Context) string {
	// 优先检查 X-Forwarded-For
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		return xff
	}
	// 其次检查 X-Real-IP
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	// 最后使用 RemoteAddr
	return c.ClientIP()
}

// TokenBucketLimiter 令牌桶限流器。
//
// 中文说明：
// - 令牌桶算法实现，支持突发流量；
// - rate: 每秒放入令牌的数量；
// - burst: 桶的最大容量；
// - 适合允许一定突发流量的场景。
type TokenBucketLimiter struct {
	rate     float64       // 每秒放入令牌数
	burst    int           // 桶最大容量
	tokens   float64       // 当前令牌数
	lastTime time.Time     // 上次更新时间
	mu       sync.Mutex    // 互斥锁
}

// NewTokenBucketLimiter 创建令牌桶限流器。
//
// 中文说明：
// - rate: 每秒放入令牌的数量；
// - burst: 桶的最大容量。
func NewTokenBucketLimiter(rate float64, burst int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:     rate,
		burst:    burst,
		tokens:   float64(burst),
		lastTime: time.Now(),
	}
}

// Allow 判断是否允许请求通过。
func (l *TokenBucketLimiter) Allow(_ string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds()
	l.lastTime = now

	// 补充令牌
	l.tokens += elapsed * l.rate
	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}

	// 判断是否有令牌
	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

// SlidingWindowLimiter 滑动窗口限流器。
//
// 中文说明：
// - 滑动窗口算法实现，精确控制时间窗口内的请求数；
// - limit: 时间窗口内允许的最大请求数；
// - window: 时间窗口大小；
// - 相比固定窗口，能更平滑地处理边界问题。
type SlidingWindowLimiter struct {
	limit   int           // 窗口内最大请求数
	window  time.Duration // 时间窗口
	counts  map[string]*windowRecord
	mu      sync.Mutex
}

type windowRecord struct {
	timestamps []time.Time
}

// NewSlidingWindowLimiter 创建滑动窗口限流器。
//
// 中文说明：
// - limit: 时间窗口内允许的最大请求数；
// - window: 时间窗口大小。
func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		limit:  limit,
		window: window,
		counts: make(map[string]*windowRecord),
	}
}

// Allow 判断是否允许请求通过。
func (l *SlidingWindowLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	threshold := now.Add(-l.window)

	record, exists := l.counts[key]
	if !exists {
		l.counts[key] = &windowRecord{
			timestamps: []time.Time{now},
		}
		return true
	}

	// 移除过期的请求记录
	validIdx := 0
	for i, t := range record.timestamps {
		if t.After(threshold) {
			validIdx = i
			break
		}
	}
	record.timestamps = record.timestamps[validIdx:]

	// 判断是否超过限制
	if len(record.timestamps) >= l.limit {
		return false
	}

	record.timestamps = append(record.timestamps, now)
	return true
}

// FixedWindowLimiter 固定窗口限流器。
//
// 中文说明：
// - 固定窗口算法实现，最简单的限流方式；
// - limit: 时间窗口内允许的最大请求数；
// - window: 时间窗口大小；
// - 简单高效，但在窗口边界可能出现突发流量。
type FixedWindowLimiter struct {
	limit   int           // 窗口内最大请求数
	window  time.Duration // 时间窗口
	counts  map[string]int
	windows map[string]time.Time
	mu      sync.Mutex
}

// NewFixedWindowLimiter 创建固定窗口限流器。
func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		limit:   limit,
		window:  window,
		counts:  make(map[string]int),
		windows: make(map[string]time.Time),
	}
}

// Allow 判断是否允许请求通过。
func (l *FixedWindowLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart, exists := l.windows[key]

	// 如果窗口不存在或已过期，重置窗口
	if !exists || now.Sub(windowStart) >= l.window {
		l.windows[key] = now
		l.counts[key] = 1
		return true
	}

	// 检查是否超过限制
	if l.counts[key] >= l.limit {
		return false
	}

	l.counts[key]++
	return true
}