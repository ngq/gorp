// Application scenarios:
// - Verify the stability of default, recommended, internal, and admin middleware presets.
// - Guard preset ordering and default-value decisions against accidental drift.
// - Ensure preset-based assembly keeps producing the intended HTTP baseline behavior.
//
// 适用场景：
// - 验证 default、recommended、internal、admin 中间件预设的稳定性。
// - 防止预设顺序与默认值决策发生意外漂移。
// - 确保基于预设的装配持续产出预期的 HTTP 基线行为。
package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

type denyCircuitBreaker struct{}

func (denyCircuitBreaker) Allow(context.Context, string) error          { return errors.New("blocked") }
func (denyCircuitBreaker) RecordSuccess(context.Context, string)        {}
func (denyCircuitBreaker) RecordFailure(context.Context, string, error) {}
func (denyCircuitBreaker) Do(ctx context.Context, resource string, fn func() error) error {
	if fn == nil {
		return nil
	}
	return fn()
}
func (denyCircuitBreaker) State(context.Context, string) resiliencecontract.CircuitBreakerState {
	return resiliencecontract.CircuitBreakerStateClosed
}

type denyCircuitBreakerWithAllowError struct{}

func (denyCircuitBreakerWithAllowError) Allow(context.Context, string) error {
	return errors.New("open")
}
func (denyCircuitBreakerWithAllowError) RecordSuccess(context.Context, string)        {}
func (denyCircuitBreakerWithAllowError) RecordFailure(context.Context, string, error) {}
func (denyCircuitBreakerWithAllowError) Do(ctx context.Context, resource string, fn func() error) error {
	if fn == nil {
		return nil
	}
	return fn()
}
func (denyCircuitBreakerWithAllowError) State(context.Context, string) resiliencecontract.CircuitBreakerState {
	return resiliencecontract.CircuitBreakerStateOpen
}

// TestDefaultMiddlewareSetStableSize verifies that the default middleware preset keeps a stable cardinality.
//
// TestDefaultMiddlewareSetStableSize 验证默认中间件预设保持稳定的数量。
func TestDefaultMiddlewareSetStableSize(t *testing.T) {
	set := DefaultMiddlewareSet(nil)
	if len(set) != 3 {
		t.Fatalf("expected 3 default middleware entries, got %d", len(set))
	}
}

// TestRecommendedAPIMiddlewareSetAppliesDefaultHeaders verifies the default public API preset behavior.
//
// TestRecommendedAPIMiddlewareSetAppliesDefaultHeaders 验证默认对外 API 预设行为。
func TestRecommendedAPIMiddlewareSetAppliesDefaultHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, RecommendedAPIMiddlewareSet(nil, RecommendedMiddlewareOptions{})...)
	router.GET("/orders/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/orders/42?lang=en-US", nil)
	req.Header.Set("Origin", "https://example.com")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id header")
	}
	if recorder.Header().Get("X-Trace-Id") == "" {
		t.Fatal("expected X-Trace-Id header")
	}
	if recorder.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("expected X-Frame-Options DENY, got %q", recorder.Header().Get("X-Frame-Options"))
	}
	if recorder.Header().Get("Content-Language") != "en" {
		t.Fatalf("expected Content-Language en, got %q", recorder.Header().Get("Content-Language"))
	}
	if recorder.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected CORS disabled by default, got %q", recorder.Header().Get("Access-Control-Allow-Origin"))
	}
}

// TestRecommendedAPIMiddlewareSetEnablesConfiguredCORS verifies explicit CORS enablement in the recommended API preset.
//
// TestRecommendedAPIMiddlewareSetEnablesConfiguredCORS 验证推荐 API 预设中显式启用 CORS 的行为。
func TestRecommendedAPIMiddlewareSetEnablesConfiguredCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	options := RecommendedMiddlewareOptions{
		CORS: func() *CORSOptions {
			opts := DefaultCORSOptions()
			return &opts
		}(),
	}
	applyTransportMiddleware(router, RecommendedAPIMiddlewareSet(nil, options)...)
	router.GET("/orders/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/orders/42", nil)
	req.Header.Set("Origin", "https://example.com")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected CORS wildcard origin, got %q", recorder.Header().Get("Access-Control-Allow-Origin"))
	}
}

// TestInternalAPIMiddlewareSetDisablesPublicFacingDefaults verifies that internal presets disable public-facing defaults.
//
// TestInternalAPIMiddlewareSetDisablesPublicFacingDefaults 验证内网预设会关闭面向公网的默认能力。
func TestInternalAPIMiddlewareSetDisablesPublicFacingDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, InternalAPIMiddlewareSet(nil, InternalMiddlewareOptions{})...)
	router.GET("/internal/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/internal/ping?lang=en", nil)
	req.Header.Set("Origin", "https://example.com")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id header")
	}
	if recorder.Header().Get("Content-Language") != "" {
		t.Fatalf("expected locale middleware disabled, got %q", recorder.Header().Get("Content-Language"))
	}
	if recorder.Header().Get("X-Frame-Options") != "" {
		t.Fatalf("expected security headers disabled, got %q", recorder.Header().Get("X-Frame-Options"))
	}
	if recorder.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected CORS disabled, got %q", recorder.Header().Get("Access-Control-Allow-Origin"))
	}
}

