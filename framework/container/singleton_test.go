// Package container_test provides unit tests for singleton lifecycle and self-reference safety.
//
// 适用场景：
// - 验证单例绑定的状态机正确性。
// - 验证 sync.Once 替换后不再出现 fatal deadlock。
// - 验证并发安全、错误缓存、可重入检测。
package container

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

// TestSingleton_BasicResolution verifies that a singleton binding resolves to the same instance.
//
// TestSingleton_BasicResolution 验证单例绑定解析返回同一实例。
func TestSingleton_BasicResolution(t *testing.T) {
	c := New()
	calls := atomic.Int32{}
	c.Bind("svc", func(runtimecontract.Container) (any, error) {
		calls.Add(1)
		return "value", nil
	}, true)

	v1, err := c.Make("svc")
	require.NoError(t, err)
	v2, err := c.Make("svc")
	require.NoError(t, err)
	require.Equal(t, v1, v2)
	require.Equal(t, int32(1), calls.Load())
}

// TestSingleton_FactoryErrorNotCached verifies that a singleton factory error
// is NOT cached — subsequent Make calls will retry the factory.
// This is the correct behavior: caching an error would permanently prevent
// recovery from transient failures.
//
// TestSingleton_FactoryErrorNotCached 验证单例工厂错误不被缓存——
// 后续 Make 调用将重试工厂。
// 这是正确的行为：缓存错误会永久阻止从瞬态故障中恢复。
func TestSingleton_FactoryErrorNotCached(t *testing.T) {
	c := New()
	calls := atomic.Int32{}
	c.Bind("failing", func(runtimecontract.Container) (any, error) {
		calls.Add(1)
		return nil, errors.New("factory failed")
	}, true)

	_, err := c.Make("failing")
	require.Error(t, err)
	require.Equal(t, "factory failed", err.Error())

	_, err = c.Make("failing")
	require.Error(t, err)
	// Factory is called again because error is not cached.
	// 工厂再次被调用，因为错误不被缓存。
	require.Equal(t, int32(2), calls.Load())
}

// TestSingleton_SelfReferenceReturnsError verifies that a singleton factory
// calling Make on its own key returns CircularDependencyError instead of fatal deadlock.
//
// TestSingleton_SelfReferenceReturnsError 验证单例工厂内调用 Make 同一 key
// 返回 CircularDependencyError 而不是 fatal deadlock。
func TestSingleton_SelfReferenceReturnsError(t *testing.T) {
	c := New()
	c.Bind("self-ref", func(c runtimecontract.Container) (any, error) {
		return c.Make("self-ref")
	}, true)

	_, err := c.Make("self-ref")
	require.Error(t, err)

	var cde *runtimecontract.CircularDependencyError
	require.True(t, errors.As(err, &cde), "expected CircularDependencyError, got: %v", err)
	require.True(t, errors.Is(err, runtimecontract.ErrCircularDependency))
	require.Equal(t, "self-ref", cde.Key)
	require.Contains(t, cde.Chain, "self-ref")
}

// TestSingleton_IndirectCircularDependency verifies that A→B→A circular chain
// is detected and returns a friendly error.
//
// TestSingleton_IndirectCircularDependency 验证 A→B→A 间接循环依赖被检测到
// 并返回友好错误。
func TestSingleton_IndirectCircularDependency(t *testing.T) {
	c := New()
	c.Bind("a", func(c runtimecontract.Container) (any, error) {
		return c.Make("b")
	}, true)
	c.Bind("b", func(c runtimecontract.Container) (any, error) {
		return c.Make("a")
	}, true)

	_, err := c.Make("a")
	require.Error(t, err)

	var cde *runtimecontract.CircularDependencyError
	require.True(t, errors.As(err, &cde), "expected CircularDependencyError, got: %v", err)
	require.True(t, errors.Is(err, runtimecontract.ErrCircularDependency))
	// Chain should contain the full cycle
	require.Contains(t, cde.Chain, "a")
	require.Contains(t, cde.Chain, "b")
}

// TestSingleton_ThreeWayCircularDependency verifies A→B→C→A detection.
//
// TestSingleton_ThreeWayCircularDependency 验证 A→B→C→A 三方循环检测。
func TestSingleton_ThreeWayCircularDependency(t *testing.T) {
	c := New()
	c.Bind("a", func(c runtimecontract.Container) (any, error) {
		return c.Make("b")
	}, true)
	c.Bind("b", func(c runtimecontract.Container) (any, error) {
		return c.Make("c")
	}, true)
	c.Bind("c", func(c runtimecontract.Container) (any, error) {
		return c.Make("a")
	}, true)

	_, err := c.Make("a")
	require.Error(t, err)

	var cde *runtimecontract.CircularDependencyError
	require.True(t, errors.As(err, &cde))
	require.True(t, errors.Is(err, runtimecontract.ErrCircularDependency))
}

// TestSingleton_ConcurrentResolution verifies that multiple goroutines
// resolving the same singleton key all get the same instance.
//
// TestSingleton_ConcurrentResolution 验证多个 goroutine 并发解析同一单例 key
// 都获得同一实例。
func TestSingleton_ConcurrentResolution(t *testing.T) {
	c := New()
	var calls atomic.Int32
	c.Bind("shared", func(runtimecontract.Container) (any, error) {
		calls.Add(1)
		return "singleton-value", nil
	}, true)

	const n = 100
	var wg sync.WaitGroup
	results := make([]any, n)
	errs := make([]error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = c.Make("shared")
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		require.NoError(t, errs[i])
		require.Equal(t, "singleton-value", results[i])
	}
	require.Equal(t, int32(1), calls.Load())
}

// TestSingleton_ConcurrentResolutionDifferentKeys verifies that concurrent
// resolution of different singleton keys works correctly without false circular errors.
//
// TestSingleton_ConcurrentResolutionDifferentKeys 验证并发解析不同单例 key
// 不会产生虚假的循环依赖错误。
func TestSingleton_ConcurrentResolutionDifferentKeys(t *testing.T) {
	c := New()
	c.Bind("x", func(runtimecontract.Container) (any, error) {
		return "x-value", nil
	}, true)
	c.Bind("y", func(runtimecontract.Container) (any, error) {
		return "y-value", nil
	}, true)

	const n = 50
	var wg sync.WaitGroup
	errs := make([]error, n*2)

	for i := 0; i < n; i++ {
		wg.Add(2)
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = c.Make("x")
		}(i)
		go func(idx int) {
			defer wg.Done()
			_, errs[n+idx] = c.Make("y")
		}(i)
	}
	wg.Wait()

	for _, err := range errs {
		require.NoError(t, err)
	}
}

// TestTransient_CircularDependency verifies that circular dependencies in transient
// bindings are also detected and returned as friendly errors.
//
// TestTransient_CircularDependency 验证 transient 绑定中的循环依赖也会被检测并返回友好错误。
func TestTransient_CircularDependency(t *testing.T) {
	c := New()
	c.Bind("a", func(c runtimecontract.Container) (any, error) {
		return c.Make("b")
	}, false)
	c.Bind("b", func(c runtimecontract.Container) (any, error) {
		return c.Make("a")
	}, false)

	_, err := c.Make("a")
	require.Error(t, err)

	var cde *runtimecontract.CircularDependencyError
	require.True(t, errors.As(err, &cde), "expected CircularDependencyError, got: %v", err)
	require.True(t, errors.Is(err, runtimecontract.ErrCircularDependency))
	require.Contains(t, cde.Chain, "a")
	require.Contains(t, cde.Chain, "b")
}
