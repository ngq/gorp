// Application scenarios:
// - Verify newly added tenant and body-dump middleware on the HTTP mainline.
// - Lock tenant resolution, required-tenant enforcement, and request-response capture behavior.
// - Keep these newer middleware categories stable as they formally enter the mainline.
//
// 适用场景：
// - 验证新加入主线的租户与 body-dump 中间件能力。
// - 锁定租户解析、强制租户校验以及请求响应捕获行为。
// - 在这些新中间件正式进入主线时保持行为稳定。
package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// TestTenantResolvesFromParamAndWritesHeader verifies tenant resolution from route params and optional response header output.
//
// TestTenantResolvesFromParamAndWritesHeader 验证从路由参数解析租户以及可选响应头输出。
func TestTenantResolvesFromParamAndWritesHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	opts := DefaultTenantOptions()
	opts.WriteHeader = true
	applyTransportMiddleware(router, Tenant(opts))
	router.GET("/tenants/:tenant/orders", func(c *gin.Context) {
		c.String(http.StatusOK, GetTenant(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/tenants/acme/orders", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Body.String() != "acme" {
		t.Fatalf("expected tenant acme, got %q", recorder.Body.String())
	}
	if recorder.Header().Get("X-Tenant-ID") != "acme" {
		t.Fatalf("expected tenant response header, got %q", recorder.Header().Get("X-Tenant-ID"))
	}
}

// TestTenantRequiredRejectsMissingTenant verifies unified bad-request output when tenant is required but missing.
//
// TestTenantRequiredRejectsMissingTenant 验证租户必填但缺失时会返回统一错误请求响应。
func TestTenantRequiredRejectsMissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	opts := DefaultTenantOptions()
	opts.Required = true
	applyTransportMiddleware(router, Tenant(opts))
	router.GET("/orders", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

// TestBodyDumpCapturesExchange verifies request-response capture, tenant propagation, and bounded truncation.
//
// TestBodyDumpCapturesExchange 验证请求响应捕获、租户透传以及有限长度截断。
func TestBodyDumpCapturesExchange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var captured *HTTPExchangeCapture
	tenantOpts := DefaultTenantOptions()
	tenantOpts.WriteHeader = true

	applyTransportMiddleware(router,
		RequestIdentity(),
		Tenant(tenantOpts),
		BodyDump(BodyDumpOptions{
			CaptureRequestHeaders:  true,
			CaptureResponseHeaders: true,
			CaptureRequestBody:     true,
			CaptureResponseBody:    true,
			MaxBodyBytes:           4,
			OnCapture: func(ctx transportcontract.HTTPContext, dump *HTTPExchangeCapture) {
				captured = dump
				if tenant, ok := supportcontract.FromTenantContext(ctx.Context()); ok && dump.Tenant != tenant {
					t.Fatalf("expected dump tenant %q to match context tenant %q", dump.Tenant, tenant)
				}
			},
		}),
	)
	router.POST("/tenants/:tenant/orders", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.Header("X-Reply", "captured")
		c.String(http.StatusCreated, string(body))
	})

	req := httptest.NewRequest(http.MethodPost, "/tenants/acme/orders", bytes.NewBufferString("payload"))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-Trace-Id", "trace-1")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
	if captured == nil {
		t.Fatal("expected captured exchange")
	}
	if captured.Method != http.MethodPost {
		t.Fatalf("expected POST method, got %q", captured.Method)
	}
	if captured.Path != "/tenants/acme/orders" {
		t.Fatalf("expected request path, got %q", captured.Path)
	}
	if captured.Route != "/tenants/:tenant/orders" {
		t.Fatalf("expected route path, got %q", captured.Route)
	}
	if captured.Status != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", captured.Status)
	}
	if captured.Tenant != "acme" {
		t.Fatalf("expected tenant acme, got %q", captured.Tenant)
	}
	if captured.TraceID != "trace-1" {
		t.Fatalf("expected trace id trace-1, got %q", captured.TraceID)
	}
	if captured.RequestHeaders["Content-Type"] != "text/plain" {
		t.Fatalf("expected request content-type, got %q", captured.RequestHeaders["Content-Type"])
	}
	if captured.ResponseHeaders["X-Reply"] != "captured" {
		t.Fatalf("expected response header X-Reply, got %q", captured.ResponseHeaders["X-Reply"])
	}
	if string(captured.RequestBody) != "payl" || !captured.RequestTruncated {
		t.Fatalf("expected truncated request body payl, got %q truncated=%v", string(captured.RequestBody), captured.RequestTruncated)
	}
	if string(captured.ResponseBody) != "payl" || !captured.ResponseTruncated {
		t.Fatalf("expected truncated response body payl, got %q truncated=%v", string(captured.ResponseBody), captured.ResponseTruncated)
	}
}
