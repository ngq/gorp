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
