// Package middleware_test provides unit tests for recovery middleware.
//
// 适用场景：
// - Panic 恢复与统一错误响应
// - Panic 日志记录
package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	frameworkbizlog "github.com/ngq/gorp/framework/log"
)

// TestRecoveryMiddlewareRecoversAndLogsPanic verifies panic recovery, unified 500 output, and panic logging.
//
// TestRecoveryMiddlewareRecoversAndLogsPanic 验证 panic 恢复、统一 500 输出和 panic 日志记录。
func TestRecoveryMiddlewareRecoversAndLogsPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := newStubLogger()
	router := NewTestEngine()
	applyTransportMiddleware(router,
		func(next transportcontract.Handler) transportcontract.Handler {
			return func(c transportcontract.Context) {
				c.Set("logger", logger)
				// Also update gin.Request.Context for context.Context value propagation
				if gc, ok := unwrapGinContext(c); ok && gc.Request != nil {
					gc.Request = gc.Request.WithContext(frameworkbizlog.WithContext(gc.Request.Context(), logger))
				}
				next(c)
			}
		},
		RecoveryMiddleware(),
	)
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
	var resp Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Message != "internal server error" {
		t.Fatalf("expected internal server error message, got %q", resp.Message)
	}

	entries := logger.Entries()
	if len(entries) == 0 {
		t.Fatal("expected panic log entry")
	}
	last := entries[len(entries)-1]
	if last.level != "error" || last.msg != "http panic recovered" {
		t.Fatalf("expected recovery error log, got level=%q msg=%q", last.level, last.msg)
	}
	if fieldValue(last.fields, "panic") != "boom" {
		t.Fatalf("expected panic field boom, got %v", fieldValue(last.fields, "panic"))
	}
}
