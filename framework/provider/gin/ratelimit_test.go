package gin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type denyLimiter struct{}

func (denyLimiter) Allow(_ string) bool { return false }

func TestRateLimitMiddlewareUsesUnifiedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimitMiddleware(denyLimiter{}, IPKeyFunc))
	r.GET("/demo", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/demo", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != CodeTooManyRequests {
		t.Fatalf("expected business code %d, got %d", CodeTooManyRequests, resp.Code)
	}
	if resp.Message != "rate limit exceeded" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}
