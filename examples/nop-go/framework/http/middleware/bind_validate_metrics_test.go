// Package middleware_test provides unit tests for bind, validation, and metrics middleware.
//
// 适用场景：
// - 请求体绑定与校验
// - Prometheus 指标记录
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
)

// TestBindAndValidateJSONStoresValidatedBody verifies that successful validation stores the validated body in request context.
//
// TestBindAndValidateJSONStoresValidatedBody 验证成功校验后会把已校验对象写入请求上下文。
func TestBindAndValidateJSONStoresValidatedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	validator := &stubValidator{
		validateFn: func(ctx context.Context, obj any) error {
			input, ok := obj.(*struct {
				Name string `json:"name"`
			})
			if !ok || strings.TrimSpace(input.Name) == "" {
				return resiliencecontract.BadRequest(resiliencecontract.ErrorReasonBadRequest, "name is required")
			}
			return nil
		},
	}

	router.POST("/validate", func(c *gin.Context) {
		httpCtx := newContext(c)
		input := &struct {
			Name string `json:"name"`
		}{}
		if err := BindAndValidateJSON(httpCtx, validator, input); err != nil {
			return
		}
		validatedBody, ok := supportcontract.FromValidatedBodyContext(c.Request.Context())
		if !ok {
			c.String(http.StatusInternalServerError, "missing validated body")
			return
		}
		body := validatedBody.(*struct {
			Name string `json:"name"`
		})
		c.String(http.StatusOK, body.Name)
	})

	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{"name":"alice"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Body.String() != "alice" {
		t.Fatalf("expected validated body name alice, got %q", recorder.Body.String())
	}
}

// TestBindAndValidateJSONReturnsUnifiedError verifies unified validation error output and detail propagation.
//
// TestBindAndValidateJSONReturnsUnifiedError 验证统一校验错误输出与详情透传。
func TestBindAndValidateJSONReturnsUnifiedError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	validator := &stubValidator{
		validateFn: func(context.Context, any) error {
			return resiliencecontract.BadRequest(
				resiliencecontract.ErrorReasonBadRequest,
				"validation failed",
			).WithMetadata(map[string]string{"validation_errors": `["name is required"]`})
		},
	}

	router.POST("/validate", func(c *gin.Context) {
		httpCtx := newContext(c)
		input := &struct {
			Name string `json:"name"`
		}{}
		_ = BindAndValidateJSON(httpCtx, validator, input)
	})

	req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{"name":""}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	var resp ValidateErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode validation response: %v", err)
	}
	if resp.Message != "validation failed" {
		t.Fatalf("expected validation failed message, got %q", resp.Message)
	}
	if resp.Details != `["name is required"]` {
		t.Fatalf("expected validation details, got %q", resp.Details)
	}
}

// TestMetricsMiddlewareRecordsRequestCount verifies request counter increments with method, route, and status labels.
//
// TestMetricsMiddlewareRecordsRequestCount 验证请求计数器会按 method、route 和 status 标签递增。
func TestMetricsMiddlewareRecordsRequestCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	labels := map[string]string{
		"method": http.MethodGet,
		"path":   "/metrics/:id",
		"status": "204",
	}
	beforeCount := counterValue("gorp_http_requests_total", labels)

	router := NewTestEngine()
	applyTransportMiddleware(router, MetricsMiddleware())
	router.GET("/metrics/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics/42", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	afterCount := counterValue("gorp_http_requests_total", labels)
	if afterCount != beforeCount+1 {
		t.Fatalf("expected request counter to increase by 1, got before=%v after=%v", beforeCount, afterCount)
	}
}
