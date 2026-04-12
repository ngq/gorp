package container

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ngq/gorp/framework/contract"
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
func (p *testProvider) Register(c contract.Container) error {
	*p.loaded++
	for _, k := range p.provide {
		key := k
		c.Bind(key, func(contract.Container) (any, error) { return "ok", nil }, true)
	}
	return nil
}
func (p *testProvider) Boot(contract.Container) error {
	*p.booted++
	return nil
}

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
