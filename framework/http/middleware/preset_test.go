// Application scenarios:
// - Verify the stability of default, recommended, internal, and admin middleware presets.
// - Guard preset ordering and default-value decisions against accidental drift.
// - Ensure preset-based assembly keeps producing the intended HTTP baseline behavior.
//
// 适用场景：
// - 验证 default、recommended、internal 与 admin 中间件预设的稳定性。
// - 防止预设顺序与默认值决策发生意外漂移。
// - 确保基于预设的装配持续产出预期的 HTTP 基线行为。
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

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
// TestInternalAPIMiddlewareSetDisablesPublicFacingDefaults 验证内部预设会关闭面向公网的默认能力。
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
// TestRecommendedAPIMiddlewareSetAppliesBodyLimit 验证推荐 API 预设中请求体限制保持生效。
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
