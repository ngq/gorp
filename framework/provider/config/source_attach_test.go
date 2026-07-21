package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	"github.com/stretchr/testify/require"
)

type attachedSource struct {
	values map[string]any
}

func (s *attachedSource) Load(context.Context) (map[string]any, error) { return s.values, nil }
func (s *attachedSource) Get(context.Context, string) (any, error)     { return nil, nil }
func (s *attachedSource) Set(context.Context, string, any) error       { return errors.New("unsupported") }
func (s *attachedSource) Watch(context.Context, string) (datacontract.ConfigWatcher, error) {
	return nil, errors.New("unsupported")
}
func (s *attachedSource) Close() error { return nil }

func TestAttachConfigSourceReloadsRemoteOverrides(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "config"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "config", "app.yaml"), []byte("feature:\n  enabled: false\n"), 0o600))
	t.Setenv("APP_BASE_PATH", root)

	svc := NewService()
	require.NoError(t, svc.Load(EnvTest))
	require.False(t, svc.GetBool("feature.enabled"))
	require.NoError(t, svc.AttachConfigSource(&attachedSource{values: map[string]any{"feature.enabled": true}}))
	require.NoError(t, svc.Reload(context.Background()))
	require.True(t, svc.GetBool("feature.enabled"))
}
