package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

type retryStub struct{}

func (retryStub) Do(ctx context.Context, fn func() error) error                         { return fn() }
func (retryStub) DoWithResult(ctx context.Context, fn func() (any, error)) (any, error) { return fn() }
func (retryStub) IsRetryable(err error) bool                                            { return err != nil }
func TestRetryMiddlewareUsesUnifiedTimeoutResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	policy := resiliencecontract.RetryPolicy{MaxAttempts: 2, RetryableCodes: []int{http.StatusServiceUnavailable}}
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
	policy := resiliencecontract.RetryPolicy{MaxAttempts: 1, RetryableCodes: []int{http.StatusServiceUnavailable}}
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

func TestRetryMiddlewareSkipsNonIdempotentMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	policy := resiliencecontract.RetryPolicy{MaxAttempts: 3, RetryableCodes: []int{http.StatusServiceUnavailable}}
	hits := 0
	r.Use(RetryMiddleware(retryStub{}, policy))
	r.POST("/retry", func(c *gin.Context) {
		hits++
		c.Status(http.StatusServiceUnavailable)
	})

	req := httptest.NewRequest(http.MethodPost, "/retry", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if hits != 1 {
		t.Fatalf("expected non-idempotent POST not retried, got hits=%d", hits)
	}
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestRetryMiddlewareRetriesIdempotentRequestsAndOnlyFlushesFinalResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	policy := resiliencecontract.RetryPolicy{
		MaxAttempts:    3,
		RetryableCodes: []int{http.StatusServiceUnavailable},
		InitialDelay:   0,
		MaxDelay:       0,
	}
	hits := 0
	r.Use(RetryMiddleware(retryStub{}, policy))
	r.GET("/retry", func(c *gin.Context) {
		hits++
		if hits == 1 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"attempt": hits})
			return
		}
		c.JSON(http.StatusOK, gin.H{"attempt": hits})
	})

	req := httptest.NewRequest(http.MethodGet, "/retry", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if hits != 2 {
		t.Fatalf("expected handler retried once, got hits=%d", hits)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected final 200, got %d", w.Code)
	}
	if got := w.Body.String(); got != "{\"attempt\":2}" {
		t.Fatalf("expected only final response body, got %q", got)
	}
}

func TestRetryAllMethodsMiddlewareReplaysRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	policy := resiliencecontract.RetryPolicy{
		MaxAttempts:    3,
		RetryableCodes: []int{http.StatusServiceUnavailable},
		InitialDelay:   0,
		MaxDelay:       0,
	}
	var bodies []string
	r.Use(RetryAllMethodsMiddleware(retryStub{}, policy))
	r.POST("/retry", func(c *gin.Context) {
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		bodies = append(bodies, string(data))
		if len(bodies) == 1 {
			c.Status(http.StatusServiceUnavailable)
			return
		}
		c.JSON(http.StatusOK, gin.H{"body": string(data), "attempt": len(bodies)})
	})

	req := httptest.NewRequest(http.MethodPost, "/retry", bytes.NewBufferString("payload"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(bodies) != 2 {
		t.Fatalf("expected two attempts, got %d", len(bodies))
	}
	if bodies[0] != "payload" || bodies[1] != "payload" {
		t.Fatalf("expected request body replayed on both attempts, got %#v", bodies)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected final 200, got %d", w.Code)
	}
	if got := w.Body.String(); got != "{\"attempt\":2,\"body\":\"payload\"}" {
		t.Fatalf("unexpected final body %q", got)
	}
}
