package middleware

import (
	"sync"
	"testing"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

// dummyMiddleware 返回一个空操作中间件，仅用于测试注册与查找。
func dummyMiddleware(name string) transportcontract.HTTPMiddleware {
	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if next != nil {
				next(c)
			}
		}
	}
}

// TestMiddlewareRegistry_RegisterAndLookup 验证基本注册与查找。
func TestMiddlewareRegistry_RegisterAndLookup(t *testing.T) {
	r := NewMiddlewareRegistry()

	// 空注册表查找应返回 false。
	_, ok := r.Lookup("auth")
	require.False(t, ok)

	// 注册后应能查找到。
	mw := dummyMiddleware("auth")
	r.Register("auth", mw)
	found, ok := r.Lookup("auth")
	require.True(t, ok)
	require.NotNil(t, found)

	// 未注册的名称应返回 false。
	_, ok = r.Lookup("logging")
	require.False(t, ok)
}

// TestMiddlewareRegistry_RegisterOverwrite 验证同名注册会覆盖。
func TestMiddlewareRegistry_RegisterOverwrite(t *testing.T) {
	r := NewMiddlewareRegistry()

	mw1 := dummyMiddleware("auth-v1")
	mw2 := dummyMiddleware("auth-v2")
	r.Register("auth", mw1)
	r.Register("auth", mw2)

	found, ok := r.Lookup("auth")
	require.True(t, ok)
	require.NotNil(t, found)
}

// TestMiddlewareRegistry_LookupAll 验证批量查找。
func TestMiddlewareRegistry_LookupAll(t *testing.T) {
	r := NewMiddlewareRegistry()

	r.Register("auth", dummyMiddleware("auth"))
	r.Register("logging", dummyMiddleware("logging"))
	// 不注册 "ratelimit"

	result := r.LookupAll([]string{"auth", "ratelimit", "logging"})
	// 应返回 2 个（跳过 ratelimit）。
	require.Len(t, result, 2)
}

// TestMiddlewareRegistry_Names 验证名称列举。
func TestMiddlewareRegistry_Names(t *testing.T) {
	r := NewMiddlewareRegistry()

	require.Empty(t, r.Names())

	r.Register("auth", dummyMiddleware("auth"))
	r.Register("logging", dummyMiddleware("logging"))

	names := r.Names()
	require.Len(t, names, 2)
}

// TestMiddlewareRegistry_ConcurrentSafety 验证并发安全。
func TestMiddlewareRegistry_ConcurrentSafety(t *testing.T) {
	r := NewMiddlewareRegistry()
	var wg sync.WaitGroup

	// 并发注册。
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r.Register("mw", dummyMiddleware("concurrent"))
		}(i)
	}

	// 并发查找。
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Lookup("mw")
		}()
	}

	wg.Wait()
	// 不 panic 即通过。
}

// TestMiddlewareRegistry_LookupAllEmpty 验证空名称列表。
func TestMiddlewareRegistry_LookupAllEmpty(t *testing.T) {
	r := NewMiddlewareRegistry()
	r.Register("auth", dummyMiddleware("auth"))

	result := r.LookupAll(nil)
	require.Empty(t, result)

	result = r.LookupAll([]string{})
	require.Empty(t, result)
}
