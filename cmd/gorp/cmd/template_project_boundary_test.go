package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestProjectTemplateStaysSharedBaseOnly(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "project-base")
	data := buildScaffoldData(scaffoldInput{
		Name:            "project-base",
		Module:          "example.com/project-base",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, "templates/project", projectDir, data))
	assertGeneratedProjectHasNoTemplateArtifacts(t, projectDir)

	mustExist := []string{
		".gorp-template.yml",
		".gitignore",
		"README.md",
		"docs/structure.md",
		"docs/deploy.md",
		"docs/next-steps.md",
		"deploy/kubernetes/base/kustomization.yaml",
	}
	for _, rel := range mustExist {
		_, err := os.Stat(filepath.Join(projectDir, filepath.FromSlash(rel)))
		require.NoError(t, err, rel)
	}

	mustNotExist := []string{
		"cmd/app/main.go",
		"app/http/routes.go",
		"app/service/ping.go",
		"config/app.yaml",
		"go.mod",
		"Dockerfile",
		"Makefile",
		".github/workflows/ci.yml",
		"storage/log",
		"storage/runtime",
	}
	for _, rel := range mustNotExist {
		_, err := os.Stat(filepath.Join(projectDir, filepath.FromSlash(rel)))
		require.True(t, os.IsNotExist(err), rel)
	}
}

func TestProjectTemplateDocsDescribeSharedBaseBoundaries(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "project-docs")
	data := buildScaffoldData(scaffoldInput{
		Name:            "project-docs",
		Module:          "example.com/project-docs",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, "templates/project", projectDir, data))

	readme := mustReadGeneratedFile(t, projectDir, "README.md")
	require.Contains(t, readme, "共享模板基底")
	require.Contains(t, readme, "不在这里放 `cmd/app`")
	require.NotContains(t, readme, "Quick start")
	require.NotContains(t, readme, "go run ./cmd/app")

	structure := mustReadGeneratedFile(t, projectDir, "docs/structure.md")
	require.Contains(t, structure, "当前不保留")
	require.Contains(t, structure, "`cmd/app`")
	require.Contains(t, structure, "`internal/server`")

	nextSteps := mustReadGeneratedFile(t, projectDir, "docs/next-steps.md")
	require.Contains(t, nextSteps, "先稳定 `project` 共享基底")
	require.Contains(t, nextSteps, "默认入口消费 facade 契约")

	deploy := mustReadGeneratedFile(t, projectDir, "docs/deploy.md")
	require.Contains(t, deploy, "共享交付片段")
	require.Contains(t, deploy, "当前不包含")
	require.Contains(t, deploy, "Dockerfile")
}

func TestProjectTemplateStableInputRootsStayExplicit(t *testing.T) {
	stableRoots := []string{
		"cmd/gorp/cmd/templates/project/.gorp-template.yml.tmpl",
		"cmd/gorp/cmd/templates/project/.gitignore.tmpl",
		"cmd/gorp/cmd/templates/project/README.md.tmpl",
		"cmd/gorp/cmd/templates/project/docs/structure.md.tmpl",
		"cmd/gorp/cmd/templates/project/docs/deploy.md.tmpl",
		"cmd/gorp/cmd/templates/project/docs/next-steps.md.tmpl",
		"cmd/gorp/cmd/templates/project/deploy/kubernetes/base/kustomization.yaml.tmpl",
	}
	for _, rel := range stableRoots {
		_, err := os.Stat(filepath.Join("E:/project/gin_plantfrom", filepath.FromSlash(rel)))
		require.NoError(t, err, rel)
	}
}

func TestProjectTemplateDoesNotRegrowStarterSpecificRoots(t *testing.T) {
	blockedRoots := []string{
		"cmd/gorp/cmd/templates/project/cmd",
		"cmd/gorp/cmd/templates/project/app",
		"cmd/gorp/cmd/templates/project/config",
		"cmd/gorp/cmd/templates/project/storage",
		"cmd/gorp/cmd/templates/project/go.mod.tmpl",
		"cmd/gorp/cmd/templates/project/Dockerfile.tmpl",
		"cmd/gorp/cmd/templates/project/Makefile.tmpl",
		"cmd/gorp/cmd/templates/project/.github",
	}
	for _, rel := range blockedRoots {
		_, err := os.Stat(filepath.Join("E:/project/gin_plantfrom", filepath.FromSlash(rel)))
		require.True(t, os.IsNotExist(err), rel)
	}
}

func mustReadGeneratedFile(t *testing.T, projectDir, rel string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(rel)))
	require.NoError(t, err)
	return string(content)
}
