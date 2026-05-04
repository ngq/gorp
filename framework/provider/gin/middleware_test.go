package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
)

// TestRequestID 验证 RequestID 中间件是否正常工作。
func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(adaptMiddleware(RequestID()))

	r.GET("/test", func(c *gin.Context) {
		rid := GetRequestID(c)
		ctxRID, ok := supportcontract.FromRequestIDContext(c.Request.Context())
		if !ok || ctxRID == "" {
			t.Fatal("expected request id in request context")
		}
		if ctxRID != rid {
			t.Fatalf("request context id mismatch: expected %s, got %s", rid, ctxRID)
		}
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
	r.Use(adaptMiddleware(RequestID()))

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
	r.Use(adaptMiddleware(RequestID()))

	r.Use(adaptMiddleware(TraceID()))

	r.GET("/test", func(c *gin.Context) {
		tid := GetTraceID(c)
		ctxTID, ok := supportcontract.FromTraceIDContext(c.Request.Context())
		if !ok || ctxTID == "" {
			t.Fatal("expected trace id in request context")
		}
		if ctxTID != tid {
			t.Fatalf("trace context id mismatch: expected %s, got %s", tid, ctxTID)
		}
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
	r.Use(adaptMiddleware(RequestID()))

	r.Use(adaptMiddleware(TraceID()))

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

func TestGetIDsFallbackToRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(supportcontract.NewRequestIDContext(req.Context(), "req-1"))
	req = req.WithContext(supportcontract.NewTraceIDContext(req.Context(), "trace-1"))
	ctx.Request = req

	if got := GetRequestID(ctx); got != "req-1" {
		t.Fatalf("expected request id req-1, got %s", got)
	}
	if got := GetTraceID(ctx); got != "trace-1" {
		t.Fatalf("expected trace id trace-1, got %s", got)
	}
}

func TestMountOnlyExposesReadMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	rt := newRouter(&engine.RouterGroup)
	rt.Mount("/mounted", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	getReq := httptest.NewRequest(http.MethodGet, "/mounted", nil)
	getW := httptest.NewRecorder()
	engine.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusNoContent {
		t.Fatalf("expected GET status %d, got %d", http.StatusNoContent, getW.Code)
	}

	postReq := httptest.NewRequest(http.MethodPost, "/mounted", nil)
	postW := httptest.NewRecorder()
	engine.ServeHTTP(postW, postReq)
	if postW.Code != http.StatusNotFound {
		t.Fatalf("expected POST status %d, got %d", http.StatusNotFound, postW.Code)
	}
}
