// Package config_test provides unit tests for the config service.
//
// 适用场景：
// - 验证 Config Service 的 Watch、Load 和 key 变更通知行为。
package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// TestEnvKeyReplacer verifies that viper environment key replacer works correctly for nested config keys.
//
// TestEnvKeyReplacer 验证 viper 环境变量键替换器对嵌套配置键的正确处理。
func TestEnvKeyReplacer(t *testing.T) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	_ = os.Setenv("REDIS_ADDR", "127.0.0.1:6379")
	require.Equal(t, "127.0.0.1:6379", v.GetString("redis.addr"))
}

func TestReadFileWithEnvSubst(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "gorp_config_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := `
app:
  name: ${APP_NAME:my_default_app}
  port: ${APP_PORT:-8080}
  env_legacy: env(LEGACY_VAR)
  dsn: ${MYSQL_DSN:root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local}
  addr: ${HTTP_ADDR::9000}
`
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	_ = tmpFile.Close()

	// 1. 测试未设置环境变量时，默认值生效，且 legacy env 缺失报错
	_, err = readFileWithEnvSubst(tmpFile.Name())
	require.Error(t, err)
	require.Contains(t, err.Error(), "LEGACY_VAR")

	// 2. 设置 LEGACY_VAR 环境变量后，全部正常解析
	_ = os.Setenv("LEGACY_VAR", "legacy_val")
	defer os.Unsetenv("LEGACY_VAR")

	out, err := readFileWithEnvSubst(tmpFile.Name())
	require.NoError(t, err)
	strOut := string(out)
	require.Contains(t, strOut, "name: my_default_app")
	require.Contains(t, strOut, "port: 8080")
	require.Contains(t, strOut, "env_legacy: legacy_val")
	require.Contains(t, strOut, "dsn: root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local")
	require.Contains(t, strOut, "addr: :9000")

	// 3. 环境变量覆盖默认值
	_ = os.Setenv("APP_NAME", "custom_app")
	defer os.Unsetenv("APP_NAME")
	_ = os.Setenv("APP_PORT", "7777")
	defer os.Unsetenv("APP_PORT")

	out2, err := readFileWithEnvSubst(tmpFile.Name())
	require.NoError(t, err)
	strOut2 := string(out2)
	require.Contains(t, strOut2, "name: custom_app")
	require.Contains(t, strOut2, "port: 7777")
}

func TestServiceLoadAndReloadWithAppConfigFileEnv(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "custom.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte("app:\n  name: ${CUSTOM_SVC_NAME:three-svc}\n  address: :9191\n"), 0o600))

	t.Setenv(EnvConfigFileKey, cfgPath)
	svc := NewService()
	require.NoError(t, svc.Load(EnvProd))
	require.Equal(t, "three-svc", svc.GetString("app.name"))
	require.Equal(t, ":9191", svc.GetString("app.address"))

	// 更新文件后 Reload 应立即生效
	require.NoError(t, os.WriteFile(cfgPath, []byte("app:\n  name: three-svc-v2\n  address: :9292\n"), 0o600))
	require.NoError(t, svc.Reload(context.Background()))
	require.Equal(t, "three-svc-v2", svc.GetString("app.name"))
	require.Equal(t, ":9292", svc.GetString("app.address"))
}

func TestServiceLoadWithAppConfigFileSupportsJSONAndFailsFastOnMissing(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "runtime.json")
	require.NoError(t, os.WriteFile(jsonPath, []byte(`{"app":{"name":"json-svc","address":":8181"}}`), 0o600))

	t.Setenv(EnvConfigFileKey, jsonPath)
	svc := NewService()
	require.NoError(t, svc.Load(EnvDev))
	require.Equal(t, "json-svc", svc.GetString("app.name"))
	require.Equal(t, ":8181", svc.GetString("app.address"))

	t.Setenv(EnvConfigFileKey, filepath.Join(dir, "missing.yaml"))
	require.Error(t, svc.Load(EnvDev))
}

func TestServiceWithExplicitFileOverridesEnvAndSupportsWatch(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, "from-env.yaml")
	explicitFile := filepath.Join(dir, "explicit.yaml")
	require.NoError(t, os.WriteFile(envFile, []byte("app:\n  name: env-app\n"), 0o600))
	require.NoError(t, os.WriteFile(explicitFile, []byte("app:\n  name: explicit-app\nfeature:\n  flag: initial\n"), 0o600))

	t.Setenv(EnvConfigFileKey, envFile)
	svc := NewServiceWithFile(explicitFile)
	require.NoError(t, svc.Load(EnvDev))
	require.Equal(t, "explicit-app", svc.GetString("app.name"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	watcher, err := svc.Watch(ctx, "feature.flag")
	require.NoError(t, err)
	defer func() { _ = watcher.Stop() }()

	var (
		mu      sync.Mutex
		updated []string
	)
	watcher.OnChange("feature.flag", func(val any) {
		mu.Lock()
		defer mu.Unlock()
		if s, ok := val.(string); ok {
			updated = append(updated, s)
		}
	})

	// 确保 mtime 发生可观测变化
	time.Sleep(30 * time.Millisecond)
	require.NoError(t, os.WriteFile(explicitFile, []byte("app:\n  name: explicit-app\nfeature:\n  flag: reloaded\n"), 0o600))
	now := time.Now().Add(2 * time.Second)
	require.NoError(t, os.Chtimes(explicitFile, now, now))

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(updated) > 0 && updated[len(updated)-1] == "reloaded"
	}, 3*time.Second, 50*time.Millisecond)
}

