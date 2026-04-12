package gin

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRequestID 验证 RequestID 中间件是否正常工作。
func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		rid := GetRequestID(c)
		c.String(200, rid)
	})

	// 测试：没有请求头时自动生成
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	rid := w.Body.String()
	if len(rid) != 32 {
		t.Errorf("expected request id length 32, got %d", len(rid))
	}

	// 验证响应头
	headerRID := w.Header().Get("X-Request-Id")
	if headerRID != rid {
		t.Errorf("header request id mismatch: expected %s, got %s", rid, headerRID)
	}
}

// TestRequestIDWithHeader 验证 RequestID 中间件能正确透传请求头中的 ID。
func TestRequestIDWithHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		rid := GetRequestID(c)
		c.String(200, rid)
	})

	// 测试：有请求头时透传
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-Id", "test-request-id-12345")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	rid := w.Body.String()
	if rid != "test-request-id-12345" {
		t.Errorf("expected request id to be preserved, got %s", rid)
	}
}

// TestTraceID 验证 TraceID 中间件是否正常工作。
func TestTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.Use(TraceID())
	r.GET("/test", func(c *gin.Context) {
		tid := GetTraceID(c)
		c.String(200, tid)
	})

	// 测试：没有请求头时使用 request id
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	tid := w.Body.String()
	if len(tid) != 32 {
		t.Errorf("expected trace id length 32, got %d", len(tid))
	}
}

// TestTraceIDWithHeader 验证 TraceID 中间件能正确透传请求头中的 ID。
func TestTraceIDWithHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.Use(TraceID())
	r.GET("/test", func(c *gin.Context) {
		tid := GetTraceID(c)
		c.String(200, tid)
	})

	// 测试：有请求头时透传
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Trace-Id", "test-trace-id-67890")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	tid := w.Body.String()
	if tid != "test-trace-id-67890" {
		t.Errorf("expected trace id to be preserved, got %s", tid)
	}
}