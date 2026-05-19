// Application scenarios:
// - Limit instantaneous in-flight requests on hot endpoints.
// - Shed excess traffic early when the service is already overloaded.
// - Build a lightweight overload-protection baseline without external dependencies.
//
// 适用场景：
// - 限制热点接口的瞬时在途请求数。
// - 在服务已经过载时尽早丢弃多余流量。
// - 在不依赖外部组件的前提下建立轻量级过载保护基线。
package middleware

import (
	"context"
	"net/http"
	"time"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// ConcurrencyLimiter is a lightweight semaphore used by HTTP protection middleware.
//
// ConcurrencyLimiter 是 HTTP 保护中间件使用的轻量级信号量。
type ConcurrencyLimiter struct {
	tokens chan struct{}
}

// NewConcurrencyLimiter creates a limiter with the given maximum concurrency.
//
// NewConcurrencyLimiter 按给定最大并发数创建一个限制器。
func NewConcurrencyLimiter(maxConcurrent int) *ConcurrencyLimiter {
	if maxConcurrent <= 0 {
		return nil
	}
	return &ConcurrencyLimiter{
		tokens: make(chan struct{}, maxConcurrent),
	}
}

// tryAcquire attempts to acquire a concurrency slot immediately.
//
// tryAcquire 尝试立即获取一个并发槽位。
func (l *ConcurrencyLimiter) tryAcquire() bool {
	if l == nil {
		return true
	}
	select {
	case l.tokens <- struct{}{}:
		return true
	default:
		return false
	}
}

// acquire waits up to the timeout for a concurrency slot.
//
// acquire 在超时时间内等待一个并发槽位。
func (l *ConcurrencyLimiter) acquire(timeout time.Duration) bool {
	if l == nil {
		return true
	}
	if timeout <= 0 {
		return l.tryAcquire()
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case l.tokens <- struct{}{}:
			return true
		case <-timer.C:
			return false
		}
	}
}

// release returns a previously acquired concurrency slot.
//
// release 归还一个之前占用的并发槽位。
func (l *ConcurrencyLimiter) release() {
	if l == nil {
		return
	}
	select {
	case <-l.tokens:
	default:
	}
}

// LoadShedding rejects the request immediately when the concurrency limit is full.
// 优先使用容器中的 LoadShedder 契约实现；若容器不可用则回退到本地 ConcurrencyLimiter。
//
// LoadShedding 在并发已满时立即拒绝请求。
// 优先使用容器中的 LoadShedder 契约实现；若容器不可用则回退到本地 ConcurrencyLimiter。
func LoadShedding(maxConcurrent int) transportcontract.Middleware {
	limiter := NewConcurrencyLimiter(maxConcurrent)
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			// 优先尝试从容器中获取 LoadShedder 契约实现
			if shedder := loadShedderFromContext(c); shedder != nil {
				if err := shedder.Allow(context.Background(), "http"); err != nil {
					respondServiceBusy(c, "server is busy")
					return
				}
				defer shedder.Done(context.Background(), "http", nil)

				if next != nil {
					next(c)
				}
				return
			}

			// 回退路径：使用本地 ConcurrencyLimiter
			if limiter == nil {
				if next != nil {
					next(c)
				}
				return
			}
			if !limiter.tryAcquire() {
				respondServiceBusy(c, "server is busy")
				return
			}
			defer limiter.release()

			if next != nil {
				next(c)
			}
		}
	}
}

// loadShedderFromContext 尝试从请求上下文中获取容器，再从容器中解析 LoadShedder 契约。
// 如果容器不可用或 LoadShedder 未注册，返回 nil。
func loadShedderFromContext(c transportcontract.Context) resiliencecontract.LoadShedder {
	if c == nil {
		return nil
	}
	containerAny, ok := supportcontract.FromContainerContext(c)
	if !ok {
		return nil
	}
	container, ok := containerAny.(runtimecontract.Container)
	if !ok || container == nil {
		return nil
	}
	if !container.IsBind(resiliencecontract.LoadShedderKey) {
		return nil
	}
	shedderAny, err := container.Make(resiliencecontract.LoadShedderKey)
	if err != nil {
		return nil
	}
	shedder, ok := shedderAny.(resiliencecontract.LoadShedder)
	if !ok {
		return nil
	}
	return shedder
}

// ConcurrencyLimit waits for a free slot until timeout, then rejects the request.
//
// ConcurrencyLimit 在超时前等待空闲槽位，超时后拒绝请求。
func ConcurrencyLimit(maxConcurrent int, timeout time.Duration) transportcontract.Middleware {
	limiter := NewConcurrencyLimiter(maxConcurrent)
	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			if limiter == nil {
				if next != nil {
					next(c)
				}
				return
			}
			if !limiter.acquire(timeout) {
				respondServiceBusy(c, "server is busy")
				return
			}
			defer limiter.release()

			if next != nil {
				next(c)
			}
		}
	}
}

// respondServiceBusy writes the unified overload response.
//
// respondServiceBusy 输出统一的服务繁忙响应。
func respondServiceBusy(c transportcontract.Context, message string) {
	if gc, ok := unwrapGinContext(c); ok {
		writeGinResponseHeaders(gc)
		resp := Response{
			Code:    CodeServiceUnavailable,
			Message: message,
		}
		gc.JSON(http.StatusServiceUnavailable, resp)
		gc.Abort()
		return
	}

	c.JSON(http.StatusServiceUnavailable, map[string]any{
		"code":    CodeServiceUnavailable,
		"message": message,
	})
}
