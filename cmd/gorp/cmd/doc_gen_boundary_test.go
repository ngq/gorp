package cmd

import (
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestGeneratedManualIndexDoesNotMentionLegacyRuntimePaths(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	index := renderIndexDoc()
	require.Contains(t, index, "当前 CLI 主线已不再把 `app / grpc / cron / build / dev / deploy` 作为 starter 项目的公开 runtime 路径")
	require.NotContains(t, index, "legacy runtime 命令组（兼容/专项）：`gorp app`、`gorp cron`、`gorp grpc`")
}

func TestGeneratedManualCLIDoesNotDescribeLegacyRuntimeAsPublicPath(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	cliDoc, err := renderCLIDoc(rootCmd)
	require.NoError(t, err)
	require.Contains(t, cliDoc, "当前 CLI 主线已不再把 `app / grpc / cron / build / dev / deploy` 作为 starter 项目的公开 runtime 路径")
	require.NotContains(t, cliDoc, "`gorp app` / `grpc` / `cron`：保留的 runtime/兼容命令组")
	require.NotContains(t, cliDoc, "`gorp build` / `dev` / `deploy`：母仓/工程层的构建、开发联调、部署辅助命令")
}
