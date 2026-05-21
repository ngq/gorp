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
	"time"

	"github.com/gin-gonic/gin"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	ginprovider "github.com/ngq/gorp/framework/provider/gin"
	"github.com/stretchr/testify/require"
)

// testContext implements Context for testing
type testContext struct {
	gin *gin.Context
}

func (c *testContext) Deadline() (deadline time.Time, ok bool) {
	return c.gin.Request.Context().Deadline()
}

func (c *testContext) Done() <-chan struct{} {
	return c.gin.Request.Context().Done()
}

func (c *testContext) Err() error {
	return c.gin.Request.Context().Err()
}

func (c *testContext) Value(key any) any {
	return c.gin.Request.Context().Value(key)
}

func (c *testContext) Request() *http.Request {
	return c.gin.Request
}

func (c *testContext) Response() http.ResponseWriter {
	return c.gin.Writer
}

func (c *testContext) Param(key string) string {
	return c.gin.Param(key)
}

func (c *testContext) Query(key string) string {
	return c.gin.Query(key)
}

func (c *testContext) DefaultQuery(key, defaultValue string) string {
	return c.gin.DefaultQuery(key, defaultValue)
}

func (c *testContext) GetHeader(key string) string {
	return c.gin.GetHeader(key)
}

func (c *testContext) SetHeader(key, value string) {
	c.gin.Header(key, value)
}

func (c *testContext) Bind(obj any) error {
	return c.gin.ShouldBind(obj)
}

func (c *testContext) BindJSON(obj any) error {
	return c.gin.ShouldBindJSON(obj)
}

func (c *testContext) BindQuery(obj any) error {
	return c.gin.ShouldBindQuery(obj)
}

func (c *testContext) JSON(status int, body any) {
	c.gin.JSON(status, body)
}

func (c *testContext) String(status int, body string) {
	c.gin.String(status, body)
}

func (c *testContext) XML(status int, body any) {
	c.gin.XML(status, body)
}

func (c *testContext) Data(status int, contentType string, body []byte) {
	c.gin.Data(status, contentType, body)
}

func (c *testContext) Redirect(status int, location string) {
	c.gin.Redirect(status, location)
}

func (c *testContext) Status(code int) {
	c.gin.Status(code)
}

func (c *testContext) RoutePath() string {
	return c.gin.FullPath()
}

func (c *testContext) ResponseStatus() int {
	return c.gin.Writer.Status()
}

func (c *testContext) Get(key string) any {
	val, _ := c.gin.Get(key)
	return val
}

func (c *testContext) Set(key string, value any) {
	c.gin.Set(key, value)
}

func (c *testContext) Abort(status int) {
	c.gin.AbortWithStatus(status)
}

func (c *testContext) AbortWithJSON(status int, body any) {
	c.gin.AbortWithStatusJSON(status, body)
}

func (c *testContext) IsAborted() bool {
	return c.gin.IsAborted()
}

func (c *testContext) Next() {
	c.gin.Next()
}

func newTestContext(c *gin.Context) transportcontract.Context {
	return &testContext{gin: c}
}

