package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"grpc-demo/services/order/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRegisterRoutes_InvalidOrderID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r, &service.Services{Order: &service.OrderService{}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/abc", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)
	require.Equal(t, http.StatusBadRequest, resp.Code)
}
