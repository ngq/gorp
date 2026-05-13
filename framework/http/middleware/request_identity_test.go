// Package middleware_test provides unit tests for request identity middleware.
//
// 适用场景：
// - 请求标识生成与传播
// - X-Request-Id 和 X-Trace-Id 头注入
// - 已有请求标识的复用
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRequestIdentityInjectsHeadersAndContext verifies automatic request identity generation and propagation.
//
// TestRequestIdentityInjectsHeadersAndContext 验证请求标识的自动生成与传播。
func TestRequestIdentityInjectsHeadersAndContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, RequestIdentity())
	router.GET("/identity", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"request_id": GetRequestID(c),
			"trace_id":   GetTraceID(c),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/identity", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id header")
	}
	if recorder.Header().Get("X-Trace-Id") == "" {
		t.Fatal("expected X-Trace-Id header")
	}
	if recorder.Header().Get("X-Request-Id") != recorder.Header().Get("X-Trace-Id") {
		t.Fatal("expected trace id to fall back to request id when no trace header is provided")
	}
}

// TestRequestIdentityReusesProvidedHeaders verifies that incoming request and trace ids are preserved.
//
// TestRequestIdentityReusesProvidedHeaders 验证传入的 request id 与 trace id 会被保留。
func TestRequestIdentityReusesProvidedHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	applyTransportMiddleware(router, RequestIdentity())
	router.GET("/identity", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/identity", nil)
	req.Header.Set("X-Request-Id", "req-123")
	req.Header.Set("X-Trace-Id", "trace-456")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Header().Get("X-Request-Id") != "req-123" {
		t.Fatalf("expected request id passthrough, got %q", recorder.Header().Get("X-Request-Id"))
	}
	if recorder.Header().Get("X-Trace-Id") != "trace-456" {
		t.Fatalf("expected trace id passthrough, got %q", recorder.Header().Get("X-Trace-Id"))
	}
}
