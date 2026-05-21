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
// Uses RemoteAddr directly since X-Forwarded-For is not trusted by default.
//
// TestIPAllowlistAndDenylist 验证 allowlist 放行与 denylist 拦截行为。
// 直接使用 RemoteAddr，因为 X-Forwarded-For 默认不被信任。
func TestIPAllowlistAndDenylist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	allowRouter := NewTestEngine()
	applyTransportMiddleware(allowRouter, IPAllowlist("10.0.0.0/8"))
	allowRouter.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	allowReq := httptest.NewRequest(http.MethodGet, "/admin", nil)
	allowReq.RemoteAddr = "10.1.2.3:12345"
	allowRecorder := httptest.NewRecorder()
	allowRouter.ServeHTTP(allowRecorder, allowReq)

	if allowRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected allowlist request 204, got %d", allowRecorder.Code)
	}

	denyRouter := NewTestEngine()
	applyTransportMiddleware(denyRouter, IPDenylist("10.0.0.0/8"))
	denyRouter.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	denyReq := httptest.NewRequest(http.MethodGet, "/admin", nil)
	denyReq.RemoteAddr = "10.1.2.3:12345"
	denyRecorder := httptest.NewRecorder()
	denyRouter.ServeHTTP(denyRecorder, denyReq)

	if denyRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected denylist request 403, got %d", denyRecorder.Code)
	}
}

// TestIPAllowlistWithTrustedProxies verifies X-Forwarded-For is used when trusted proxies are configured.
//
// TestIPAllowlistWithTrustedProxies 验证配置可信代理后 X-Forwarded-For 被正确使用。
func TestIPAllowlistWithTrustedProxies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Configure trusted proxy
	SetTrustedProxies([]string{"127.0.0.1"})
	defer SetTrustedProxies(nil)

	allowRouter := NewTestEngine()
	applyTransportMiddleware(allowRouter, IPAllowlist("10.0.0.0/8"))
	allowRouter.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	allowReq := httptest.NewRequest(http.MethodGet, "/admin", nil)
	allowReq.RemoteAddr = "127.0.0.1:12345"
	allowReq.Header.Set("X-Forwarded-For", "10.1.2.3")
	allowRecorder := httptest.NewRecorder()
	allowRouter.ServeHTTP(allowRecorder, allowReq)

	if allowRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected allowlist request with trusted proxy 204, got %d", allowRecorder.Code)
	}
}

// TestIPAllowlistRejectsSpoofedXFF verifies X-Forwarded-For is ignored when no trusted proxy.
//
// TestIPAllowlistRejectsSpoofedXFF 验证无可信代理时 X-Forwarded-For 被忽略。
func TestIPAllowlistRejectsSpoofedXFF(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// No trusted proxies configured — XFF should be ignored
	// 未配置可信代理——XFF 应被忽略

	allowRouter := NewTestEngine()
	applyTransportMiddleware(allowRouter, IPAllowlist("10.0.0.0/8"))
	allowRouter.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	// Request from 192.168.x.x with spoofed XFF claiming to be 10.x.x.x
	// 来自 192.168.x.x 的请求，伪造 XFF 声称来自 10.x.x.x
	allowReq := httptest.NewRequest(http.MethodGet, "/admin", nil)
	allowReq.RemoteAddr = "192.168.1.1:12345"
	allowReq.Header.Set("X-Forwarded-For", "10.1.2.3")
	allowRecorder := httptest.NewRecorder()
	allowRouter.ServeHTTP(allowRecorder, allowReq)

	// Should be forbidden because XFF is ignored and real IP is 192.168.1.1
	// 应被拒绝，因为 XFF 被忽略，真实 IP 为 192.168.1.1
	if allowRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected spoofed XFF request 403, got %d", allowRecorder.Code)
	}
}

// TestLocaleUsesQueryThenHeaderThenDefault verifies locale negotiation order and response header output.
//
// TestLocaleUsesQueryThenHeaderThenDefault 验证语言协商顺序与响应头输出。
func TestLocaleUsesQueryThenHeaderThenDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queryRouter := NewTestEngine()
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
