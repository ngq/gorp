package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestTemplateDiffShowsReadOnlyWorkflow(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "diffcheck")
	data := buildScaffoldData(scaffoldInput{
		Name:            "diffcheck",
		Module:          "example.com/diffcheck",
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
	templateDiffCmd.SetOut(buf)
	templateDiffCmd.SetErr(buf)
	require.NoError(t, templateDiffCmd.RunE(templateDiffCmd, nil))

	out := buf.String()
	require.Contains(t, out, "Detected template: golayout")
	require.Contains(t, out, "Template Diff Report")
	require.Contains(t, out, "Next: run 'gorp template upgrade --dry-run' to preview an upgrade workflow.")
}

func TestTemplateUpgradeWithoutDryRunStaysAdvisory(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "upgradecheck")
	data := buildScaffoldData(scaffoldInput{
		Name:            "upgradecheck",
		Module:          "example.com/upgradecheck",
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
	templateUpgradeCmd.Flags().Set("dry-run", "false")
	templateUpgradeCmd.Flags().Set("force", "false")
	require.NoError(t, templateUpgradeCmd.RunE(templateUpgradeCmd, nil))

	out := buf.String()
	require.Contains(t, out, "Template upgrade is a destructive operation.")
	require.Contains(t, out, "Please use --dry-run first to preview changes.")
	require.Contains(t, out, "Run: gorp template upgrade --dry-run")
	require.Contains(t, out, "Run: gorp template upgrade --force")
}

func TestTemplateUpgradeDryRunUsesRealDiffResults(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "upgradedryrun")
	data := buildScaffoldData(scaffoldInput{
		Name:            "upgradedryrun",
		Module:          "example.com/upgradedryrun",
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
	templateUpgradeCmd.Flags().Set("dry-run", "true")
	templateUpgradeCmd.Flags().Set("force", "false")
	require.NoError(t, templateUpgradeCmd.RunE(templateUpgradeCmd, nil))

	out := buf.String()
	require.Contains(t, out, "Dry-run mode: previewing upgrade workflow from real diff results")
	require.Contains(t, out, "Files only in template")
	require.Contains(t, out, "Files with different content")
}
