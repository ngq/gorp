// Package jwt_test provides unit tests for JWT auth middleware request context behavior.
//
// 适用场景：
// - 验证 JWT auth 中间件向 request context 写入鉴权信息的正确性。
// - 确保 Gin 集成下 middleware 的上下文传递行为稳定。
package jwt

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

// TestAuthMiddlewareWritesRequestContext verifies the auth middleware correctly writes JWT claims to request context.
//
// TestAuthMiddlewareWritesRequestContext 验证认证中间件正确地将 JWT claims 写入请求上下文。
func TestAuthMiddlewareWritesRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := NewJWTService("secret", "issuer", "aud")
	claims := jwtSvc.NewClaims(7, "customer", "alice", []string{"user"}, 60)
	token, err := jwtSvc.Sign(claims)
	require.NoError(t, err)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
		wrapped := AuthMiddleware(jwtSvc, "customer")(func(inner transportcontract.HTTPContext) {
			if inner != nil && inner.Request() != nil {
				c.Request = inner.Request()
			}
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/me", func(c *gin.Context) {
		gotClaims, ok := securitycontract.FromJWTClaimsContext(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, int64(7), gotClaims.SubjectID)

		subjectID, ok := securitycontract.FromSubjectIDContext(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, int64(7), subjectID)

		subjectType, ok := securitycontract.FromSubjectTypeContext(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, "customer", subjectType)

		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}

// TestRequestContextHelpers verifies helper functions extract JWT claims, subject ID, and subject type from context.
//
// TestRequestContextHelpers 验证辅助函数能从上下文正确提取 JWT claims、subject ID 和 subject type。
func TestRequestContextHelpers(t *testing.T) {
	claims := &securitycontract.JWTClaims{
		SubjectID:   9,
		SubjectType: "admin",
		SubjectName: "alice",
	}
	ctx := securitycontract.NewJWTClaimsContext(t.Context(), claims)
	ctx = securitycontract.NewSubjectIDContext(ctx, claims.SubjectID)
	ctx = securitycontract.NewSubjectTypeContext(ctx, claims.SubjectType)

	gotClaims, ok := ClaimsFromRequestContext(ctx)
	require.True(t, ok)
	require.Equal(t, int64(9), gotClaims.SubjectID)

	subjectID, ok := SubjectIDFromRequestContext(ctx)
	require.True(t, ok)
	require.Equal(t, int64(9), subjectID)

	subjectType, ok := SubjectTypeFromRequestContext(ctx)
	require.True(t, ok)
	require.Equal(t, "admin", subjectType)
}

// TestSubjectIDFromContextNilSafe verifies SubjectIDFromContext returns zero value and false for nil context.
//
// TestSubjectIDFromContextNilSafe 验证 SubjectIDFromContext 对 nil 上下文返回零值和 false。
func TestSubjectIDFromContextNilSafe(t *testing.T) {
	subjectID, ok := SubjectIDFromContext(nil)
	require.False(t, ok)
	require.Equal(t, int64(0), subjectID)
}

// TestAuthMiddlewareRejectsMissingBearerToken verifies that requests without Bearer token are rejected with 401.
//
// TestAuthMiddlewareRejectsMissingBearerToken 验证不带 Bearer token 的请求会被 401 拒绝。
func TestAuthMiddlewareRejectsMissingBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := NewJWTService("secret", "issuer", "aud")

	r := gin.New()
	r.Use(func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
		wrapped := AuthMiddleware(jwtSvc, "")(func(inner transportcontract.HTTPContext) {
			if inner != nil && inner.Request() != nil {
				c.Request = inner.Request()
			}
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	// 不设置 Authorization header
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAuthMiddlewareRejectsInvalidToken verifies that requests with an invalid token are rejected with 401.
//
// TestAuthMiddlewareRejectsInvalidToken 验证无效 token 的请求会被 401 拒绝。
func TestAuthMiddlewareRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := NewJWTService("secret", "issuer", "aud")

	r := gin.New()
	r.Use(func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
		wrapped := AuthMiddleware(jwtSvc, "")(func(inner transportcontract.HTTPContext) {
			if inner != nil && inner.Request() != nil {
				c.Request = inner.Request()
			}
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAuthMiddlewareRejectsSubjectTypeMismatch verifies that token with unexpected subject type is rejected with 403.
//
// TestAuthMiddlewareRejectsSubjectTypeMismatch 验证主体类型不匹配时请求会被 403 拒绝。
func TestAuthMiddlewareRejectsSubjectTypeMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := NewJWTService("secret", "issuer", "aud")
	// 签发 subjectType = "admin" 的 token
	claims := jwtSvc.NewClaims(7, "admin", "admin1", []string{"admin"}, 60)
	token, err := jwtSvc.Sign(claims)
	require.NoError(t, err)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
		// 期望 subjectType 为 "customer"
		wrapped := AuthMiddleware(jwtSvc, "customer")(func(inner transportcontract.HTTPContext) {
			if inner != nil && inner.Request() != nil {
				c.Request = inner.Request()
			}
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

// TestAuthMiddlewareRejectsNilJWTService verifies that nil JWT service results in 401.
//
// TestAuthMiddlewareRejectsNilJWTService 验证 JWT 服务为 nil 时返回 401。
func TestAuthMiddlewareRejectsNilJWTService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
		wrapped := AuthMiddleware(nil, "")(func(inner transportcontract.HTTPContext) {
			if inner != nil && inner.Request() != nil {
				c.Request = inner.Request()
			}
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}
