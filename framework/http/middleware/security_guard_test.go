package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/http/middleware"
	ginprovider "github.com/ngq/gorp/framework/provider/gin"
	"github.com/stretchr/testify/require"
)

func TestSecurityGuardMiddleware_BlocksAttacks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.Use(ginprovider.AdaptMiddleware(middleware.SecurityGuardMiddleware()))

	engine.POST("/api/submit", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 1. Normal Payload -> 200 OK
	req1 := httptest.NewRequest(http.MethodPost, "/api/submit", bytes.NewBufferString(`{"name":"Alice"}`))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	engine.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusOK, w1.Code)

	// 2. SQL Injection in URL -> 400 Bad Request
	req2 := httptest.NewRequest(http.MethodPost, "/api/submit?id=1%20UNION%20SELECT%20*%20FROM%20users", nil)
	w2 := httptest.NewRecorder()
	engine.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusBadRequest, w2.Code)

	// 3. XSS in Body -> 400 Bad Request
	req3 := httptest.NewRequest(http.MethodPost, "/api/submit", bytes.NewBufferString(`{"comment":"<script>alert(1)</script>"}`))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	engine.ServeHTTP(w3, req3)
	require.Equal(t, http.StatusBadRequest, w3.Code)

	// 4. Path Traversal in URL -> 400 Bad Request
	req4 := httptest.NewRequest(http.MethodPost, "/api/submit?file=../../etc/passwd", nil)
	w4 := httptest.NewRecorder()
	engine.ServeHTTP(w4, req4)
	require.Equal(t, http.StatusBadRequest, w4.Code)
}
