package gorp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp"
	"github.com/stretchr/testify/require"
)

func TestEngine_Default_BasicRoutesAndCapabilities(t *testing.T) {
	app := gorp.Default()
	require.NotNil(t, app)
	require.NotNil(t, app.Gin())
	require.NotNil(t, app.Container())
	require.NotNil(t, app.Config())
	require.NotNil(t, app.Logger())
	require.Nil(t, app.DB()) // Config-driven: nil when not configured in YAML, no panic!

	// Test routing
	app.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	app.Gin().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"message":"pong"`)
}

func TestEngine_RunContext_GracefulShutdown(t *testing.T) {
	app := gorp.Default()
	app.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := app.RunContext(ctx, "127.0.0.1:0")
	require.NoError(t, err)
}
