package token

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
)

type middlewareAuthStub struct {
	err error
}

func (s *middlewareAuthStub) Authenticate(ctx context.Context) (*contract.ServiceIdentity, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &contract.ServiceIdentity{ServiceName: "order-service"}, nil
}
func (s *middlewareAuthStub) AuthenticateWithToken(ctx context.Context, token string) (*contract.ServiceIdentity, error) {
	return s.Authenticate(ctx)
}
func (s *middlewareAuthStub) AuthenticateWithCert(ctx context.Context, cert *tls.Certificate) (*contract.ServiceIdentity, error) {
	return nil, errors.New("not implemented")
}
func (s *middlewareAuthStub) GenerateToken(ctx context.Context, targetService string) (string, error) {
	return "", nil
}
func (s *middlewareAuthStub) VerifyToken(ctx context.Context, tokenString string) (*contract.ServiceIdentity, error) {
	return s.Authenticate(ctx)
}

func TestServiceAuthHTTPMiddleware_AllowsAnonymousHTTPRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ServiceAuthHTTPMiddleware(&middlewareAuthStub{err: errors.New("should not be called")}))
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
	r.Use(ServiceAuthHTTPMiddleware(&middlewareAuthStub{err: errors.New("bad token")}))
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
