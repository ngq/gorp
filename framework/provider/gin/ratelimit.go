package gin

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type RateLimiter interface {
	Allow(key string) bool
}

func RateLimit(limiter resiliencecontract.RateLimiter, resource string) transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
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
			if err := limiter.Allow(c.Context(), target); err != nil {
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

func IPKeyFunc(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	return c.ClientIP()
}

type TokenBucketLimiter struct {
	rate     float64
	burst    int
	tokens   float64
	lastTime time.Time
	mu       sync.Mutex
}

func NewTokenBucketLimiter(rate float64, burst int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:     rate,
		burst:    burst,
		tokens:   float64(burst),
		lastTime: time.Now(),
	}
}

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

type SlidingWindowLimiter struct {
	limit  int
	window time.Duration
	counts map[string]*windowRecord
	mu     sync.Mutex
}

type windowRecord struct {
	timestamps []time.Time
}

func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		limit:  limit,
		window: window,
		counts: make(map[string]*windowRecord),
	}
}

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
	validIdx := 0
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

type FixedWindowLimiter struct {
	limit   int
	window  time.Duration
	counts  map[string]int
	windows map[string]time.Time
	mu      sync.Mutex
}

func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		limit:   limit,
		window:  window,
		counts:  make(map[string]int),
		windows: make(map[string]time.Time),
	}
}

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
