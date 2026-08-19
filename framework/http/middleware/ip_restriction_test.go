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
		// 信任 localhost 代理，测试中 httptest 的 RemoteAddr 默认为 192.0.2.1:1234
		middleware.WithTrustedProxies("192.0.2.0/24"),
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

// TestIPRestrictionMiddleware_SpoofingProtection 验证不受信任的来源伪造 XFF 被正确忽略。
func TestIPRestrictionMiddleware_SpoofingProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(ginprovider.AdaptMiddleware(middleware.IPRestrictionMiddleware(
		middleware.WithAllowIPs("10.10.10.10"),
		// 不信任默认 httptest 的 RemoteAddr (192.0.2.1)
		// 默认 TrustedProxies 只有 127.0.0.1 和 ::1
	)))

	engine.GET("/api/resource", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

	// 攻击者伪造 XFF 想冒充白名单 IP 10.10.10.10，但来源不受信任 -> 使用 RemoteAddr 判定 -> 403
	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req.Header.Set("X-Forwarded-For", "10.10.10.10")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}
