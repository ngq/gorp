package cmd

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestBootstrapCompatibilityHooksDoNotAffectDefaultsWhenUnused(t *testing.T) {
	old := readBootstrapHooks()
	defer RegisterBootstrapProviders(old.runtimeProvider, old.extraProviders...)

	RegisterBootstrapProviders(nil)

	cfg := readBootstrapHooks()
	require.Nil(t, cfg.runtimeProvider)
	require.Len(t, cfg.extraProviders, 0)
}

func TestBootstrapCompatibilityHooksOnlyStoreExplicitProjectOverrides(t *testing.T) {
	old := readBootstrapHooks()
	defer RegisterBootstrapProviders(old.runtimeProvider, old.extraProviders...)

	runtimeProvider := &bootstrapProviderStub{name: "runtime-compat"}
	extra := &bootstrapProviderStub{name: "extra-compat"}
	RegisterBootstrapProviders(runtimeProvider, extra)

	cfg := readBootstrapHooks()
	require.Equal(t, runtimeProvider, cfg.runtimeProvider)
	require.Equal(t, []contract.ServiceProvider{extra}, cfg.extraProviders)
}
