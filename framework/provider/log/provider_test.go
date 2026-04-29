package log

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type stubRoot struct {
	log string
}

func (s stubRoot) BasePath() string    { return "." }
func (s stubRoot) StoragePath() string { return "." }
func (s stubRoot) RuntimePath() string { return "." }
func (s stubRoot) LogPath() string     { return s.log }
func (s stubRoot) ConfigPath() string  { return "." }
func (s stubRoot) TempPath() string    { return "." }

type stubConfig struct {
	values map[string]any
}

func (s *stubConfig) Env() string                 { return "testing" }
func (s *stubConfig) Get(key string) any          { return s.values[key] }
func (s *stubConfig) GetString(key string) string { v, _ := s.values[key].(string); return v }
func (s *stubConfig) GetInt(key string) int       { v, _ := s.values[key].(int); return v }
func (s *stubConfig) GetBool(key string) bool     { v, _ := s.values[key].(bool); return v }
func (s *stubConfig) GetFloat(string) float64     { return 0 }
func (s *stubConfig) Unmarshal(string, any) error { return nil }
func (s *stubConfig) Watch(_ context.Context, _ string) (contract.ConfigWatcher, error) {
	return nil, nil
}
func (s *stubConfig) Reload(_ context.Context) error { return nil }

func TestProviderMeta(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "log", p.Name())
	require.False(t, p.IsDefer())
	require.Equal(t, []string{contract.LogKey}, p.Provides())
}

func TestProviderUsesRootLogPathWhenRootExists(t *testing.T) {
	c := container.New()
	c.Bind(contract.ConfigKey, func(contract.Container) (any, error) {
		return &stubConfig{values: map[string]any{
			"log.driver": "single",
		}}, nil
	}, true)
	c.Bind(contract.RootKey, func(contract.Container) (any, error) {
		return stubRoot{log: "var/logs"}, nil
	}, true)

	require.NoError(t, c.RegisterProvider(NewProvider()))
	v, err := c.Make(contract.LogKey)
	require.NoError(t, err)
	_, ok := v.(contract.Logger)
	require.True(t, ok)
}

func TestProviderFallsBackToRelativeLogFileWithoutRoot(t *testing.T) {
	c := container.New()
	c.Bind(contract.ConfigKey, func(contract.Container) (any, error) {
		return &stubConfig{values: map[string]any{
			"log.driver": "single",
		}}, nil
	}, true)

	require.NoError(t, c.RegisterProvider(NewProvider()))
	v, err := c.Make(contract.LogKey)
	require.NoError(t, err)
	_, ok := v.(contract.Logger)
	require.True(t, ok)
}

func TestProviderBindLoggerWithBoolOverrides(t *testing.T) {
	c := container.New()
	c.Bind(contract.ConfigKey, func(contract.Container) (any, error) {
		return &stubConfig{values: map[string]any{
			"log.driver":     "stdout",
			"log.local_time": false,
			"log.compress":   true,
		}}, nil
	}, true)

	require.NoError(t, c.RegisterProvider(NewProvider()))
	v, err := c.Make(contract.LogKey)
	require.NoError(t, err)
	_, ok := v.(contract.Logger)
	require.True(t, ok)
}
