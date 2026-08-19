package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/http/middleware"
	ginprovider "github.com/ngq/gorp/framework/provider/gin"
	"github.com/stretchr/testify/require"
)

func TestDataMaskingMiddleware_MasksJSONResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.Use(ginprovider.AdaptMiddleware(middleware.DataMaskingMiddleware()))

	engine.GET("/api/user/profile", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"mobile":    "13812345678",
			"id_card":   "110101199003072345",
			"bank_card": "6222021234567890123",
			"email":     "alice@example.com",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/profile", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	bodyStr := w.Body.String()

	// 1. Mobile Masking
	require.Contains(t, bodyStr, "138****5678")
	require.NotContains(t, bodyStr, "13812345678")

	// 2. ID Card Masking
	require.Contains(t, bodyStr, "110101********2345")
	require.NotContains(t, bodyStr, "110101199003072345")

	// 3. Bank Card Masking
	require.Contains(t, bodyStr, "622202******0123")
	require.NotContains(t, bodyStr, "6222021234567890123")

	// 4. Email Masking
	require.Contains(t, bodyStr, "al***@example.com")
	require.NotContains(t, bodyStr, "alice@example.com")
}
