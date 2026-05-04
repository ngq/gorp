package ent

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "orm.ent", p.Name())
	require.False(t, p.IsDefer())
	require.Equal(t, []string{datacontract.EntClientKey}, p.Provides())
}

type entConfigStub struct{}

func (entConfigStub) Env() string        { return "dev" }
func (entConfigStub) Get(key string) any { return nil }
func (entConfigStub) GetString(key string) string {
	if key == "database.driver" {
		return "sqlite"
	}
	return ""
}
func (entConfigStub) GetInt(string) int           { return 0 }
func (entConfigStub) GetBool(string) bool         { return false }
func (entConfigStub) GetFloat(string) float64     { return 0 }
func (entConfigStub) Unmarshal(string, any) error { return nil }
func (entConfigStub) Watch(context.Context, string) (datacontract.ConfigWatcher, error) {
	return nil, nil
}
func (entConfigStub) Reload(context.Context) error { return nil }

func TestProviderRequiresProjectLevelFactory(t *testing.T) {
	c := container.New()
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) { return entConfigStub{}, nil }, true)

	p := NewProvider()
	require.NoError(t, p.Register(c))

	_, err := c.Make(datacontract.EntClientKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), datacontract.EntClientFactoryKey)
	require.Contains(t, err.Error(), "database.backend=ent is selected")
}
