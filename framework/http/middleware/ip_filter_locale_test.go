// Package middleware_test provides unit tests for IP filtering and locale middleware.
//
// 适用场景：
// - IP 白名单与黑名单
// - 语言协商与本地化
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestIPAllowlistAndDenylist verifies allowlist pass-through and denylist blocking behavior.
//
// TestIPAllowlistAndDenylist 验证 allowlist 放行与 denylist 拦截行为。
func TestIPAllowlistAndDenylist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	allowRouter := gin.New()
	applyTransportMiddleware(allowRouter, IPAllowlist("10.0.0.0/8"))
	allowRouter.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	allowReq := httptest.NewRequest(http.MethodGet, "/admin", nil)
	allowReq.Header.Set("X-Forwarded-For", "10.1.2.3")
	allowRecorder := httptest.NewRecorder()
	allowRouter.ServeHTTP(allowRecorder, allowReq)

	if allowRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected allowlist request 204, got %d", allowRecorder.Code)
	}

	denyRouter := gin.New()
	applyTransportMiddleware(denyRouter, IPDenylist("10.0.0.0/8"))
	denyRouter.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	denyReq := httptest.NewRequest(http.MethodGet, "/admin", nil)
	denyReq.Header.Set("X-Forwarded-For", "10.1.2.3")
	denyRecorder := httptest.NewRecorder()
	denyRouter.ServeHTTP(denyRecorder, denyReq)

	if denyRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected denylist request 403, got %d", denyRecorder.Code)
	}
}

// TestLocaleUsesQueryThenHeaderThenDefault verifies locale negotiation order and response header output.
//
// TestLocaleUsesQueryThenHeaderThenDefault 验证语言协商顺序与响应头输出。
func TestLocaleUsesQueryThenHeaderThenDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queryRouter := gin.New()
	applyTransportMiddleware(queryRouter, Locale(DefaultLocaleOptions()))
	queryRouter.GET("/locale", func(c *gin.Context) {
		c.String(http.StatusOK, GetLocale(c))
	})

	queryReq := httptest.NewRequest(http.MethodGet, "/locale?lang=en-US", nil)
	queryRecorder := httptest.NewRecorder()
	queryRouter.ServeHTTP(queryRecorder, queryReq)

	if queryRecorder.Body.String() != "en" {
		t.Fatalf("expected query locale en, got %q", queryRecorder.Body.String())
	}
	if queryRecorder.Header().Get("Content-Language") != "en" {
		t.Fatalf("expected Content-Language en, got %q", queryRecorder.Header().Get("Content-Language"))
	}

	headerReq := httptest.NewRequest(http.MethodGet, "/locale", nil)
	headerReq.Header.Set("Accept-Language", "en-GB,en;q=0.8,zh;q=0.7")
	headerRecorder := httptest.NewRecorder()
	queryRouter.ServeHTTP(headerRecorder, headerReq)

	if headerRecorder.Body.String() != "en" {
		t.Fatalf("expected header locale en, got %q", headerRecorder.Body.String())
	}

	defaultReq := httptest.NewRequest(http.MethodGet, "/locale", nil)
	defaultRecorder := httptest.NewRecorder()
	queryRouter.ServeHTTP(defaultRecorder, defaultReq)

	if defaultRecorder.Body.String() != "zh" {
		t.Fatalf("expected default locale zh, got %q", defaultRecorder.Body.String())
	}
}
