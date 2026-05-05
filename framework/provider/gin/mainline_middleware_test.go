package gin

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

func TestTimeoutMainlineMiddlewareUsesUnifiedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(adaptMiddleware(Timeout(10 * time.Millisecond)))
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond)
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Message != "request timeout" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

func TestTimeoutMainlineMiddlewareNoopWhenTimeoutDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(adaptMiddleware(Timeout(0)))
	r.GET("/fast", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestRateLimitMainlineMiddlewareUsesUnifiedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(adaptMiddleware(RateLimit(denyContractLimiter{}, "")))
	r.GET("/demo", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/demo", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Message != "rate limit exceeded" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

func TestRateLimitMainlineMiddlewareUsesRoutePathAsResource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	limiter := &captureLimiter{}
	r.Use(adaptMiddleware(RateLimit(limiter, "")))
	r.GET("/orders/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/orders/42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if limiter.resource != "/orders/:id" {
		t.Fatalf("expected route path resource, got %q", limiter.resource)
	}
}

func TestRateLimitMainlineMiddlewareFallsBackToMethodAndURLPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	limiter := &captureLimiter{}
	r.Use(adaptMiddleware(RateLimit(limiter, "")))
	r.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/raw/path", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if limiter.resource != "GET /raw/path" {
		t.Fatalf("expected method+url path fallback resource, got %q", limiter.resource)
	}
}

func TestIdempotencyMainlineMiddlewareCachesStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	store := NewMemoryIdempotencyStore()
	hits := 0
	r.Use(adaptMiddleware(Idempotency(store, time.Minute)))
	r.POST("/orders", func(c *gin.Context) {
		hits++
		c.Status(http.StatusCreated)
	})

	req1 := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req1.Header.Set(IdempotencyKeyHeader, "dup-1")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req2.Header.Set(IdempotencyKeyHeader, "dup-1")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if hits != 1 {
		t.Fatalf("expected handler hit once, got %d", hits)
	}
	if w2.Code != http.StatusCreated {
		t.Fatalf("expected replay status 201, got %d", w2.Code)
	}
}

func TestIdempotencyMainlineMiddlewareReplaysHeadersAndJSONBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	store := NewMemoryIdempotencyStore()
	hits := 0
	r.Use(adaptMiddleware(Idempotency(store, time.Minute)))
	r.POST("/orders", func(c *gin.Context) {
		hits++
		c.Header("X-Order-Source", "cacheable")
		c.JSON(http.StatusCreated, gin.H{"id": 1, "status": "created"})
	})

	req1 := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req1.Header.Set(IdempotencyKeyHeader, "dup-json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req2.Header.Set(IdempotencyKeyHeader, "dup-json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if hits != 1 {
		t.Fatalf("expected handler hit once, got %d", hits)
	}
	if w2.Code != http.StatusCreated {
		t.Fatalf("expected replay status 201, got %d", w2.Code)
	}
	if w2.Header().Get("X-Order-Source") != "cacheable" {
		t.Fatalf("expected replay header, got %q", w2.Header().Get("X-Order-Source"))
	}
	if w2.Body.String() != w1.Body.String() {
		t.Fatalf("expected replay body %q, got %q", w1.Body.String(), w2.Body.String())
	}
}

func TestIdempotencyMainlineMiddlewareDoesNotCacheFailedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	store := NewMemoryIdempotencyStore()
	hits := 0
	r.Use(adaptMiddleware(Idempotency(store, time.Minute)))
	r.POST("/orders", func(c *gin.Context) {
		hits++
		c.Status(http.StatusInternalServerError)
	})

	req1 := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req1.Header.Set(IdempotencyKeyHeader, "dup-fail")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req2.Header.Set(IdempotencyKeyHeader, "dup-fail")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if hits != 2 {
		t.Fatalf("expected failed responses not cached, got handler hits %d", hits)
	}
	if exists, _ := store.Check("dup-fail"); exists {
		t.Fatal("expected failed response not stored")
	}
	if w2.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w2.Code)
	}
}

func TestIdempotencyMainlineMiddlewareSkipsCacheForUnsupportedMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	store := NewMemoryIdempotencyStore()
	hits := 0
	r.Use(adaptMiddleware(Idempotency(store, time.Minute)))
	r.GET("/orders", func(c *gin.Context) {
		hits++
		c.JSON(http.StatusOK, gin.H{"ok": true, "hits": hits})
	})

	req1 := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req1.Header.Set(IdempotencyKeyHeader, "dup-get")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req2.Header.Set(IdempotencyKeyHeader, "dup-get")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if hits != 2 {
		t.Fatalf("expected GET requests not cached, got handler hits %d", hits)
	}
	if exists, _ := store.Check("dup-get"); exists {
		t.Fatal("expected GET response not stored")
	}
	if w1.Body.String() == w2.Body.String() {
		t.Fatalf("expected fresh handler execution, got identical bodies %q", w1.Body.String())
	}
}