// TestAuthMiddlewareWritesRequestContext verifies the auth middleware correctly writes JWT claims to request context.
//
// TestAuthMiddlewareWritesRequestContext 验证认证中间件正确地将 JWT claims 写入请求上下文。
func TestAuthMiddlewareWritesRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := NewJWTService("secret", "issuer", "aud")
	claims := jwtSvc.NewClaims(7, "customer", "alice", []string{"user"}, 60)
	token, err := jwtSvc.Sign(claims)
	require.NoError(t, err)

	r := ginprovider.NewTestEngine()
	r.Use(func(c *gin.Context) {
		httpCtx := newTestContext(c)
		wrapped := AuthMiddleware(jwtSvc, "customer")(func(inner transportcontract.Context) {
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/me", func(c *gin.Context) {
		// Claims are stored via c.Set()
		claimsVal, exists := c.Get(ContextJWTClaimsKey)
		require.True(t, exists)
		gotClaims, ok := claimsVal.(*securitycontract.JWTClaims)
		require.True(t, ok)
		require.Equal(t, int64(7), gotClaims.SubjectID)

		subjectIDVal, exists := c.Get(ContextSubjectIDKey)
		require.True(t, exists)
		subjectID, ok := subjectIDVal.(int64)
		require.True(t, ok)
		require.Equal(t, int64(7), subjectID)

		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

// TestAuthMiddlewareRejectsMissingToken verifies the middleware rejects requests without bearer token.
//
// TestAuthMiddlewareRejectsMissingToken 验证中间件拒绝没有 bearer token 的请求。
func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := NewJWTService("secret", "issuer", "aud")

	r := ginprovider.NewTestEngine()
	r.Use(func(c *gin.Context) {
		httpCtx := newTestContext(c)
		wrapped := AuthMiddleware(jwtSvc, "customer")(func(inner transportcontract.Context) {
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAuthMiddlewareRejectsInvalidToken verifies the middleware rejects invalid tokens.
//
// TestAuthMiddlewareRejectsInvalidToken 验证中间件拒绝无效 token。
func TestAuthMiddlewareRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := NewJWTService("secret", "issuer", "aud")

	r := ginprovider.NewTestEngine()
	r.Use(func(c *gin.Context) {
		httpCtx := newTestContext(c)
		wrapped := AuthMiddleware(jwtSvc, "customer")(func(inner transportcontract.Context) {
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAuthMiddlewareRejectsWrongSubjectType verifies the middleware rejects tokens with wrong subject type.
//
// TestAuthMiddlewareRejectsWrongSubjectType 验证中间件拒绝主体类型不匹配的 token。
func TestAuthMiddlewareRejectsWrongSubjectType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := NewJWTService("secret", "issuer", "aud")
	claims := jwtSvc.NewClaims(7, "admin", "alice", []string{"user"}, 60)
	token, err := jwtSvc.Sign(claims)
	require.NoError(t, err)

	r := ginprovider.NewTestEngine()
	r.Use(func(c *gin.Context) {
		httpCtx := newTestContext(c)
		wrapped := AuthMiddleware(jwtSvc, "customer")(func(inner transportcontract.Context) {
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

// TestAuthMiddlewareWithSkipPaths verifies skip paths bypass authentication.
//
// TestAuthMiddlewareWithSkipPaths 验证跳过路径绕过认证。
func TestAuthMiddlewareWithSkipPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := NewJWTService("secret", "issuer", "aud")

	r := ginprovider.NewTestEngine()
	r.Use(func(c *gin.Context) {
		httpCtx := newTestContext(c)
		wrapped := AuthMiddleware(jwtSvc, "customer", WithSkipPaths("/health", "/login"))(func(inner transportcontract.Context) {
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Test skip paths work without token
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/login", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Test non-skip path requires token
	req = httptest.NewRequest(http.MethodGet, "/me", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAuthMiddlewareWithRequiredRoles verifies role checking.
//
// TestAuthMiddlewareWithRequiredRoles 验证角色校验。
func TestAuthMiddlewareWithRequiredRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := NewJWTService("secret", "issuer", "aud")

	// User with admin role
	adminClaims := jwtSvc.NewClaims(1, "user", "admin", []string{"admin"}, 60)
	adminToken, err := jwtSvc.Sign(adminClaims)
	require.NoError(t, err)

	// User without admin role
	userClaims := jwtSvc.NewClaims(2, "user", "user", []string{"user"}, 60)
	userToken, err := jwtSvc.Sign(userClaims)
	require.NoError(t, err)

	r := ginprovider.NewTestEngine()
	r.Use(func(c *gin.Context) {
		httpCtx := newTestContext(c)
		wrapped := AuthMiddleware(jwtSvc, "user", WithRequiredRoles("admin"))(func(inner transportcontract.Context) {
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Admin can access
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Regular user cannot access
	req = httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}
