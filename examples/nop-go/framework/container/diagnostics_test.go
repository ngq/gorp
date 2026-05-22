// Package container_test provides unit tests for container diagnostics.
//
// 适用场景：
// - 验证 RegisteredProviders 返回正确的 provider 状态信息。
// - 验证 DebugPrint 输出包含关键信息。
package container

import (
	"strings"
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

// TestRegisteredProviders_EmptyContainer verifies that an empty container has no providers.
//
// TestRegisteredProviders_EmptyContainer 验证空容器没有 provider。
func TestRegisteredProviders_EmptyContainer(t *testing.T) {
	c := New()
	infos := c.RegisteredProviders()
	require.Empty(t, infos)
}

// TestRegisteredProviders_WithProviders verifies that RegisteredProviders returns correct info.
//
// TestRegisteredProviders_WithProviders 验证 RegisteredProviders 返回正确信息。
func TestRegisteredProviders_WithProviders(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := &testProvider{deferLoad: false, provide: []string{"a"}, loaded: &loaded, booted: &booted}
	require.NoError(t, c.RegisterProvider(p))

	infos := c.RegisteredProviders()
	require.Len(t, infos, 1)
	require.Equal(t, "test", infos[0].Name)
	require.True(t, infos[0].Loaded)
	require.True(t, infos[0].Booted)
	require.False(t, infos[0].IsDefer)
}

// TestRegisteredProviders_DeferredProvider verifies IsDefer flag for deferred providers.
//
// TestRegisteredProviders_DeferredProvider 验证延迟 provider 的 IsDefer 标志。
func TestRegisteredProviders_DeferredProvider(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := &deferredProvider{name: "deferred-svc", keys: []string{"svc"}, loaded: &loaded, booted: &booted, value: "ok"}
	require.NoError(t, c.RegisterProvider(p))

	infos := c.RegisteredProviders()
	require.Len(t, infos, 1)
	require.True(t, infos[0].IsDefer)
	require.False(t, infos[0].Loaded)
	require.False(t, infos[0].Booted)
}

// TestDebugPrint_ContainsBindingInfo verifies that DebugPrint includes binding information.
//
// TestDebugPrint_ContainsBindingInfo 验证 DebugPrint 包含绑定信息。
func TestDebugPrint_ContainsBindingInfo(t *testing.T) {
	c := New()
	c.Bind("my-service", func(runtimecontract.Container) (any, error) {
		return "value", nil
	}, true)
	c.Bind("my-transient", func(runtimecontract.Container) (any, error) {
		return "value", nil
	}, false)

	output := c.DebugPrint()
	require.Contains(t, output, "my-service")
	require.Contains(t, output, "singleton")
	require.Contains(t, output, "my-transient")
	require.Contains(t, output, "transient")
}

// TestDebugPrint_ContainsNamedBindings verifies that DebugPrint includes named bindings.
//
// TestDebugPrint_ContainsNamedBindings 验证 DebugPrint 包含命名绑定信息。
func TestDebugPrint_ContainsNamedBindings(t *testing.T) {
	c := New()
	c.NamedBind("redis", "cache", func(runtimecontract.Container) (any, error) {
		return "redis-cache", nil
	}, true)

	output := c.DebugPrint()
	require.Contains(t, output, "redis")
	require.Contains(t, output, "cache")
}

// TestDebugPrint_ContainsProviderInfo verifies that DebugPrint includes provider information.
//
// TestDebugPrint_ContainsProviderInfo 验证 DebugPrint 包含 provider 信息。
func TestDebugPrint_ContainsProviderInfo(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := &testProvider{deferLoad: false, provide: []string{"a"}, loaded: &loaded, booted: &booted}
	require.NoError(t, c.RegisterProvider(p))

	output := c.DebugPrint()
	require.Contains(t, output, "test")
}

// TestDebugPrint_ContainsDestroyState verifies that DebugPrint shows destroyed state.
//
// TestDebugPrint_ContainsDestroyState 验证 DebugPrint 显示销毁状态。
func TestDebugPrint_ContainsDestroyState(t *testing.T) {
	c := New()
	output := c.DebugPrint()
	require.Contains(t, output, "Destroyed: false")

	require.NoError(t, c.Destroy())
	output = c.DebugPrint()
	require.Contains(t, output, "Destroyed: true")
}

// TestDebugPrint_ContainsCloserInfo verifies that DebugPrint shows registered closers.
//
// TestDebugPrint_ContainsCloserInfo 验证 DebugPrint 显示注册的 closer。
func TestDebugPrint_ContainsCloserInfo(t *testing.T) {
	c := New()
	c.RegisterCloser("db", &closeFunc{})
	c.RegisterCloser("redis", &closeFunc{})

	output := c.DebugPrint()
	require.Contains(t, output, "db")
	require.Contains(t, output, "redis")
}

// TestDebugPrint_ContainsDeferredKeys verifies that DebugPrint shows deferred key mappings.
//
// TestDebugPrint_ContainsDeferredKeys 验证 DebugPrint 显示延迟 key 映射。
func TestDebugPrint_ContainsDeferredKeys(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := &deferredProvider{name: "deferred-svc", keys: []string{"svc.key"}, loaded: &loaded, booted: &booted, value: "ok"}
	require.NoError(t, c.RegisterProvider(p))

	output := c.DebugPrint()
	require.True(t, strings.Contains(output, "svc.key") || strings.Contains(output, "Deferred"))
}
