package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

func TestRoutePermissionCache_Basic(t *testing.T) {
	cache := NewRoutePermissionCache()
	require.Equal(t, 0, cache.Len())

	cache.Set("GET", "/api/v1/users", PermissionModeCasbin)
	mode, ok := cache.Get("get", "/api/v1/users")
	require.True(t, ok)
	require.Equal(t, PermissionModeCasbin, mode)
	require.Equal(t, 1, cache.Len())

	// Batch pre-warming
	cache.SetBatch(map[string]string{
		"POST:/api/v1/auth/login": PermissionModeIgnored,
		"GET:/api/v1/profile":     PermissionModeAuth,
	})
	require.Equal(t, 3, cache.Len())

	mode, ok = cache.Get("POST", "/api/v1/auth/login")
	require.True(t, ok)
	require.Equal(t, PermissionModeIgnored, mode)

	// Invalidate
	cache.Invalidate("get", "/api/v1/users")
	_, ok = cache.Get("GET", "/api/v1/users")
	require.False(t, ok)
	require.Equal(t, 2, cache.Len())

	// Clear
	cache.Clear()
	require.Equal(t, 0, cache.Len())
}

func TestRoutePermissionCache_Resolve_HitAndMiss(t *testing.T) {
	cache := NewRoutePermissionCache()
	var resolverCallCount int32

	resolver := func(ctx context.Context, method, path string) (string, error) {
		atomic.AddInt32(&resolverCallCount, 1)
		if path == "/dynamic/public" {
			return PermissionModeIgnored, nil
		}
		return PermissionModeCasbin, nil
	}

	// 1. First call: miss -> invokes resolver
	ctx := context.Background()
	mode, err := cache.Resolve(ctx, "GET", "/dynamic/public", resolver)
	require.NoError(t, err)
	require.Equal(t, PermissionModeIgnored, mode)
	require.Equal(t, int32(1), atomic.LoadInt32(&resolverCallCount))

	// 2. Second call: hit -> 0 resolver calls!
	mode2, err2 := cache.Resolve(ctx, "GET", "/dynamic/public", resolver)
	require.NoError(t, err2)
	require.Equal(t, PermissionModeIgnored, mode2)
	require.Equal(t, int32(1), atomic.LoadInt32(&resolverCallCount))
}

func TestRoutePermissionCache_ConcurrencyRace(t *testing.T) {
	cache := NewRoutePermissionCache()
	cache.Set("GET", "/initial", "casbin")

	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := "/route"
			if idx%2 == 0 {
				cache.Set("GET", path, "auth")
			} else {
				_, _ = cache.Get("GET", path)
			}
			_, _ = cache.Resolve(context.Background(), "POST", path, func(ctx context.Context, method, path string) (string, error) {
				return "casbin", nil
			})
		}(i)
	}
	wg.Wait()
}

func TestCachedRouteMiddleware_IgnoredMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	cache := NewRoutePermissionCache()
	cache.Set("GET", "/public/ping", PermissionModeIgnored)

	applyTransportMiddleware(router, CachedRouteMiddleware(WithRoutePermissionCache(cache)))
	router.GET("/public/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	req := httptest.NewRequest(http.MethodGet, "/public/ping", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestCachedRouteMiddleware_AuthOnlyMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache := NewRoutePermissionCache()
	cache.Set("GET", "/me", PermissionModeAuth)

	// 1. Unauthenticated request -> 401 Unauthorized
	routerUnauth := NewTestEngine()
	applyTransportMiddleware(routerUnauth, CachedRouteMiddleware(WithRoutePermissionCache(cache)))
	routerUnauth.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user": "someone"})
	})

	req1 := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec1 := httptest.NewRecorder()
	routerUnauth.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusUnauthorized, rec1.Code)

	// 2. Authenticated request with Claims -> 200 OK
	routerAuth := NewTestEngine()
	applyTransportMiddleware(routerAuth,
		func(next transportcontract.Handler) transportcontract.Handler {
			return func(c transportcontract.Context) {
				claims := &securitycontract.JWTClaims{
					SubjectID: 123,
					Roles:     []string{"user"},
				}
				c.Set(securitycontract.ContextJWTClaimsKey, claims)
				next(c)
			}
		},
		CachedRouteMiddleware(WithRoutePermissionCache(cache)),
	)
	routerAuth.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user": "someone"})
	})

	req2 := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec2 := httptest.NewRecorder()
	routerAuth.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusOK, rec2.Code)
}

func TestCachedRouteMiddleware_ContextPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewTestEngine()
	cache := NewRoutePermissionCache()
	cache.Set("GET", "/context-check", PermissionModeCasbin)

	var recordedMode string
	applyTransportMiddleware(router, CachedRouteMiddleware(WithRoutePermissionCache(cache)))
	router.GET("/context-check", func(c *gin.Context) {
		if mode, ok := RoutePermissionModeFromContext(c.Request.Context()); ok {
			recordedMode = mode
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/context-check", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, PermissionModeCasbin, recordedMode)
}
