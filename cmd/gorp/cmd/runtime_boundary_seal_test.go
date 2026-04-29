package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestFrameworkBootstrapDoesNotDependOnCLIRoot(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	content, err := os.ReadFile(filepath.Join("framework", "bootstrap", "http_service.go"))
	require.NoError(t, err)
	text := string(content)

	require.NotContains(t, text, "cmd/gorp/cmd")
	require.NotContains(t, text, "cobra")
	require.NotContains(t, text, "cmd.Execute(")
}

func TestCLIMainlineStaysToolingOnly(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	content, err := os.ReadFile(filepath.Join("cmd", "gorp", "cmd", "root.go"))
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "framework + starter templates + developer toolchain")
	require.Contains(t, text, "legacy runtime 命令（app / grpc / cron / build / dev / deploy）已退役")
	require.NotContains(t, text, "默认公开业务启动入口")
}

func TestCLIBootstrapStaysCompatibilityLayer(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	content, err := os.ReadFile(filepath.Join("cmd", "gorp", "cmd", "bootstrap.go"))
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "共享 CLI / legacy 辅助命令链路")
	require.Contains(t, text, "业务服务默认主线仍落在项目自己的启动入口")
	require.True(t, strings.Contains(text, "NewCLIApplication"))
}
