// Application scenarios:
// - Cache route permission modes in memory to avoid querying the database on every HTTP request.
// - High-concurrency authorization decision pipeline with pre-warming and dynamic invalidation.
// - Provide fine-grained access mode resolution (ignored, auth-only, RBAC/Casbin).
//
// 适用场景：
// - 在内存中高速缓存路由权限模式，避免每个 HTTP 请求都直连数据库查路由表。
// - 支持预热与动态失效的高并发鉴权决策管道。
// - 提供细粒度路由访问模式判定（公开免鉴权、仅需登录态、细粒度 RBAC/Casbin）。
package middleware

import (
	"context"
	"strings"
	"sync"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Standard permission modes.
const (
	// PermissionModeIgnored indicates public routes that bypass authentication and authorization.
	PermissionModeIgnored = "ignored"

	// PermissionModeAuth indicates routes that only require valid authentication (logged-in), without role/permission checks.
	PermissionModeAuth = "auth"

	// PermissionModeCasbin indicates routes that require fine-grained Casbin/RBAC policy evaluation.
	PermissionModeCasbin = "casbin"
)

type routeModeContextKey struct{}

// WithRoutePermissionMode injects the resolved permission mode into context.
func WithRoutePermissionMode(ctx context.Context, mode string) context.Context {
	return context.WithValue(ctx, routeModeContextKey{}, mode)
}

// RoutePermissionModeFromContext retrieves the resolved permission mode from context.
func RoutePermissionModeFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	val, ok := ctx.Value(routeModeContextKey{}).(string)
	return val, ok
}

// RouteModeResolver resolves the permission mode for a given method and path from an external source (e.g. database).
type RouteModeResolver func(ctx context.Context, method, path string) (string, error)

// RoutePermissionCache provides thread-safe in-memory caching of route permission modes.
type RoutePermissionCache struct {
	mu      sync.RWMutex
	entries map[string]string
}

// NewRoutePermissionCache creates an empty route permission cache.
func NewRoutePermissionCache() *RoutePermissionCache {
	return &RoutePermissionCache{
		entries: make(map[string]string),
	}
}

func routeKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + ":" + strings.TrimSpace(path)
}

// Get retrieves the cached permission mode for a route.
func (c *RoutePermissionCache) Get(method, path string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	mode, ok := c.entries[routeKey(method, path)]
	return mode, ok
}

// Set sets the permission mode for a route.
func (c *RoutePermissionCache) Set(method, path, mode string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[routeKey(method, path)] = mode
}

// SetBatch pre-warms the cache with multiple route entries.
func (c *RoutePermissionCache) SetBatch(routes map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range routes {
		// Key can be "GET:/api/v1/users" or method, path
		parts := strings.SplitN(k, ":", 2)
		if len(parts) == 2 {
			c.entries[routeKey(parts[0], parts[1])] = v
		} else {
			c.entries[k] = v
		}
	}
}

// Invalidate removes a route from the cache.
func (c *RoutePermissionCache) Invalidate(method, path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, routeKey(method, path))
}

// Clear flushes all cached route entries.
func (c *RoutePermissionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]string)
}

// Len returns the number of cached routes.
func (c *RoutePermissionCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Resolve returns the permission mode from cache, or falls back to resolver and caches the result.
func (c *RoutePermissionCache) Resolve(ctx context.Context, method, path string, fallback RouteModeResolver) (string, error) {
	if mode, ok := c.Get(method, path); ok {
		return mode, nil
	}

	if fallback == nil {
		return "", nil
	}

	mode, err := fallback(ctx, method, path)
	if err != nil {
		return "", err
	}

	if mode != "" {
		c.Set(method, path, mode)
	}
	return mode, nil
}

// RouteCacheOptions configures CachedRouteAuthorizer.
type RouteCacheOptions struct {
	Cache         *RoutePermissionCache
	Resolver      RouteModeResolver
	IgnoredModes  []string
	AuthOnlyModes []string
}

// RouteCacheOption configures RouteCacheOptions.
type RouteCacheOption func(*RouteCacheOptions)

// WithRoutePermissionCache provides a custom cache instance.
func WithRoutePermissionCache(cache *RoutePermissionCache) RouteCacheOption {
	return func(o *RouteCacheOptions) {
		o.Cache = cache
	}
}

// WithRouteModeResolver provides a resolver function for cache misses.
func WithRouteModeResolver(resolver RouteModeResolver) RouteCacheOption {
	return func(o *RouteCacheOptions) {
		o.Resolver = resolver
	}
}

// WithIgnoredModes specifies modes that allow immediate bypass without authentication.
func WithIgnoredModes(modes ...string) RouteCacheOption {
	return func(o *RouteCacheOptions) {
		o.IgnoredModes = append(o.IgnoredModes, modes...)
	}
}

// WithAuthOnlyModes specifies modes that only require authentication without RBAC checks.
func WithAuthOnlyModes(modes ...string) RouteCacheOption {
	return func(o *RouteCacheOptions) {
		o.AuthOnlyModes = append(o.AuthOnlyModes, modes...)
	}
}

// CachedRouteMiddleware creates a transport middleware that resolves and caches route permission modes,
// completely eliminating per-request database lookups.
func CachedRouteMiddleware(opts ...RouteCacheOption) transportcontract.Middleware {
	cfg := RouteCacheOptions{
		Cache:         NewRoutePermissionCache(),
		IgnoredModes:  []string{PermissionModeIgnored, "public", "whitelist"},
		AuthOnlyModes: []string{PermissionModeAuth, "login_required"},
	}
	for _, o := range opts {
		o(&cfg)
	}

	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			req := c.Request()
			if req == nil {
				if next != nil {
					next(c)
				}
				return
			}

			method := req.Method
			path := c.RoutePath()
			if path == "" {
				path = req.URL.Path
			}

			ctx := req.Context()
			mode, err := cfg.Cache.Resolve(ctx, method, path, cfg.Resolver)
			if err != nil {
				respondForbidden(c, "route permission resolution error")
				return
			}

			// Injected resolved mode into request context for downstream handlers
			if mode != "" {
				*req = *req.WithContext(WithRoutePermissionMode(req.Context(), mode))
			}

			// 1. Ignored / Public modes -> bypass immediately
			for _, m := range cfg.IgnoredModes {
				if strings.EqualFold(mode, m) {
					if next != nil {
						next(c)
					}
					return
				}
			}

			// 2. Auth-only modes -> require valid JWT claims then proceed
			for _, m := range cfg.AuthOnlyModes {
				if strings.EqualFold(mode, m) {
					if _, ok := claimsFromContext(c); !ok {
						respondUnauthorized(c, "authentication required")
						return
					}
					if next != nil {
						next(c)
					}
					return
				}
			}

			// 3. Casbin or default mode -> proceed downstream (e.g. to Authorize / Casbin middleware)
			if next != nil {
				next(c)
			}
		}
	}
}
