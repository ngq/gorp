package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_Load_MultiFile_EnvDir_AndEnvPlaceholder(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "config"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "config", "testing"), 0o755))

	// base files
	require.NoError(t, os.WriteFile(filepath.Join(root, "config", "app.yaml"), []byte("app:\n  address: ':8080'\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "config", "database.yaml"), []byte("database:\n  dsn: 'env(DB_DSN)'\n"), 0o644))

	// legacy env overlay file
	require.NoError(t, os.WriteFile(filepath.Join(root, "config", "app.testing.yaml"), []byte("redis:\n  addr: '127.0.0.1:6379'\n"), 0o644))

	// env directory overlay
	require.NoError(t, os.WriteFile(filepath.Join(root, "config", "testing", "redis.yaml"), []byte("redis:\n  addr: '127.0.0.1:6380'\n"), 0o644))

	restoreWD, _ := os.Getwd()
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(restoreWD) })

	require.NoError(t, os.Setenv("DB_DSN", "sqlite://memory"))
	t.Cleanup(func() { _ = os.Unsetenv("DB_DSN") })

	s := NewService()
	require.NoError(t, s.Load("testing"))

	require.Equal(t, ":8080", s.GetString("app.address"))
	// env(DB_DSN) placeholder substitution
	require.Equal(t, "sqlite://memory", s.GetString("database.dsn"))
	// env dir overlay should win over app.testing.yaml
	require.Equal(t, "127.0.0.1:6380", s.GetString("redis.addr"))
}
