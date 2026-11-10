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
	"time"

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
const (
	maxRoutePermissionCacheEntries = 4096
	defaultRoutePermissionCacheTTL = 5 * time.Minute
)

type routePermissionEntry struct {
	mode      string
	expiresAt time.Time
}

type RoutePermissionCache struct {
	mu      sync.RWMutex
	entries map[string]routePermissionEntry
	order   []string
	ttl     time.Duration
}

// NewRoutePermissionCache creates an empty route permission cache.
func NewRoutePermissionCache() *RoutePermissionCache {
	return NewRoutePermissionCacheWithTTL(defaultRoutePermissionCacheTTL)
}

// NewRoutePermissionCacheWithTTL creates a route permission cache with an explicit entry TTL.
func NewRoutePermissionCacheWithTTL(ttl time.Duration) *RoutePermissionCache {
	if ttl <= 0 {
		ttl = defaultRoutePermissionCacheTTL
	}
	return &RoutePermissionCache{
		entries: make(map[string]routePermissionEntry),
		ttl:     ttl,
	}
}

func routeKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + ":" + strings.TrimSpace(path)
}

// Get retrieves the cached permission mode for a route.
func (c *RoutePermissionCache) Get(method, path string) (string, bool) {
	key := routeKey(method, path)
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return "", false
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		if current, exists := c.entries[key]; exists && current.expiresAt == entry.expiresAt {
			delete(c.entries, key)
			c.removeOrderLocked(key)
		}
		c.mu.Unlock()
		return "", false
	}
	return entry.mode, true
}

// Set sets the permission mode for a route.
func (c *RoutePermissionCache) Set(method, path, mode string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setLocked(routeKey(method, path), mode)
}

func (c *RoutePermissionCache) setLocked(key, mode string) {
	if c.entries == nil {
		c.entries = make(map[string]routePermissionEntry)
	}
	if c.ttl <= 0 {
		c.ttl = defaultRoutePermissionCacheTTL
	}
	if _, exists := c.entries[key]; !exists {
		if len(c.entries) >= maxRoutePermissionCacheEntries && len(c.order) > 0 {
			oldest := c.order[0]
			c.order = c.order[1:]
			delete(c.entries, oldest)
		}
		c.order = append(c.order, key)
	}
	c.entries[key] = routePermissionEntry{mode: mode, expiresAt: time.Now().Add(c.ttl)}
}

func (c *RoutePermissionCache) removeOrderLocked(key string) {
	for i, candidate := range c.order {
		if candidate == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			return
		}
	}
}

// SetBatch pre-warms the cache with multiple route entries.
func (c *RoutePermissionCache) SetBatch(routes map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range routes {
		// Key can be "GET:/api/v1/users" or method, path
		parts := strings.SplitN(k, ":", 2)
		if len(parts) == 2 {
			c.setLocked(routeKey(parts[0], parts[1]), v)
		} else {
			c.setLocked(k, v)
		}
	}
}

// Invalidate removes a route from the cache.
func (c *RoutePermissionCache) Invalidate(method, path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := routeKey(method, path)
	delete(c.entries, key)
	c.removeOrderLocked(key)
}

// Clear flushes all cached route entries.
func (c *RoutePermissionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]routePermissionEntry)
	c.order = nil
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
