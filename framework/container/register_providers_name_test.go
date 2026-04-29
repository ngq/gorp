package container

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestNewContainerBindsContainerKey(t *testing.T) {
	c := New()
	v, err := c.Make(contract.ContainerKey)
	require.NoError(t, err)
	require.Equal(t, c, v)
}

func TestRegisterProviderRejectsEmptyName(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := testProvider{deferLoad: false, provide: []string{"a"}, loaded: &loaded, booted: &booted}

	// 覆盖 Name() 为空的最小变体
	empty := &struct{ testProvider }{p}
	err := c.RegisterProvider(serviceProviderWithName(empty, ""))
	require.EqualError(t, err, "provider name is empty")
}

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
	contract.ServiceProvider
	name string
}

func (p *namedProviderWrapper) Name() string { return p.name }

func serviceProviderWithName(p contract.ServiceProvider, name string) contract.ServiceProvider {
	return &namedProviderWrapper{ServiceProvider: p, name: name}
}
