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
	"runtime"
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

// allowCircuitBreaker 允许所有请求通过的熔断器 stub。
type allowCircuitBreaker struct{}

func (allowCircuitBreaker) Allow(context.Context, string) error          { return nil }
func (allowCircuitBreaker) RecordSuccess(context.Context, string)        {}
func (allowCircuitBreaker) RecordFailure(context.Context, string, error) {}
func (allowCircuitBreaker) Do(ctx context.Context, resource string, fn func() error) error {
	if fn == nil {
		return nil
	}
	return fn()
}
func (allowCircuitBreaker) State(context.Context, string) resiliencecontract.CircuitBreakerState {
	return resiliencecontract.CircuitBreakerStateClosed
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
	expectedMaxConcurrent := runtime.GOMAXPROCS(0) * 100
	if defaults.MaxConcurrent != expectedMaxConcurrent {
		t.Fatalf("expected default MaxConcurrent %d (CPU核数×100), got %d", expectedMaxConcurrent, defaults.MaxConcurrent)
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
	// 默认集包含：request_identity, logging, recovery (3) + security_headers (1) + timeout (1) + loadshedding (1) + body_limit (1) + locale (1) + metrics (1) = 9
	if len(set) != 9 {
		t.Fatalf("expected 9 default governance middleware entries (including loadshedding), got %d", len(set))
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

// TestDefaultHTTPServiceGovernanceOrderStable verifies the HTTP server governance order remains stable.
// This is the HTTP-side counterpart of rpc/governance.TestDefaultClientPresetOrderStable.
//
// TestDefaultHTTPServiceGovernanceOrderStable 验证 HTTP 服务端治理顺序保持稳定。
// 这是 rpc/governance.TestDefaultClientPresetOrderStable 的 HTTP 侧对称测试。
func TestDefaultHTTPServiceGovernanceOrderStable(t *testing.T) {
	order := DefaultHTTPServiceGovernanceOrder()
	if len(order) != 13 {
		t.Fatalf("expected 13 HTTP governance stages, got %d", len(order))
	}
	expected := []string{
		"request_identity",
		"logging",
		"recovery",
		"cors",
		"security_headers",
		"timeout",
		"load_shedding",
		"rate_limit",
		"circuit_breaker",
		"body_limit",
		"locale",
		"metrics",
		"compression",
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("expected stable order %v, got %v", expected, order)
		}
	}
	// 入站链路最先是 request_identity，最后是 compression
	if order[0] != "request_identity" || order[len(order)-1] != "compression" {
		t.Fatalf("unexpected order %v", order)
	}
	// 服务保护类必须在 body_limit 之前
	protectionStages := []string{"load_shedding", "rate_limit", "circuit_breaker"}
	bodyLimitIdx := indexOfStr(order, "body_limit")
	for _, stage := range protectionStages {
		idx := indexOfStr(order, stage)
		if idx >= bodyLimitIdx {
			t.Fatalf("expected %s before body_limit, but %s at %d and body_limit at %d", stage, stage, idx, bodyLimitIdx)
		}
	}
	// recovery 必须在 timeout 之前，确保 panic 能被捕获后再走超时逻辑
	recoveryIdx := indexOfStr(order, "recovery")
	timeoutIdx := indexOfStr(order, "timeout")
	if recoveryIdx >= timeoutIdx {
		t.Fatalf("expected recovery before timeout, but recovery at %d and timeout at %d", recoveryIdx, timeoutIdx)
	}
}

// TestDefaultHTTPServiceGovernanceSetFullChainMatchesOrder verifies that when all optional stages are enabled,
// the actual middleware slice has one entry per order slot.
//
// TestDefaultHTTPServiceGovernanceSetFullChainMatchesOrder 验证当所有可选阶段启用时，
// 实际中间件切片的长度与正式顺序列表一致。
func TestDefaultHTTPServiceGovernanceSetFullChainMatchesOrder(t *testing.T) {
	order := DefaultHTTPServiceGovernanceOrder()
	set := DefaultHTTPServiceGovernanceSet(nil, DefaultHTTPServiceGovernanceOptions{
		API: RecommendedMiddlewareOptions{
			CORS: func() *CORSOptions {
				opts := DefaultCORSOptions()
				return &opts
			}(),
			EnableMetrics:     true,
			EnableCompression: true,
		},
		RateLimiter:    denyContractLimiter{},
		CircuitBreaker: denyCircuitBreaker{},
		MaxConcurrent:  8,
	})
	// 全量启用后，实际中间件数量应与正式顺序列表完全一致
	if len(set) != len(order) {
		t.Fatalf("expected full governance set size %d (matching order), got %d", len(order), len(set))
	}
}

// TestDefaultHTTPServiceGovernanceSetDefaultChainMatchesActiveStages verifies that the default governance set
// (without optional stages) only includes stages that are active by default.
//
// TestDefaultHTTPServiceGovernanceSetDefaultChainMatchesActiveStages 验证默认治理集
// （不含可选阶段）只包含默认启用的阶段。
func TestDefaultHTTPServiceGovernanceSetDefaultChainMatchesActiveStages(t *testing.T) {
	order := DefaultHTTPServiceGovernanceOrder()
	set := DefaultHTTPServiceGovernanceSet(nil, DefaultHTTPServiceGovernanceOptions{})
	// 默认集包含始终启用的阶段 + 默认启用的可选阶段
	// 默认启用：request_identity, logging, recovery, security_headers, timeout, loadshedding, body_limit, locale, metrics = 9
	// 默认不启用：cors, rate_limit, circuit_breaker, compression
	if len(set) != 9 {
		t.Fatalf("expected default governance set size 9 (including loadshedding), got %d", len(set))
	}
	// 确保默认启用集的数量不超过正式顺序列表
	if len(set) > len(order) {
		t.Fatalf("default set size %d exceeds order size %d", len(set), len(order))
	}
}

func indexOfStr(slice []string, target string) int {
	for i, s := range slice {
		if s == target {
			return i
		}
	}
	return -1
}

// TestGovernancePresetProviderReplacementChangesRuntimeBehavior verifies that swapping a provider
// in the governance preset actually changes the observed HTTP response behavior.
// This addresses verification requirement #5: "Provider 替换后的最终主线行为".
//
// TestGovernancePresetProviderReplacementChangesRuntimeBehavior 验证在治理预设中替换 provider
// 会实际改变 HTTP 响应行为。对应验收要求 #5："Provider 替换后的最终主线行为"。
func TestGovernancePresetProviderReplacementChangesRuntimeBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 场景 1：使用开放熔断器的治理集 → 请求被 503 拒绝
	openBreakerRouter := gin.New()
	applyTransportMiddleware(openBreakerRouter, DefaultHTTPServiceGovernanceSet(nil, DefaultHTTPServiceGovernanceOptions{
		API: RecommendedMiddlewareOptions{
			DisableLocale:          true,
			DisableSecurityHeaders: true,
			EnableMetrics:          false,
		},
		CircuitBreaker: denyCircuitBreakerWithAllowError{},
	})...)
	openBreakerRouter.GET("/api/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	recorder := httptest.NewRecorder()
	openBreakerRouter.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 with open breaker, got %d", recorder.Code)
	}

	// 场景 2：使用允许通过的熔断器 → 请求正常 200
	allowBreakerRouter := gin.New()
	applyTransportMiddleware(allowBreakerRouter, DefaultHTTPServiceGovernanceSet(nil, DefaultHTTPServiceGovernanceOptions{
		API: RecommendedMiddlewareOptions{
			DisableLocale:          true,
			DisableSecurityHeaders: true,
			EnableMetrics:          false,
		},
		CircuitBreaker: allowCircuitBreaker{},
	})...)
	allowBreakerRouter.GET("/api/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	recorder2 := httptest.NewRecorder()
	allowBreakerRouter.ServeHTTP(recorder2, req2)
	if recorder2.Code != http.StatusOK {
		t.Fatalf("expected 200 with allowing breaker, got %d", recorder2.Code)
	}

	// 场景 3：不提供熔断器（等效于替换为 noop）→ 请求正常 200
	noBreakerRouter := gin.New()
	applyTransportMiddleware(noBreakerRouter, DefaultHTTPServiceGovernanceSet(nil, DefaultHTTPServiceGovernanceOptions{
		API: RecommendedMiddlewareOptions{
			DisableLocale:          true,
			DisableSecurityHeaders: true,
			EnableMetrics:          false,
		},
	})...)
	noBreakerRouter.GET("/api/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req3 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	recorder3 := httptest.NewRecorder()
	noBreakerRouter.ServeHTTP(recorder3, req3)
	if recorder3.Code != http.StatusOK {
		t.Fatalf("expected 200 with no breaker, got %d", recorder3.Code)
	}
}
