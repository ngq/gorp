package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"github.com/ngq/gorp/framework/http/middleware"
	ginprovider "github.com/ngq/gorp/framework/provider/gin"
	"github.com/stretchr/testify/require"
)

func TestRequireRole_And_RequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// 模拟已认证的 JWT Claims 注入
	engine.Use(func(c *gin.Context) {
		claims := &securitycontract.JWTClaims{
			SubjectID: 101,
			Roles:     []string{"admin", "order:create"},
		}
		c.Set(securitycontract.ContextJWTClaimsKey, claims)
		c.Next()
	})

	engine.GET("/admin/dashboard", ginprovider.AdaptMiddleware(middleware.RequireRole("admin")), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	engine.GET("/user/only", ginprovider.AdaptMiddleware(middleware.RequireRole("vip_user")), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 1. 拥有 admin 角色 -> 200 OK
	req1 := httptest.NewRequest(http.MethodGet, "/admin/dashboard", nil)
	w1 := httptest.NewRecorder()
	engine.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusOK, w1.Code)

	// 2. 缺少 vip_user 角色 -> 403 Forbidden
	req2 := httptest.NewRequest(http.MethodGet, "/user/only", nil)
	w2 := httptest.NewRecorder()
	engine.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusForbidden, w2.Code)
}
