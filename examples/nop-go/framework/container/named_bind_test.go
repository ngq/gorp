// Package container_test provides unit tests for named bindings.
//
// 适用场景：
// - 验证 NamedBind / MakeNamed 的基本语义。
// - 验证同 key 不同 name 的绑定可以独立解析。
// - 验证 IsBindNamed 行为。
package container

import (
	"errors"
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

// TestNamedBind_BasicResolution verifies that a named binding can be resolved by name and key.
//
// TestNamedBind_BasicResolution 验证命名绑定可以通过名称和 key 解析。
func TestNamedBind_BasicResolution(t *testing.T) {
	c := New()
	c.NamedBind("primary", "cache", func(runtimecontract.Container) (any, error) {
		return "redis", nil
	}, true)

	v, err := c.MakeNamed("primary", "cache")
	require.NoError(t, err)
	require.Equal(t, "redis", v)
}

// TestNamedBind_MultipleImplementations verifies that different names under the same key
// resolve to different implementations.
//
// TestNamedBind_MultipleImplementations 验证同 key 不同 name 解析到不同实现。
func TestNamedBind_MultipleImplementations(t *testing.T) {
	c := New()
	c.NamedBind("redis", "cache", func(runtimecontract.Container) (any, error) {
		return "redis-cache", nil
	}, true)
	c.NamedBind("memory", "cache", func(runtimecontract.Container) (any, error) {
		return "memory-cache", nil
	}, true)

	v1, err := c.MakeNamed("redis", "cache")
	require.NoError(t, err)
	require.Equal(t, "redis-cache", v1)

	v2, err := c.MakeNamed("memory", "cache")
	require.NoError(t, err)
	require.Equal(t, "memory-cache", v2)
}

// TestNamedBind_DoesNotAffectRegularBind verifies that NamedBind and Bind are independent.
//
// TestNamedBind_DoesNotAffectRegularBind 验证 NamedBind 和 Bind 互不影响。
func TestNamedBind_DoesNotAffectRegularBind(t *testing.T) {
	c := New()
	c.Bind("cache", func(runtimecontract.Container) (any, error) {
		return "default-cache", nil
	}, true)
	c.NamedBind("redis", "cache", func(runtimecontract.Container) (any, error) {
		return "redis-cache", nil
	}, true)

	// Regular Make resolves the default binding.
	v, err := c.Make("cache")
	require.NoError(t, err)
	require.Equal(t, "default-cache", v)

	// MakeNamed resolves the named binding.
	v2, err := c.MakeNamed("redis", "cache")
	require.NoError(t, err)
	require.Equal(t, "redis-cache", v2)
}

// TestNamedBind_NotFound verifies that MakeNamed returns an error for unbound names.
//
// TestNamedBind_NotFound 验证 MakeNamed 对未绑定的名称返回错误。
func TestNamedBind_NotFound(t *testing.T) {
	c := New()
	c.NamedBind("redis", "cache", func(runtimecontract.Container) (any, error) {
		return "redis-cache", nil
	}, true)

	_, err := c.MakeNamed("missing", "cache")
	require.Error(t, err)
}

// TestNamedBind_IsBindNamed verifies IsBindNamed behavior.
//
// TestNamedBind_IsBindNamed 验证 IsBindNamed 行为。
func TestNamedBind_IsBindNamed(t *testing.T) {
	c := New()
	c.NamedBind("redis", "cache", func(runtimecontract.Container) (any, error) {
		return "redis-cache", nil
	}, true)

	require.True(t, c.IsBindNamed("redis", "cache"))
	require.False(t, c.IsBindNamed("memory", "cache"))
	require.False(t, c.IsBindNamed("redis", "missing"))
}

// TestNamedBind_SingletonBehavior verifies that named singleton bindings cache the result.
//
// TestNamedBind_SingletonBehavior 验证命名单例绑定缓存结果。
func TestNamedBind_SingletonBehavior(t *testing.T) {
	c := New()
	calls := 0
	c.NamedBind("primary", "svc", func(runtimecontract.Container) (any, error) {
		calls++
		return calls, nil
	}, true)

	v1, err := c.MakeNamed("primary", "svc")
	require.NoError(t, err)
	v2, err := c.MakeNamed("primary", "svc")
	require.NoError(t, err)
	require.Equal(t, v1, v2)
	require.Equal(t, 1, calls)
}

// TestNamedBind_TransientBehavior verifies that named transient bindings create fresh instances.
//
// TestNamedBind_TransientBehavior 验证命名瞬态绑定每次创建新实例。
func TestNamedBind_TransientBehavior(t *testing.T) {
	c := New()
	calls := 0
	c.NamedBind("primary", "svc", func(runtimecontract.Container) (any, error) {
		calls++
		return calls, nil
	}, false)

	v1, err := c.MakeNamed("primary", "svc")
	require.NoError(t, err)
	v2, err := c.MakeNamed("primary", "svc")
	require.NoError(t, err)
	require.Equal(t, 1, v1)
	require.Equal(t, 2, v2)
}

// TestNamedBind_CircularDependency verifies that circular dependencies in named
// singleton bindings are detected.
//
// TestNamedBind_CircularDependency 验证命名单例绑定中的循环依赖被检测到。
func TestNamedBind_CircularDependency(t *testing.T) {
	c := New()
	c.NamedBind("primary", "svc", func(c runtimecontract.Container) (any, error) {
		return c.MakeNamed("primary", "svc")
	}, true)

	_, err := c.MakeNamed("primary", "svc")
	require.Error(t, err)
	require.True(t, errors.Is(err, runtimecontract.ErrCircularDependency),
		"expected CircularDependencyError, got: %v", err)
}
