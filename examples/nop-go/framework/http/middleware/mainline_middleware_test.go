// Package middleware_test provides unit tests for core HTTP middleware behavior.
//
// 适用场景：
// - 验证 HTTP 主线上的超时、限流和幂等等核心中间件行为。
// - 在中间件重构过程中保持统一响应和资源选择规则稳定。
// - 复用轻量 Gin 桥接助手，在真实请求流中演练 transport 中间件。
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type denyContractLimiter struct{}

func (denyContractLimiter) Allow(context.Context, string) error       { return errors.New("limited") }
func (denyContractLimiter) AllowN(context.Context, string, int) error { return errors.New("limited") }
func (denyContractLimiter) Reserve(context.Context, string) resiliencecontract.Reservation {
	return noopReservation{}
}
func (denyContractLimiter) Wait(context.Context, string) error { return errors.New("limited") }
func (denyContractLimiter) WaitTimeout(context.Context, string, time.Duration) error {
	return errors.New("limited")
}

type noopReservation struct{}

func (noopReservation) OK() bool             { return false }
func (noopReservation) Delay() time.Duration { return 0 }
func (noopReservation) Cancel()              {}
func (noopReservation) CancelAt(time.Time)   {}

type captureLimiter struct {
	resource string
}

func (l *captureLimiter) Allow(_ context.Context, resource string) error {
	l.resource = resource
	return nil
}
func (l *captureLimiter) AllowN(context.Context, string, int) error { return nil }
func (l *captureLimiter) Reserve(context.Context, string) resiliencecontract.Reservation {
	return noopReservation{}
}
func (l *captureLimiter) Wait(context.Context, string) error                       { return nil }
func (l *captureLimiter) WaitTimeout(context.Context, string, time.Duration) error { return nil }

// applyTransportMiddleware mounts transport-level middleware onto a Gin engine for test execution.
//
// applyTransportMiddleware 为测试执行把 transport 层中间件挂到 Gin engine 上。
func applyTransportMiddleware(router *gin.Engine, middleware ...transportcontract.Middleware) {
	handlers := make([]gin.HandlerFunc, 0, len(middleware))
	for _, mw := range middleware {
		if mw == nil {
			continue
		}
		handler := mw(func(c transportcontract.Context) {
			if gc, ok := unwrapGinContext(c); ok {
				gc.Next()
			}
		})
		handlers = append(handlers, func(handler transportcontract.Handler) gin.HandlerFunc {
			return func(c *gin.Context) {
				handler(newContext(c))
			}
		}(handler))
	}
	if len(handlers) > 0 {
		router.Use(handlers...)
	}
}

// TestTimeoutUsesUnifiedResponse verifies that timeout middleware emits the unified timeout response.
//
// TestTimeoutUsesUnifiedResponse 验证超时中间件会输出统一的超时响应。
func TestTimeoutUsesUnifiedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, Timeout(10*time.Millisecond))
	router.GET("/slow", func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond)
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", recorder.Code)
	}

	var resp Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Message != "request timeout" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

// TestTimeoutNoopWhenDisabled verifies that disabled timeout middleware does not alter request flow.
//
// TestTimeoutNoopWhenDisabled 验证禁用超时后中间件不会改变请求流程。
func TestTimeoutNoopWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, Timeout(0))
	router.GET("/fast", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
}

// TestRateLimitUsesUnifiedResponse verifies that the rate-limit middleware returns the unified 429 response.
//
// TestRateLimitUsesUnifiedResponse 验证限流中间件会返回统一的 429 响应。
func TestRateLimitUsesUnifiedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, RateLimit(denyContractLimiter{}, ""))
	router.GET("/demo", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/demo", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", recorder.Code)
	}

	var resp Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Message != "rate limit exceeded" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

// TestRateLimitUsesRoutePathAsResource verifies that route templates are preferred as rate-limit resources.
//
// TestRateLimitUsesRoutePathAsResource 验证限流资源优先使用路由模板。
func TestRateLimitUsesRoutePathAsResource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	limiter := &captureLimiter{}
	applyTransportMiddleware(router, RateLimit(limiter, ""))
	router.GET("/orders/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/orders/42", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if limiter.resource != "/orders/:id" {
		t.Fatalf("expected route path resource, got %q", limiter.resource)
	}
}

// TestRateLimitFallsBackToMethodAndURLPath verifies the fallback resource key when no route template exists.
//
// TestRateLimitFallsBackToMethodAndURLPath 验证在没有路由模板时的资源回退键。
func TestRateLimitFallsBackToMethodAndURLPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	limiter := &captureLimiter{}
	applyTransportMiddleware(router, RateLimit(limiter, ""))
	router.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/raw/path", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if limiter.resource != "GET /raw/path" {
		t.Fatalf("expected method+url path fallback resource, got %q", limiter.resource)
	}
}

