package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKernelAboveApplicationPlanDeclaresTemplateAcceptanceState(t *testing.T) {
	content, err := os.ReadFile("E:/project/gin_plantfrom/.private-docs/manual/v1-prelude/gorp-platform-kernel-above-application-实施方案.zh-CN.md")
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "当前模板侧验收状态：")
	require.Contains(t, text, "`golayout`、`multi-flat-wire`、`multi-independent` 已按该顺序完成整体重写与封板")
	require.Contains(t, text, "默认入口已收敛到 application-first 主线")
	require.Contains(t, text, "### 13.2 当前模板验收结论")
	require.Contains(t, text, "保留目录不再以空目录、`placeholder.txt`、`.gitkeep` 形式存在")
	require.NotContains(t, text, "三类公开模板尚未按 `framework-starter-template-三类模板直接重写方案.zh-CN.md` 完成整体重写与封板")
}
