// Package bootstrap_test provides integration and boundary tests for HTTP service runtime bootstrapping.
//
// 适用场景：
// - 验证 HTTP Service runtime 的初始化、governance override 和 provider 注册行为。
// - 验证 pprof / governance inspect 端点的路由注册与响应格式。
// - 验证 governance summary 与 diagnostic 的构建与格式化逻辑。
package bootstrap

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Pprof 端点注册与行为
// =============================================================================

func TestAutoMigrateModelsNilRuntime(t *testing.T) {
	require.NoError(t, AutoMigrateModels(nil, struct{}{}))
}

func TestAutoMigrateModelsNilDB(t *testing.T) {
	rt := &HTTPServiceRuntime{}
	require.NoError(t, AutoMigrateModels(rt, struct{}{}))
}

func TestRegisterPprofEndpointsUsesMount(t *testing.T) {
	router := &recordingRouter{}

	RegisterPprofEndpoints(router)

	require.Contains(t, router.mounted, "/debug/pprof/")
	require.Contains(t, router.mounted, "/debug/pprof/cmdline")
	require.Contains(t, router.mounted, "/debug/pprof/profile")
	require.Contains(t, router.mounted, "/debug/pprof/symbol")
	require.Contains(t, router.mounted, "/debug/pprof/trace")
}

func TestRegisterPprofEndpointsServesIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &ginTestRouter{engine: engine}

	RegisterPprofEndpoints(router)

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.True(t, strings.Contains(w.Body.String(), "profile") || strings.Contains(w.Body.String(), "pprof"))
}

func TestRegisterPprofEndpointsRejectsPost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	router := &ginTestRouter{engine: engine}

	RegisterPprofEndpoints(router)

	req := httptest.NewRequest(http.MethodPost, "/debug/pprof/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}
