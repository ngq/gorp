package baseconfigsource_test

import (
	"context"
	"testing"

	"github.com/ngq/gorp/contrib/internal/baseconfigsource"
	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

type mockConfigSource struct {
	closed bool
}

func (s *mockConfigSource) Load(_ context.Context) (map[string]any, error) { return nil, nil }
func (s *mockConfigSource) Get(_ context.Context, _ string) (any, error)   { return nil, nil }
func (s *mockConfigSource) Set(_ context.Context, _ string, _ any) error   { return nil }
func (s *mockConfigSource) Watch(_ context.Context, _ string) (datacontract.ConfigWatcher, error) {
	return nil, nil
}
func (s *mockConfigSource) Close() error {
	s.closed = true
	return nil
}

func TestBaseConfigSourceProvider_RegisterAndDestroy(t *testing.T) {
	var src *mockConfigSource

	base := &baseconfigsource.BaseConfigSourceProvider{
		NameStr: "configsource.mock",
		GetConfig: func(c runtimecontract.Container) (any, error) {
			return "mock-cfg", nil
		},
		NewSource: func(cfg any) (datacontract.ConfigSource, error) {
			src = &mockConfigSource{}
			return src, nil
		},
	}

	c := container.New()
	err := base.Register(c)
	require.NoError(t, err)

	srcAny, err := c.Make(datacontract.ConfigSourceKey)
	require.NoError(t, err)
	require.NotNil(t, srcAny)

	err = c.Destroy()
	require.NoError(t, err)
	require.True(t, src.closed)
}

func TestBaseConfigSourceProvider_ProvidesCorrectKeys(t *testing.T) {
	base := &baseconfigsource.BaseConfigSourceProvider{NameStr: "configsource.mock"}
	require.Equal(t, []string{datacontract.ConfigSourceKey}, base.Provides())
}

func TestBaseConfigSourceProvider_Metadata(t *testing.T) {
	base := &baseconfigsource.BaseConfigSourceProvider{NameStr: "configsource.mock"}
	require.Equal(t, "configsource.mock", base.Name())
	require.True(t, base.IsDefer())
}
