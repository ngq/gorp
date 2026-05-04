package token

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type middlewareAuthStub struct {
	err error
}

func (s *middlewareAuthStub) Authenticate(ctx context.Context) (*securitycontract.ServiceIdentity, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &securitycontract.ServiceIdentity{ServiceName: "order-service"}, nil
}
func (s *middlewareAuthStub) GenerateToken(ctx context.Context, targetService string) (string, error) {
	return "", nil
}
func (s *middlewareAuthStub) VerifyToken(ctx context.Context, tokenString string) (*securitycontract.ServiceIdentity, error) {
	return s.Authenticate(ctx)
}

func TestServiceAuthHTTPMiddleware_AllowsAnonymousHTTPRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetResponseFuncs(
			c.JSON,
			func(code int, body string) { c.String(code, body) },
			c.XML,
			c.Data,
			c.Redirect,
			c.Status,
			func() int { return c.Writer.Status() },
		)
		wrapped := ServiceAuthHTTPMiddleware(&middlewareAuthStub{err: errors.New("should not be called")})(func(inner transportcontract.HTTPContext) {
			if inner != nil && inner.Request() != nil {
				c.Request = inner.Request()
			}
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestServiceAuthHTTPMiddleware_RejectsInvalidServiceToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetResponseFuncs(
			c.JSON,
			func(code int, body string) { c.String(code, body) },
			c.XML,
			c.Data,
			c.Redirect,
			c.Status,
			func() int { return c.Writer.Status() },
		)
		wrapped := ServiceAuthHTTPMiddleware(&middlewareAuthStub{err: errors.New("bad token")})(func(inner transportcontract.HTTPContext) {
			if inner != nil && inner.Request() != nil {
				c.Request = inner.Request()
			}
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Service-Token", "bad-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d, body=%s", w.Code, w.Body.String())
	}
}
