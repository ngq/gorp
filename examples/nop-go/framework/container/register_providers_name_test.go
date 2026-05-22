// Package container_test provides unit tests for service provider name and key binding.
//
// 适用场景：
// - 验证容器对服务商名称的注册、识别与 key 绑定行为。
package container

import (
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

// TestNewContainerBindsContainerKey verifies that a newly created container
// automatically binds the container key to itself.
//
// TestNewContainerBindsContainerKey 验证新创建的容器自动将自身绑定到容器 key。
func TestNewContainerBindsContainerKey(t *testing.T) {
	c := New()
	v, err := c.Make(runtimecontract.ContainerKey)
	require.NoError(t, err)
	require.Equal(t, c, v)
}

// TestRegisterProviderRejectsEmptyName verifies that registering a provider
// with an empty name returns an error.
//
// TestRegisterProviderRejectsEmptyName 验证注册空名称的服务提供商会返回错误。
func TestRegisterProviderRejectsEmptyName(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := testProvider{deferLoad: false, provide: []string{"a"}, loaded: &loaded, booted: &booted}

	// 覆盖 Name() 为空的最小变体
	empty := &struct{ testProvider }{p}
	err := c.RegisterProvider(serviceProviderWithName(empty, ""))
	require.EqualError(t, err, "provider name is empty")
}

// TestRegisterProviderRejectsDuplicateName verifies that registering a provider
// with a duplicate name returns an error and does not execute the duplicate.
//
// TestRegisterProviderRejectsDuplicateName 验证注册重名服务提供商会返回错误
// 且不会执行重复的提供商。
func TestRegisterProviderRejectsDuplicateName(t *testing.T) {
	c := New()
	loaded1, booted1 := 0, 0
	loaded2, booted2 := 0, 0
	p1 := &testProvider{deferLoad: false, provide: []string{"a"}, loaded: &loaded1, booted: &booted1}
	p2 := &testProvider{deferLoad: false, provide: []string{"b"}, loaded: &loaded2, booted: &booted2}

	require.NoError(t, c.RegisterProvider(serviceProviderWithName(p1, "dup")))
	err := c.RegisterProvider(serviceProviderWithName(p2, "dup"))
	require.EqualError(t, err, "provider already registered: dup")
	require.Equal(t, 1, loaded1)
	require.Equal(t, 1, booted1)
	require.Equal(t, 0, loaded2)
	require.Equal(t, 0, booted2)
}

type namedProviderWrapper struct {
	runtimecontract.ServiceProvider
	name string
}

func (p *namedProviderWrapper) Name() string { return p.name }

func serviceProviderWithName(p runtimecontract.ServiceProvider, name string) runtimecontract.ServiceProvider {
	return &namedProviderWrapper{ServiceProvider: p, name: name}
}
