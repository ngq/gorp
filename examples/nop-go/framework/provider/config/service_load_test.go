// Package config_test provides unit tests for config service file loading and path resolution.
//
// 适用场景：
// - 验证 LoadLocalConfigToViper 对 app base path 的尊重行为。
// - 确保配置目录解析和文件加载路径正确。
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// TestLoadLocalConfigToViper_RespectsAppBasePath verifies that LoadLocalConfigToViper respects the APP_BASE_PATH environment variable.
//
// TestLoadLocalConfigToViper_RespectsAppBasePath 验证 LoadLocalConfigToViper 正确遵循 APP_BASE_PATH 环境变量。
func TestLoadLocalConfigToViper_RespectsAppBasePath(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "app.yaml"), []byte("service:\n  name: app-base\n"), 0o644))

	t.Setenv("APP_BASE_PATH", root)

	v := viper.New()
	require.NoError(t, LoadLocalConfigToViper(v, "", ""))
	require.Equal(t, "app-base", v.GetString("service.name"))
}

// TestService_Load_MultiFile_EnvDir_AndEnvPlaceholder verifies multi-file loading, environment directory overlays, and environment variable placeholder substitution.
//
// TestService_Load_MultiFile_EnvDir_AndEnvPlaceholder 验证多文件加载、环境目录覆盖和环境变量占位符替换。
func TestService_Load_MultiFile_EnvDir_AndEnvPlaceholder(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "config"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "config", "testing"), 0o755))

	// base files
	require.NoError(t, os.WriteFile(filepath.Join(root, "config", "app.yaml"), []byte("app:\n  address: ':8080'\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "config", "database.yaml"), []byte("database:\n  dsn: 'env(DB_DSN)'\n"), 0o644))

	// env overlay file
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
