package cmd

import (
	"bytes"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestTemplateVersionStatesEmbeddedMatrixBoundary(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	buf := new(bytes.Buffer)
	templateVersionCmd.SetOut(buf)
	templateVersionCmd.SetErr(buf)
	require.NoError(t, templateVersionCmd.RunE(templateVersionCmd, nil))

	out := buf.String()
	require.Contains(t, out, "Templates are embedded at build time.")
	require.Contains(t, out, "Upgrade CLI to get latest templates.")
	require.NotContains(t, out, "latest remote release")
	require.NotContains(t, out, "template marketplace")
}

func TestTemplateUpgradeForceStillStaysAdvisoryWithoutWritePath(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	buf := new(bytes.Buffer)
	templateUpgradeCmd.SetOut(buf)
	templateUpgradeCmd.SetErr(buf)
	require.NoError(t, templateUpgradeCmd.Flags().Set("dry-run", "false"))
	require.NoError(t, templateUpgradeCmd.Flags().Set("force", "true"))

	// 即使传入 --force，当前实现也仍受项目识别与治理辅助边界约束，不会直接进入写入路径。
	err := templateUpgradeCmd.RunE(templateUpgradeCmd, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot detect project template type")
}

func TestReleaseTemplateSourceRoleStaysExplicit(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantRoot string
	}{
		{name: "golayout", template: starterTemplateGoLayout, wantRoot: "templates/release/golayout/project"},
		{name: "multi-flat-wire", template: starterTemplateMultiFlatWire, wantRoot: "templates/multi-flat-wire/project"},
		{name: "multi-independent", template: starterTemplateMultiIndependent, wantRoot: "templates/multi-independent/project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotRoot := releaseTemplateSource(tt.template)
			require.Equal(t, tt.wantRoot, gotRoot)
		})
	}
}
