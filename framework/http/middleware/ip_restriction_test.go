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

func TestIPRestrictionMiddleware_CIDRAndGeoIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	geoResolver := func(ip string) string {
		if ip == "1.2.3.4" {
			return "CN"
		}
		if ip == "8.8.8.8" {
			return "US"
		}
		return "UNKNOWN"
	}

	engine := gin.New()
	engine.Use(ginprovider.AdaptMiddleware(middleware.IPRestrictionMiddleware(
		middleware.WithBlockIPs("10.0.0.1", "192.168.1.0/24"),
		middleware.WithAllowCountries("CN"),
		middleware.WithGeoIPResolver(geoResolver),
	)))

	engine.GET("/api/resource", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "access granted"})
	})

	// 1. Blacklisted CIDR 192.168.1.50 -> 403 Forbidden
	req1 := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req1.Header.Set("X-Forwarded-For", "192.168.1.50")
	w1 := httptest.NewRecorder()
	engine.ServeHTTP(w1, req1)
	require.Equal(t, http.StatusForbidden, w1.Code)

	// 2. Blocked Geo-IP US (8.8.8.8) -> 403 Forbidden
	req2 := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req2.Header.Set("X-Forwarded-For", "8.8.8.8")
	w2 := httptest.NewRecorder()
	engine.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusForbidden, w2.Code)

	// 3. Allowed Geo-IP CN (1.2.3.4) -> 200 OK
	req3 := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req3.Header.Set("X-Forwarded-For", "1.2.3.4")
	w3 := httptest.NewRecorder()
	engine.ServeHTTP(w3, req3)
	require.Equal(t, http.StatusOK, w3.Code)
}
