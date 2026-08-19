package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/httpx"
	"github.com/stretchr/testify/require"
)

func TestMountDebugDashboard_EndPoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	c := container.New()

	httpx.MountDebugDashboard(engine, c, "/debug/gorp")

	// 注册示例模拟业务路由
	engine.GET("/api/v1/users/:id", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"id": ctx.Param("id")})
	})
	engine.POST("/api/v1/orders", func(ctx *gin.Context) {
		ctx.JSON(201, gin.H{"order_id": "1001"})
	})

	// 1. 测试 Admin Web Dashboard HTML 端点
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/debug/gorp", nil)
	engine.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusOK, w1.Code)
	require.Contains(t, w1.Body.String(), "gorp Developer Console")
	require.Contains(t, w1.Body.String(), "DI 容器拓扑 (DAG)")

	// 2. 测试 /debug/gorp/api/container 端点
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/debug/gorp/api/container", nil)
	engine.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusOK, w2.Code)
	require.Contains(t, w2.Body.String(), "[")

	// 3. 测试 /debug/gorp/api/routes 端点
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/debug/gorp/api/routes", nil)
	engine.ServeHTTP(w3, req3)
	require.Equal(t, http.StatusOK, w3.Code)
	require.Contains(t, w3.Body.String(), "/api/v1/users/:id")

	// 4. 测试 /debug/gorp/api/metrics 端点
	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodGet, "/debug/gorp/api/metrics", nil)
	engine.ServeHTTP(w4, req4)
	require.Equal(t, http.StatusOK, w4.Code)
	require.Contains(t, w4.Body.String(), "goroutines")

	// 5. 测试 /debug/gorp/openapi.json 端点 (OpenAPI 3 自动导出)
	w5 := httptest.NewRecorder()
	req5 := httptest.NewRequest(http.MethodGet, "/debug/gorp/openapi.json", nil)
	engine.ServeHTTP(w5, req5)
	require.Equal(t, http.StatusOK, w5.Code)
	require.Contains(t, w5.Body.String(), `"openapi": "3.0.0"`)
	require.Contains(t, w5.Body.String(), "/api/v1/users/{id}")
}

func TestAdminDashboard_ProductionProtection(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	engine := gin.New()
	c := container.New()

	httpx.RegisterAdminDashboard(engine, c, "/debug/gorp")

	// 生产环境下默认无 Token 访问返回 403 Forbidden
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/debug/gorp/api/routes", nil)
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), "disabled in production mode")
}

func TestAdminDashboard_TokenAuth(t *testing.T) {
	engine := gin.New()
	c := container.New()

	httpx.RegisterAdminDashboard(engine, c, "/debug/gorp", httpx.WithAdminToken("my-secret-admin-token"))

	// 1. 无 Token 访问 -> 401
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/debug/gorp/api/metrics", nil)
	engine.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusUnauthorized, w1.Code)

	// 2. 带有效 Query Token 访问 -> 200
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/debug/gorp/api/metrics?token=my-secret-admin-token", nil)
	engine.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusOK, w2.Code)

	// 3. 带有效 Bearer Header Token 访问 -> 200
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/debug/gorp/api/metrics", nil)
	req3.Header.Set("Authorization", "Bearer my-secret-admin-token")
	engine.ServeHTTP(w3, req3)
	require.Equal(t, http.StatusOK, w3.Code)
}

