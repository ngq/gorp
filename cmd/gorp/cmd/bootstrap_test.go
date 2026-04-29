package cmd

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type bootstrapProviderStub struct{ name string }

func (p *bootstrapProviderStub) Name() string       { return p.name }
func (p *bootstrapProviderStub) Register(contract.Container) error { return nil }
func (p *bootstrapProviderStub) Boot(contract.Container) error     { return nil }
func (p *bootstrapProviderStub) IsDefer() bool                     { return false }
func (p *bootstrapProviderStub) Provides() []string                { return nil }

func TestRegisterBootstrapProvidersOverridesRuntimeAndExtras(t *testing.T) {
	old := readBootstrapHooks()
	defer RegisterBootstrapProviders(old.runtimeProvider, old.extraProviders...)

	runtimeProvider := &bootstrapProviderStub{name: "runtime-a"}
	extraA := &bootstrapProviderStub{name: "extra-a"}
	extraB := &bootstrapProviderStub{name: "extra-b"}

	RegisterBootstrapProviders(runtimeProvider, extraA, extraB)

	cfg := readBootstrapHooks()
	require.Equal(t, runtimeProvider, cfg.runtimeProvider)
	require.Len(t, cfg.extraProviders, 2)
	require.Equal(t, extraA, cfg.extraProviders[0])
	require.Equal(t, extraB, cfg.extraProviders[1])
}

func TestWithExtraProvidersIgnoresNil(t *testing.T) {
	cfg := bootstrapConfig{}
	extraA := &bootstrapProviderStub{name: "extra-a"}
	WithExtraProviders(nil, extraA, nil)(&cfg)
	require.Len(t, cfg.extraProviders, 1)
	require.Equal(t, extraA, cfg.extraProviders[0])
}

func TestWithRuntimeProviderIgnoresNil(t *testing.T) {
	cfg := bootstrapConfig{}
	WithRuntimeProvider(nil)(&cfg)
	require.Nil(t, cfg.runtimeProvider)

	runtimeProvider := &bootstrapProviderStub{name: "runtime-a"}
	WithRuntimeProvider(runtimeProvider)(&cfg)
	require.Equal(t, runtimeProvider, cfg.runtimeProvider)
}