// TestIdempotencyCachesStatus verifies that successful write requests are replayed by status code.
//
// TestIdempotencyCachesStatus 验证成功写请求会按状态码被回放。
func TestIdempotencyCachesStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	store := NewMemoryIdempotencyStore()
	hits := 0
	applyTransportMiddleware(router, Idempotency(store, time.Minute))
	router.POST("/orders", func(c *gin.Context) {
		hits++
		c.Status(http.StatusCreated)
	})

	req1 := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req1.Header.Set(IdempotencyKeyHeader, "dup-1")
	recorder1 := httptest.NewRecorder()
	router.ServeHTTP(recorder1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req2.Header.Set(IdempotencyKeyHeader, "dup-1")
	recorder2 := httptest.NewRecorder()
	router.ServeHTTP(recorder2, req2)

	if hits != 1 {
		t.Fatalf("expected handler hit once, got %d", hits)
	}
	if recorder2.Code != http.StatusCreated {
		t.Fatalf("expected replay status 201, got %d", recorder2.Code)
	}
}

// TestIdempotencyReplaysHeadersAndJSONBody verifies that replayed responses preserve headers and JSON body.
//
// TestIdempotencyReplaysHeadersAndJSONBody 验证幂等回放会保留响应头和 JSON 响应体。
func TestIdempotencyReplaysHeadersAndJSONBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	store := NewMemoryIdempotencyStore()
	hits := 0
	applyTransportMiddleware(router, Idempotency(store, time.Minute))
	router.POST("/orders", func(c *gin.Context) {
		hits++
		c.Header("X-Order-Source", "cacheable")
		c.JSON(http.StatusCreated, gin.H{"id": 1, "status": "created"})
	})

	req1 := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req1.Header.Set(IdempotencyKeyHeader, "dup-json")
	recorder1 := httptest.NewRecorder()
	router.ServeHTTP(recorder1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req2.Header.Set(IdempotencyKeyHeader, "dup-json")
	recorder2 := httptest.NewRecorder()
	router.ServeHTTP(recorder2, req2)

	if hits != 1 {
		t.Fatalf("expected handler hit once, got %d", hits)
	}
	if recorder2.Code != http.StatusCreated {
		t.Fatalf("expected replay status 201, got %d", recorder2.Code)
	}
	if recorder2.Header().Get("X-Order-Source") != "cacheable" {
		t.Fatalf("expected replay header, got %q", recorder2.Header().Get("X-Order-Source"))
	}
	if recorder2.Body.String() != recorder1.Body.String() {
		t.Fatalf("expected replay body %q, got %q", recorder1.Body.String(), recorder2.Body.String())
	}
}

// TestIdempotencyDoesNotCacheFailedResponse verifies that failed responses are not cached for replay.
//
// TestIdempotencyDoesNotCacheFailedResponse 验证失败响应不会被缓存回放。
func TestIdempotencyDoesNotCacheFailedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	store := NewMemoryIdempotencyStore()
	hits := 0
	applyTransportMiddleware(router, Idempotency(store, time.Minute))
	router.POST("/orders", func(c *gin.Context) {
		hits++
		c.Status(http.StatusInternalServerError)
	})

	req1 := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req1.Header.Set(IdempotencyKeyHeader, "dup-fail")
	recorder1 := httptest.NewRecorder()
	router.ServeHTTP(recorder1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req2.Header.Set(IdempotencyKeyHeader, "dup-fail")
	recorder2 := httptest.NewRecorder()
	router.ServeHTTP(recorder2, req2)

	// The second request should get 409 Conflict because the key is reserved
	// but not committed (failed response). This prevents duplicate execution
	// of a failed write, which is the safer behavior.
	// 第二次请求应获得 409 Conflict，因为 key 已预留但未提交（失败响应）。
	// 这阻止了对失败写操作的重复执行，是更安全的行为。
	if hits != 1 {
		t.Fatalf("expected failed response not re-executed, got handler hits %d", hits)
	}
	if recorder2.Code != http.StatusConflict {
		t.Fatalf("expected 409 for reserved-but-not-committed key, got %d", recorder2.Code)
	}
}

// TestIdempotencySkipsCacheForUnsupportedMethod verifies that unsupported methods bypass idempotency caching.
//
// TestIdempotencySkipsCacheForUnsupportedMethod 验证不支持的方法会绕过幂等缓存。
func TestIdempotencySkipsCacheForUnsupportedMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	store := NewMemoryIdempotencyStore()
	hits := 0
	applyTransportMiddleware(router, Idempotency(store, time.Minute))
	router.GET("/orders", func(c *gin.Context) {
		hits++
		c.JSON(http.StatusOK, gin.H{"ok": true, "hits": hits})
	})

	req1 := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req1.Header.Set(IdempotencyKeyHeader, "dup-get")
	recorder1 := httptest.NewRecorder()
	router.ServeHTTP(recorder1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req2.Header.Set(IdempotencyKeyHeader, "dup-get")
	recorder2 := httptest.NewRecorder()
	router.ServeHTTP(recorder2, req2)

	if hits != 2 {
		t.Fatalf("expected GET requests not cached, got handler hits %d", hits)
	}
	if reserved, _ := store.Reserve("dup-get", time.Minute); !reserved {
		t.Fatal("expected GET response not stored (key should not exist in store)")
	}
	if recorder1.Body.String() == recorder2.Body.String() {
		t.Fatalf("expected fresh handler execution, got identical bodies %q", recorder1.Body.String())
	}
}