// TestAdminAPIMiddlewareSetRequiresAuthorizationByDefault verifies admin preset authorization defaults.
//
// TestAdminAPIMiddlewareSetRequiresAuthorizationByDefault 验证管理预设的默认鉴权行为。
func TestAdminAPIMiddlewareSetRequiresAuthorizationByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, AdminAPIMiddlewareSet(nil, AdminMiddlewareOptions{})...)
	router.GET("/admin/panel", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/panel", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

// TestRecommendedAPIMiddlewareSetAppliesBodyLimit verifies that body limit remains active in the recommended API preset.
//
// TestRecommendedAPIMiddlewareSetAppliesBodyLimit 验证推荐 API 预设中的请求体限制保持生效。
func TestRecommendedAPIMiddlewareSetAppliesBodyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	options := RecommendedMiddlewareOptions{
		BodyLimitBytes:         4,
		DisableLocale:          true,
		DisableSecurityHeaders: true,
		EnableMetrics:          false,
		CORS:                   nil,
	}
	applyTransportMiddleware(router, RecommendedAPIMiddlewareSet(nil, options)...)
	router.POST("/upload", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/upload", http.NoBody)
	req.ContentLength = 8
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", recorder.Code)
	}
}

// TestDefaultHTTPServiceGovernanceSetIncludesServiceProtection verifies the default HTTP governance preset order.
//
// TestDefaultHTTPServiceGovernanceSetIncludesServiceProtection 验证默认 HTTP 服务治理预设包含服务保护链路。
func TestDefaultHTTPServiceGovernanceSetIncludesServiceProtection(t *testing.T) {
	set := DefaultHTTPServiceGovernanceSet(nil, DefaultHTTPServiceGovernanceOptions{
		RateLimiter:   denyContractLimiter{},
		MaxConcurrent: 16,
	})
	if len(set) != 10 {
		t.Fatalf("expected 10 governance middleware entries, got %d", len(set))
	}
}

func TestDefaultHTTPServiceGovernanceDefaultsRemainStable(t *testing.T) {
	defaults := DefaultHTTPServiceGovernanceDefaults()
	if defaults.MaxConcurrent != 0 {
		t.Fatalf("expected default MaxConcurrent 0, got %d", defaults.MaxConcurrent)
	}
	if defaults.API.Timeout != 15*time.Second {
		t.Fatalf("expected default timeout 15s, got %s", defaults.API.Timeout)
	}
	if defaults.API.BodyLimitBytes != 2<<20 {
		t.Fatalf("expected default body limit 2097152, got %d", defaults.API.BodyLimitBytes)
	}
	if !defaults.API.EnableMetrics {
		t.Fatal("expected metrics enabled by default")
	}
	if defaults.API.EnableCompression {
		t.Fatal("expected compression disabled by default")
	}
	if defaults.API.CORS != nil {
		t.Fatalf("expected CORS disabled by default, got %#v", defaults.API.CORS)
	}
	if defaults.API.SecurityHeaders == nil {
		t.Fatal("expected security headers defaults to be present")
	}
	if defaults.API.Locale == nil {
		t.Fatal("expected locale defaults to be present")
	}
}

func TestDefaultHTTPServiceGovernanceSetStableDefaultCardinality(t *testing.T) {
	set := DefaultHTTPServiceGovernanceSet(nil, DefaultHTTPServiceGovernanceOptions{})
	if len(set) != 8 {
		t.Fatalf("expected 8 default governance middleware entries, got %d", len(set))
	}
}

func TestDefaultHTTPServiceGovernanceSetAddsOptionalStagesInStableOrder(t *testing.T) {
	set := DefaultHTTPServiceGovernanceSet(nil, DefaultHTTPServiceGovernanceOptions{
		API: RecommendedMiddlewareOptions{
			CORS: func() *CORSOptions {
				opts := DefaultCORSOptions()
				return &opts
			}(),
			EnableCompression: true,
		},
		RateLimiter:    denyContractLimiter{},
		CircuitBreaker: denyCircuitBreaker{},
		MaxConcurrent:  8,
	})
	if len(set) != 12 {
		t.Fatalf("expected 12 governance middleware entries, got %d", len(set))
	}
}

func TestDefaultHTTPServiceGovernanceSetRateLimitPrecedesBodyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	called := false
	applyTransportMiddleware(router, DefaultHTTPServiceGovernanceSet(nil, DefaultHTTPServiceGovernanceOptions{
		API: RecommendedMiddlewareOptions{
			BodyLimitBytes:         4,
			DisableLocale:          true,
			DisableSecurityHeaders: true,
			EnableMetrics:          false,
		},
		RateLimiter: denyContractLimiter{},
	})...)
	router.POST("/upload", func(c *gin.Context) {
		called = true
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/upload", http.NoBody)
	req.ContentLength = 8
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 before body limit, got %d", recorder.Code)
	}
	if called {
		t.Fatal("expected handler not to be called")
	}
}

func TestDefaultHTTPServiceGovernanceSetCircuitBreakerPrecedesBodyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	called := false
	applyTransportMiddleware(router, DefaultHTTPServiceGovernanceSet(nil, DefaultHTTPServiceGovernanceOptions{
		API: RecommendedMiddlewareOptions{
			BodyLimitBytes:         4,
			DisableLocale:          true,
			DisableSecurityHeaders: true,
			EnableMetrics:          false,
		},
		CircuitBreaker: denyCircuitBreakerWithAllowError{},
	})...)
	router.POST("/upload", func(c *gin.Context) {
		called = true
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/upload", http.NoBody)
	req.ContentLength = 8
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 before body limit, got %d", recorder.Code)
	}
	if called {
		t.Fatal("expected handler not to be called")
	}
}
