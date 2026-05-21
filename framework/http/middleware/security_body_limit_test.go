// Package middleware_test provides unit tests for security headers, body limit, and selector middleware.
//
// 适用场景：
// - 安全头注入
// - 请求体大小限制
// - 基于谓词的中间件选择
package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestSecurityHeadersWritesDefaultsAndCustomValues verifies default security headers and custom overrides.
//
// TestSecurityHeadersWritesDefaultsAndCustomValues 验证默认安全头与自定义覆盖值。
func TestSecurityHeadersWritesDefaultsAndCustomValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	defaultRouter := NewTestEngine()
	applyTransportMiddleware(defaultRouter, SecurityHeaders(SecurityHeadersOptions{}))
	defaultRouter.GET("/headers", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	defaultReq := httptest.NewRequest(http.MethodGet, "/headers", nil)
	defaultRecorder := httptest.NewRecorder()
	defaultRouter.ServeHTTP(defaultRecorder, defaultReq)

	if defaultRecorder.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("expected default X-Frame-Options, got %q", defaultRecorder.Header().Get("X-Frame-Options"))
	}
	if defaultRecorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("expected default X-Content-Type-Options, got %q", defaultRecorder.Header().Get("X-Content-Type-Options"))
	}
	if defaultRecorder.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Fatalf("unexpected Referrer-Policy %q", defaultRecorder.Header().Get("Referrer-Policy"))
	}

	customRouter := NewTestEngine()
	applyTransportMiddleware(customRouter, SecurityHeaders(SecurityHeadersOptions{
		ContentSecurityPolicy: "default-src 'self'",
		PermissionsPolicy:     "geolocation=()",
	}))
	customRouter.GET("/headers", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	customReq := httptest.NewRequest(http.MethodGet, "/headers", nil)
	customRecorder := httptest.NewRecorder()
	customRouter.ServeHTTP(customRecorder, customReq)

	if customRecorder.Header().Get("Content-Security-Policy") != "default-src 'self'" {
		t.Fatalf("unexpected Content-Security-Policy %q", customRecorder.Header().Get("Content-Security-Policy"))
	}
	if customRecorder.Header().Get("Permissions-Policy") != "geolocation=()" {
		t.Fatalf("unexpected Permissions-Policy %q", customRecorder.Header().Get("Permissions-Policy"))
	}
}

// TestCompressionCompressesWhenAccepted verifies gzip compression for clients that advertise support.
//
// TestCompressionCompressesWhenAccepted 验证对声明支持 gzip 的客户端进行压缩。
func TestCompressionCompressesWhenAccepted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, Compression())
	router.GET("/gzip", func(c *gin.Context) {
		c.String(http.StatusOK, "payload")
	})

	req := httptest.NewRequest(http.MethodGet, "/gzip", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip content encoding, got %q", recorder.Header().Get("Content-Encoding"))
	}
	if recorder.Header().Get("Vary") != "Accept-Encoding" {
		t.Fatalf("expected Vary Accept-Encoding, got %q", recorder.Header().Get("Vary"))
	}

	reader, err := gzip.NewReader(recorder.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read gzip body: %v", err)
	}
	if string(body) != "payload" {
		t.Fatalf("expected decompressed payload, got %q", string(body))
	}
}

// TestCompressionSkipsWhenNotAccepted verifies that compression is bypassed when the client does not accept gzip.
//
// TestCompressionSkipsWhenNotAccepted 验证客户端不接受 gzip 时会跳过压缩。
func TestCompressionSkipsWhenNotAccepted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, Compression())
	router.GET("/plain", func(c *gin.Context) {
		c.String(http.StatusOK, "payload")
	})

	req := httptest.NewRequest(http.MethodGet, "/plain", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Header().Get("Content-Encoding") != "" {
		t.Fatalf("expected no compression, got %q", recorder.Header().Get("Content-Encoding"))
	}
	if recorder.Body.String() != "payload" {
		t.Fatalf("expected plain payload, got %q", recorder.Body.String())
	}
}

// TestBodyLimitAllowsSmallBodyAndRejectsLargeBody verifies body-size guardrails for small and oversized payloads.
//
// TestBodyLimitAllowsSmallBodyAndRejectsLargeBody 验证请求体大小护栏对小包与超大包的处理。
func TestBodyLimitAllowsSmallBodyAndRejectsLargeBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, BodyLimit(4))
	router.POST("/upload", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.String(http.StatusOK, string(body))
	})

	smallReq := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewBufferString("ok"))
	smallRecorder := httptest.NewRecorder()
	router.ServeHTTP(smallRecorder, smallReq)

	if smallRecorder.Code != http.StatusOK {
		t.Fatalf("expected small body request 200, got %d", smallRecorder.Code)
	}
	if smallRecorder.Body.String() != "ok" {
		t.Fatalf("expected small body echo, got %q", smallRecorder.Body.String())
	}

	largeReq := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewBufferString("large"))
	largeRecorder := httptest.NewRecorder()
	router.ServeHTTP(largeRecorder, largeReq)

	if largeRecorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected large body request 413, got %d", largeRecorder.Code)
	}
}

// TestSelectorWhenAppliesOnlyOnMatchedRoutes verifies predicate-based middleware selection.
//
// TestSelectorWhenAppliesOnlyOnMatchedRoutes 验证基于谓词的中间件选择行为。
func TestSelectorWhenAppliesOnlyOnMatchedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	applyTransportMiddleware(router, When(MatchPrefix("/admin"), SecurityHeaders(SecurityHeadersOptions{
		XFrameOptions: "SAMEORIGIN",
	})))
	router.GET("/admin/panel", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	router.GET("/public/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	adminReq := httptest.NewRequest(http.MethodGet, "/admin/panel", nil)
	adminRecorder := httptest.NewRecorder()
	router.ServeHTTP(adminRecorder, adminReq)

	if adminRecorder.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Fatalf("expected selector-applied header, got %q", adminRecorder.Header().Get("X-Frame-Options"))
	}

	publicReq := httptest.NewRequest(http.MethodGet, "/public/ping", nil)
	publicRecorder := httptest.NewRecorder()
	router.ServeHTTP(publicRecorder, publicReq)

	if publicRecorder.Header().Get("X-Frame-Options") != "" {
		t.Fatalf("expected selector to skip public route, got %q", publicRecorder.Header().Get("X-Frame-Options"))
	}
}
