package container

import (
	"errors"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type failingProvider struct {
	registerErr error
	bootErr     error
	loaded      *int
	booted      *int
}

func (p *failingProvider) Name() string       { return "failing" }
func (p *failingProvider) IsDefer() bool      { return false }
func (p *failingProvider) Provides() []string { return nil }
func (p *failingProvider) Register(contract.Container) error {
	*p.loaded++
	return p.registerErr
}
func (p *failingProvider) Boot(contract.Container) error {
	*p.booted++
	return p.bootErr
}

func TestContainer_LoadProviderDoesNotRetryAfterRegisterFailure(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := &failingProvider{
		registerErr: errors.New("register failed"),
		loaded:      &loaded,
		booted:      &booted,
	}

	// 先只登记 provider，不走 RegisterProvider 的整链路
	c.providersByName[p.Name()] = &providerState{p: p}

	err := c.loadProvider(p.Name())
	require.EqualError(t, err, "register failed")
	require.Equal(t, 1, loaded)
	require.Equal(t, 0, booted)

	// 再次触发 loadProvider，不应该重复执行 Register
	err = c.loadProvider(p.Name())
	require.NoError(t, err)
	require.Equal(t, 1, loaded)
	require.Equal(t, 0, booted)
}

func TestContainer_BootProviderDoesNotRetryAfterBootFailure(t *testing.T) {
	c := New()
	loaded, booted := 0, 0
	p := &failingProvider{
		bootErr: errors.New("boot failed"),
		loaded:  &loaded,
		booted:  &booted,
	}

	c.providersByName[p.Name()] = &providerState{p: p, loaded: true}

	err := c.bootProvider(p.Name())
	require.EqualError(t, err, "boot failed")
	require.Equal(t, 0, loaded)
	require.Equal(t, 1, booted)

	// 再次触发 bootProvider，不应该重复执行 Boot
	err = c.bootProvider(p.Name())
	require.NoError(t, err)
	require.Equal(t, 0, loaded)
	require.Equal(t, 1, booted)
}
