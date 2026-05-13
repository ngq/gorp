// Package middleware_test provides unit tests for HTTP middleware preset stability.
//
// 适用场景：
// - 验证 Internal / Admin API 中间件预设的行为。
// - 确保内网预设关闭面向公网的默认能力。
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

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
