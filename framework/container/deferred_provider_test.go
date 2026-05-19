// Package container_test provides unit tests for deferred provider loading behavior.
//
// 适用场景：
// - 验证延迟（Defer）服务提供商的注册与延迟加载时机。
package container

import (
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

type deferredProvider struct {
	name   string
	keys   []string
	loaded *int
	booted *int
	value  string
}

func (p *deferredProvider) Name() string        { return p.name }
func (p *deferredProvider) IsDefer() bool       { return true }
func (p *deferredProvider) Provides() []string  { return p.keys }
func (p *deferredProvider) DependsOn() []string { return nil }
func (p *deferredProvider) Register(c runtimecontract.Container) error {
	*p.loaded++
	for _, key := range p.keys {
		value := p.value
		bindKey := key
		c.Bind(bindKey, func(runtimecontract.Container) (any, error) { return value, nil }, true)
	}
	return nil
}
func (p *deferredProvider) Boot(runtimecontract.Container) error {
	*p.booted++
	return nil
}

// TestDeferredProviderUsesFirstRegistrantForSameKey verifies that when multiple
// deferred providers register the same key, the first registrant wins.
//
// TestDeferredProviderUsesFirstRegistrantForSameKey 验证当多个延迟服务提供商
// 注册相同的 key 时，先注册者获胜。
func TestDeferredProviderUsesFirstRegistrantForSameKey(t *testing.T) {
	c := New()
	loadedA, bootedA := 0, 0
	loadedB, bootedB := 0, 0

	p1 := &deferredProvider{name: "p1", keys: []string{"shared.key"}, loaded: &loadedA, booted: &bootedA, value: "from-p1"}
	p2 := &deferredProvider{name: "p2", keys: []string{"shared.key"}, loaded: &loadedB, booted: &bootedB, value: "from-p2"}

	require.NoError(t, c.RegisterProvider(p1))
	require.NoError(t, c.RegisterProvider(p2))

	value, err := c.Make("shared.key")
	require.NoError(t, err)
	require.Equal(t, "from-p1", value)
	require.Equal(t, 1, loadedA)
	require.Equal(t, 1, bootedA)
	require.Equal(t, 0, loadedB)
	require.Equal(t, 0, bootedB)
}
