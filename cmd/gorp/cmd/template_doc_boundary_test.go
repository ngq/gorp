package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplateRewritePlanDeclaresKeepDirectoryReadmeRule(t *testing.T) {
	content, err := os.ReadFile("E:/project/gin_plantfrom/.private-docs/manual/v1-prelude/framework-starter-template-三类模板直接重写方案.zh-CN.md")
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "#### 保留目录规则")
	require.Contains(t, text, "方案里保留的目录，不允许以空目录形态进入公开 starter")
	require.Contains(t, text, "如果当前目录还没有实际代码模板，至少放一份 `README.md`")
	require.Contains(t, text, "不再使用 `placeholder.txt`、`.gitkeep`")
}
