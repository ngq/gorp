package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
)

func TestRequestIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(adaptMiddleware(RequestIdentity()))

	r.GET("/test", func(c *gin.Context) {
		rid := GetRequestID(c)
		tid := GetTraceID(c)
		ctxRID, ok := supportcontract.FromRequestIDContext(c.Request.Context())
		if !ok || ctxRID == "" {
			t.Fatal("expected request id in request context")
		}
		ctxTID, ok := supportcontract.FromTraceIDContext(c.Request.Context())
		if !ok || ctxTID == "" {
			t.Fatal("expected trace id in request context")
		}
		if rid != ctxRID {
			t.Fatalf("request id mismatch: expected %s, got %s", ctxRID, rid)
		}
		if tid != ctxTID {
			t.Fatalf("trace id mismatch: expected %s, got %s", ctxTID, tid)
		}
		c.JSON(http.StatusOK, gin.H{"request_id": rid, "trace_id": tid})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	rid := w.Header().Get("X-Request-Id")
	tid := w.Header().Get("X-Trace-Id")
	if len(rid) != 32 {
		t.Fatalf("expected request id length 32, got %d", len(rid))
	}
	if tid != rid {
		t.Fatalf("expected trace id to fall back to request id, got request_id=%s trace_id=%s", rid, tid)
	}
}

func TestRequestIdentityWithHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(adaptMiddleware(RequestIdentity()))

	r.GET("/test", func(c *gin.Context) {
		rid := GetRequestID(c)
		tid := GetTraceID(c)
		if rid != "test-request-id-12345" {
			t.Fatalf("expected request id to be preserved, got %s", rid)
		}
		if tid != "test-trace-id-67890" {
			t.Fatalf("expected trace id to be preserved, got %s", tid)
		}
		c.JSON(http.StatusOK, gin.H{"request_id": rid, "trace_id": tid})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-Id", "test-request-id-12345")
	req.Header.Set("X-Trace-Id", "test-trace-id-67890")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if got := w.Header().Get("X-Request-Id"); got != "test-request-id-12345" {
		t.Fatalf("expected request id header to be preserved, got %s", got)
	}
	if got := w.Header().Get("X-Trace-Id"); got != "test-trace-id-67890" {
		t.Fatalf("expected trace id header to be preserved, got %s", got)
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
