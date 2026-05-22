// Package middleware provides the default MiddlewareRegistry implementation.
// The registry maps named middleware instances (e.g., "auth", "logging") to
// transportcontract.Middleware, enabling proto-annotation-driven automatic mounting.
//
// 本包提供默认的 MiddlewareRegistry 实现。
// 注册表将具名中间件实例（如 "auth"、"logging"）映射到
// transportcontract.Middleware，支持 proto 注解驱动的自动挂载。
package middleware

import (
	"sync"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// defaultMiddlewareRegistry 是 MiddlewareRegistry 的默认实现。
// 使用 sync.RWMutex 保证并发安全。
// 支持注册、查找、批量查找、名称列举。
type defaultMiddlewareRegistry struct {
	mu    sync.RWMutex
	store map[string]transportcontract.Middleware
}

// NewMiddlewareRegistry creates a new empty middleware registry.
//
// NewMiddlewareRegistry 创建一个新的空中间件注册表。
func NewMiddlewareRegistry() transportcontract.MiddlewareRegistry {
	return &defaultMiddlewareRegistry{
		store: make(map[string]transportcontract.Middleware),
	}
}

// Register adds a named middleware to the registry.
// If a middleware with the same name already exists, it is replaced.
//
// Register 向注册表添加具名中间件。
// 如果同名中间件已存在，则替换。
func (r *defaultMiddlewareRegistry) Register(name string, middleware transportcontract.Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[name] = middleware
}

// Lookup retrieves a middleware by name.
// Returns the middleware and true if found, nil and false otherwise.
//
// Lookup 按名称查找中间件。
// 找到时返回中间件和 true，否则返回 nil 和 false。
func (r *defaultMiddlewareRegistry) Lookup(name string) (transportcontract.Middleware, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	mw, ok := r.store[name]
	return mw, ok
}

// LookupAll retrieves multiple middleware by names.
// Returns found middleware in order; silently skips unknown names.
//
// LookupAll 按名称列表查找多个中间件。
// 按顺序返回找到的中间件；静默跳过未知名称。
func (r *defaultMiddlewareRegistry) LookupAll(names []string) []transportcontract.Middleware {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]transportcontract.Middleware, 0, len(names))
	for _, name := range names {
		if mw, ok := r.store[name]; ok {
			result = append(result, mw)
		}
	}
	return result
}

// Names returns all registered middleware names.
//
// Names 返回所有已注册的中间件名称。
func (r *defaultMiddlewareRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.store))
	for name := range r.store {
		names = append(names, name)
	}
	return names
}
