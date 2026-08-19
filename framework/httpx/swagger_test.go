package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/httpx"
	"github.com/stretchr/testify/require"
)

func TestGinSwaggerHandler_ServesHTMLAndSpec(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	tmpDir := t.TempDir()
	specFile := filepath.Join(tmpDir, "openapi.json")
	err := os.WriteFile(specFile, []byte(`{"openapi":"3.0.0","info":{"title":"Test API"}}`), 0644)
	require.NoError(t, err)

	handler := httpx.GinSwaggerHandler(specFile)
	engine.GET("/swagger", handler)
	engine.GET("/swagger/*any", handler)

	// 1. 请求 HTML UI 页面
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	engine.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusOK, w1.Code)
	require.Contains(t, w1.Body.String(), "swagger-ui")

	// 2. 请求 spec JSON
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/swagger/spec", nil)
	engine.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusOK, w2.Code)
	require.Contains(t, w2.Body.String(), `"openapi":"3.0.0"`)
}
