package container

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type deferredProvider struct {
	name    string
	keys    []string
	loaded  *int
	booted  *int
	value   string
}

func (p *deferredProvider) Name() string       { return p.name }
func (p *deferredProvider) IsDefer() bool      { return true }
func (p *deferredProvider) Provides() []string { return p.keys }
func (p *deferredProvider) Register(c contract.Container) error {
	*p.loaded++
	for _, key := range p.keys {
		value := p.value
		bindKey := key
		c.Bind(bindKey, func(contract.Container) (any, error) { return value, nil }, true)
	}
	return nil
}
func (p *deferredProvider) Boot(contract.Container) error {
	*p.booted++
	return nil
}

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
