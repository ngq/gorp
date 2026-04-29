package gin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
)

type retryStub struct{}

func (retryStub) Do(ctx context.Context, fn func() error) error { return fn() }
func (retryStub) DoWithResult(ctx context.Context, fn func() (any, error)) (any, error) { return fn() }
func (retryStub) IsRetryable(err error) bool { return err != nil }
func (retryStub) GetDefaultPolicy() contract.RetryPolicy { return contract.RetryPolicy{} }

func TestRetryMiddlewareUsesUnifiedTimeoutResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	policy := contract.RetryPolicy{MaxAttempts: 2, RetryableCodes: []int{http.StatusServiceUnavailable}}
	r.Use(RetryMiddleware(retryStub{}, policy))
	r.GET("/retry", func(c *gin.Context) {
		c.Status(http.StatusServiceUnavailable)
	})

	req := httptest.NewRequest(http.MethodGet, "/retry", nil)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != CodeServiceUnavailable {
		t.Fatalf("expected code %d, got %d", CodeServiceUnavailable, resp.Code)
	}
	if resp.Message != "request timeout during retry" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

func TestRetryMiddlewareUsesUnifiedFinalErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	policy := contract.RetryPolicy{MaxAttempts: 1, RetryableCodes: []int{http.StatusServiceUnavailable}}
	r.Use(RetryMiddleware(retryStub{}, policy))
	r.GET("/retry", func(c *gin.Context) {
		c.Status(http.StatusServiceUnavailable)
	})

	req := httptest.NewRequest(http.MethodGet, "/retry", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != CodeServiceUnavailable {
		t.Fatalf("expected code %d, got %d", CodeServiceUnavailable, resp.Code)
	}
	if resp.Message == "" {
		t.Fatalf("expected non-empty message")
	}
	if resp.Data == nil {
		t.Fatalf("expected retry metadata in data")
	}
}
