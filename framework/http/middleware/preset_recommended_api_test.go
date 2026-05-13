// Package middleware_test provides unit tests for HTTP middleware preset stability.
//
// 适用场景：
// - 验证 Recommended API 中间件预设的行为。
// - 确保 CORS、安全头、请求追踪等功能按预期工作。
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

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
