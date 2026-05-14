package baseregistry_test

import (
	"context"
	"testing"

	"github.com/ngq/gorp/contrib/internal/baseregistry"
	"github.com/ngq/gorp/framework/container"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

type mockRegistry struct {
	closed bool
}

func (r *mockRegistry) Register(_ context.Context, _, _ string, _ map[string]string) error {
	return nil
}
func (r *mockRegistry) Deregister(_ context.Context, _, _ string) error { return nil }
func (r *mockRegistry) Discover(_ context.Context, _ string) ([]transportcontract.ServiceInstance, error) {
	return nil, nil
}
func (r *mockRegistry) Close() error {
	r.closed = true
	return nil
}

func TestBaseRegistryProvider_RegisterAndDestroy(t *testing.T) {
	var reg *mockRegistry

	base := &baseregistry.BaseRegistryProvider{
		NameStr: "registry.mock",
		GetConfig: func(c runtimecontract.Container) (any, error) {
			return "mock-cfg", nil
		},
		NewRegistry: func(cfg any) (transportcontract.ServiceRegistry, error) {
			reg = &mockRegistry{}
			return reg, nil
		},
	}

	c := container.New()
	err := base.Register(c)
	require.NoError(t, err)

	regAny, err := c.Make(transportcontract.RPCRegistryKey)
	require.NoError(t, err)
	require.NotNil(t, regAny)

	err = c.Destroy()
	require.NoError(t, err)
	require.True(t, reg.closed)
}

func TestBaseRegistryProvider_ProvidesCorrectKeys(t *testing.T) {
	base := &baseregistry.BaseRegistryProvider{NameStr: "registry.mock"}
	require.Equal(t, []string{transportcontract.RPCRegistryKey}, base.Provides())
}

func TestBaseRegistryProvider_Metadata(t *testing.T) {
	base := &baseregistry.BaseRegistryProvider{NameStr: "registry.mock"}
	require.Equal(t, "registry.mock", base.Name())
	require.True(t, base.IsDefer())
}
