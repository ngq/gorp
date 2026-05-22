// Package container_test provides unit tests for container rule resolution.
//
// 适用场景：
// - 验证容器规则解析器的注册、匹配和执行行为。
package container

import (
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

// TestContainer_IsBindRecognizesDeferredKeyBeforeLoad verifies that IsBind
// returns true for a deferred provider's key before it is loaded.
//
// TestContainer_IsBindRecognizesDeferredKeyBeforeLoad 验证 IsBind 在延迟服务
// 提供商的 key 被加载前返回 true。
func TestContainer_IsBindRecognizesDeferredKeyBeforeLoad(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := &deferredProvider{
		name:   "deferred-user",
		keys:   []string{"user.repo"},
		loaded: &loaded,
		booted: &booted,
		value:  "ok",
	}

	require.NoError(t, c.RegisterProvider(p))
	require.True(t, c.IsBind("user.repo"))
	require.Equal(t, 0, loaded)
	require.Equal(t, 0, booted)

	v, err := c.Make("user.repo")
	require.NoError(t, err)
	require.Equal(t, "ok", v)
	require.Equal(t, 1, loaded)
	require.Equal(t, 1, booted)
}

// TestContainer_MakeLoadsDeferredProviderOnlyOnce verifies that a deferred
// provider is loaded and booted exactly once regardless of how many times
// its keys are requested via Make.
//
// TestContainer_MakeLoadsDeferredProviderOnlyOnce 验证延迟服务提供商只会被
// 加载和引导一次，无论通过 Make 请求多少次。
func TestContainer_MakeLoadsDeferredProviderOnlyOnce(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := &deferredProvider{
		name:   "deferred-once",
		keys:   []string{"cache.client"},
		loaded: &loaded,
		booted: &booted,
		value:  "cache",
	}

	require.NoError(t, c.RegisterProvider(p))

	_, err := c.Make("cache.client")
	require.NoError(t, err)
	_, err = c.Make("cache.client")
	require.NoError(t, err)

	require.Equal(t, 1, loaded)
	require.Equal(t, 1, booted)
}

// TestContainer_MustMakePanicsForUnknownKey verifies that MustMake panics
// when the requested key is not bound in the container.
//
// TestContainer_MustMakePanicsForUnknownKey 验证 MustMake 在请求的 key 未绑定
// 到容器时发生 panic。
func TestContainer_MustMakePanicsForUnknownKey(t *testing.T) {
	c := New()
	require.Panics(t, func() {
		_ = c.MustMake("missing.key")
	})
}

// TestContainer_BindNonSingletonCreatesFreshInstance verifies that binding
// a non-singleton (transient) key produces a fresh instance on each Make call.
//
// TestContainer_BindNonSingletonCreatesFreshInstance 验证绑定非单例（瞬态）key
// 会在每次 Make 调用时产生新实例。
func TestContainer_BindNonSingletonCreatesFreshInstance(t *testing.T) {
	c := New()
	count := 0
	c.Bind("transient.counter", func(runtimecontract.Container) (any, error) {
		count++
		return count, nil
	}, false)

	v1, err := c.Make("transient.counter")
	require.NoError(t, err)
	v2, err := c.Make("transient.counter")
	require.NoError(t, err)

	require.Equal(t, 1, v1)
	require.Equal(t, 2, v2)
}
