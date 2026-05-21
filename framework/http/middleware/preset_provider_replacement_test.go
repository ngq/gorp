// Package middleware_test provides unit tests for HTTP middleware preset stability.
//
// 适用场景：
// - 验证 Governance Preset Provider 替换与运行时行为。
// - 确保替换 provider 后 HTTP 响应行为按预期改变。
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestGovernancePresetProviderReplacementChangesRuntimeBehavior verifies that swapping a provider
// in the governance preset actually changes the observed HTTP response behavior.
// This addresses verification requirement #5: "Provider 替换后的最终主线行为".
//
// TestGovernancePresetProviderReplacementChangesRuntimeBehavior 验证在治理预设中替换 provider
// 会实际改变 HTTP 响应行为。对应验收要求 #5："Provider 替换后的最终主线行为"。
func TestGovernancePresetProviderReplacementChangesRuntimeBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 场景 1：使用开放熔断器的治理集 → 请求被 503 拒绝
	openBreakerRouter := NewTestEngine()
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
	allowBreakerRouter := NewTestEngine()
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
	noBreakerRouter := NewTestEngine()
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
