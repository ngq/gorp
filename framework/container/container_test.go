// Package container_test provides unit tests for the service container.
//
// 适用场景：
// - 验证容器的注册、引导、延迟加载和依赖解析行为。
// - 验证服务提供商的加载顺序与错误传播。
package container

import (
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

type testProvider struct {
	deferLoad bool
	provide   []string
	loaded    *int
	booted    *int
}

func (p *testProvider) Name() string       { return "test" }
func (p *testProvider) IsDefer() bool      { return p.deferLoad }
func (p *testProvider) Provides() []string { return p.provide }
func (p *testProvider) DependsOn() []string { return nil }
func (p *testProvider) Register(c runtimecontract.Container) error {
	*p.loaded++
	for _, k := range p.provide {
		key := k
		c.Bind(key, func(runtimecontract.Container) (any, error) { return "ok", nil }, true)
	}
	return nil
}
func (p *testProvider) Boot(runtimecontract.Container) error {
	*p.booted++
	return nil
}

// TestContainer_NonDeferredProviderLoadsImmediately verifies that a non-deferred
// provider is loaded and booted immediately upon registration.
//
// TestContainer_NonDeferredProviderLoadsImmediately 验证非延迟服务提供商
// 在注册时立即加载和引导。
func TestContainer_NonDeferredProviderLoadsImmediately(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := testProvider{deferLoad: false, provide: []string{"a"}, loaded: &loaded, booted: &booted}

	err := c.RegisterProvider(&p)
	require.NoError(t, err)
	require.Equal(t, 1, loaded)
	require.Equal(t, 1, booted)

	v, err := c.Make("a")
	require.NoError(t, err)
	require.Equal(t, "ok", v)
}

// TestContainer_DeferredProviderLoadsOnFirstMake verifies that a deferred
// provider is loaded and booted only when its key is first requested via Make.
//
// TestContainer_DeferredProviderLoadsOnFirstMake 验证延迟服务提供商
// 仅在首次通过 Make 请求其 key 时才加载和引导。
func TestContainer_DeferredProviderLoadsOnFirstMake(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := testProvider{deferLoad: true, provide: []string{"a"}, loaded: &loaded, booted: &booted}

	err := c.RegisterProvider(&p)
	require.NoError(t, err)
	require.Equal(t, 0, loaded)
	require.Equal(t, 0, booted)

	v, err := c.Make("a")
	require.NoError(t, err)
	require.Equal(t, "ok", v)
	require.Equal(t, 1, loaded)
	require.Equal(t, 1, booted)

	// second Make should not boot again
	_, err = c.Make("a")
	require.NoError(t, err)
	require.Equal(t, 1, loaded)
	require.Equal(t, 1, booted)
}
