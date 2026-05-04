package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

func TestWhenMatchPrefixAppliesMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	called := false

	r.Use(adaptMiddleware(When(MatchPrefix("/api"), func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			called = true
			if next != nil {
				next(c)
			}
		}
	})))
	r.GET("/api/demo", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/demo", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected selector middleware to run")
	}
}

func TestWhenNotMatchedSkipsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	called := false

	r.Use(adaptMiddleware(When(MatchPath("/private"), func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			called = true
			if next != nil {
				next(c)
			}
		}
	})))
	r.GET("/public", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if called {
		t.Fatal("expected selector middleware to be skipped")
	}
}
