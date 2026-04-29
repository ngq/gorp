package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestTemplateUpgradeDryRunRespectsFilesFilter(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "upgradefiles")
	data := buildScaffoldData(scaffoldInput{
		Name:            "upgradefiles",
		Module:          "example.com/upgradefiles",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})
	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateGoLayout), projectDir, data))

	// 删除一个模板内默认存在的文件，让它出现在 only in template 列表里
	require.NoError(t, os.Remove(filepath.Join(projectDir, "app", "http", "routes.go")))

	require.NoError(t, os.Chdir(projectDir))
	defer func() { require.NoError(t, frameworktesting.ChdirRepoRoot()) }()

	buf := new(bytes.Buffer)
	templateUpgradeCmd.SetOut(buf)
	templateUpgradeCmd.SetErr(buf)
	require.NoError(t, templateUpgradeCmd.Flags().Set("dry-run", "true"))
	require.NoError(t, templateUpgradeCmd.Flags().Set("force", "false"))
	require.NoError(t, templateUpgradeCmd.Flags().Set("files", "app/http/routes.go"))
	require.NoError(t, templateUpgradeCmd.RunE(templateUpgradeCmd, nil))

	out := buf.String()
	require.Contains(t, out, "Scoped by --files")
	require.Contains(t, out, "Files only in template (1):")
	require.Contains(t, out, "app/http/routes.go")
}

func TestTemplateUpgradeDryRunReportsNoMatchesForSelectedFiles(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "upgradefiles-nomatch")
	data := buildScaffoldData(scaffoldInput{
		Name:            "upgradefiles-nomatch",
		Module:          "example.com/upgradefiles-nomatch",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})
	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateGoLayout), projectDir, data))

	require.NoError(t, os.Chdir(projectDir))
	defer func() { require.NoError(t, frameworktesting.ChdirRepoRoot()) }()

	buf := new(bytes.Buffer)
	templateUpgradeCmd.SetOut(buf)
	templateUpgradeCmd.SetErr(buf)
	require.NoError(t, templateUpgradeCmd.Flags().Set("dry-run", "true"))
	require.NoError(t, templateUpgradeCmd.Flags().Set("force", "false"))
	require.NoError(t, templateUpgradeCmd.Flags().Set("files", "not/exist.go"))
	require.NoError(t, templateUpgradeCmd.RunE(templateUpgradeCmd, nil))

	out := buf.String()
	require.Contains(t, out, "No matching upgrade candidates found for the selected --files.")
}

