package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestBaseTemplateStaysMinimalStarterInput(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "base-minimal")
	data := buildScaffoldData(scaffoldInput{
		Name:            "base-minimal",
		Module:          "example.com/base-minimal",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, "templates/project", projectDir, data))
	mainFile := filepath.Join(projectDir, "cmd", "app", "main.go")
	mainContent, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	mainText := string(mainContent)
	require.Contains(t, mainText, "共享 starter 基底的默认入口")
	require.Contains(t, mainText, "对外推荐入口应以 golayout / multi-flat-wire / multi-independent 为准")
	readme, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	require.NoError(t, err)
	readmeText := string(readme)
	require.Contains(t, readmeText, "共享骨架基底")
	require.NotContains(t, readmeText, "更适合本地联调的 starter")
	require.Contains(t, readmeText, "如果你明确需要一个更接近“固定版本交付物”的结果")

	deployDoc, err := os.ReadFile(filepath.Join(projectDir, "docs", "deploy.md"))
	require.NoError(t, err)
	deployText := string(deployDoc)
	require.Contains(t, deployText, "共享骨架基底只保留最小交付骨架")
	require.NotContains(t, deployText, "base starter 提供最小工程化交付骨架")

	structure, err := os.ReadFile(filepath.Join(projectDir, "docs", "structure.md"))
	require.NoError(t, err)
	structureText := string(structure)
	require.Contains(t, structureText, "共享 starter 基底目录说明")
	require.NotContains(t, structureText, "Starter 项目目录说明")

	nextSteps, err := os.ReadFile(filepath.Join(projectDir, "docs", "next-steps.md"))
	require.NoError(t, err)
	nextText := string(nextSteps)
	require.Contains(t, nextText, "数据库相关能力按需再看工具链命令")
	require.Contains(t, nextText, "固定版本交付相关动作放到真正需要发布时再做")
	require.NotContains(t, nextText, "gorp model test")

	assertGeneratedProjectHasNoTemplateArtifacts(t, projectDir)

	mustExist := []string{
		".gorp-template.yml",
		"cmd/app/main.go",
		"app/http/routes.go",
		"app/http/handler/ping.go",
		"app/service/ping.go",
		"config/app.yaml",
		"docs/structure.md",
		"deploy/kubernetes/base/kustomization.yaml",
	}
	for _, rel := range mustExist {
		_, err := os.Stat(filepath.Join(projectDir, filepath.FromSlash(rel)))
		require.NoError(t, err, rel)
	}
}

func TestBaseTemplateStableInputRootsStayExplicit(t *testing.T) {
	stableRoots := []string{
		"cmd/gorp/cmd/templates/project/cmd/app/main.go.tmpl",
		"cmd/gorp/cmd/templates/project/app/http/routes.go.tmpl",
		"cmd/gorp/cmd/templates/project/app/service/ping.go.tmpl",
		"cmd/gorp/cmd/templates/project/config/app.yaml.tmpl",
		"cmd/gorp/cmd/templates/project/docs/structure.md.tmpl",
		"cmd/gorp/cmd/templates/project/deploy/kubernetes/base/kustomization.yaml.tmpl",
	}
	for _, rel := range stableRoots {
		_, err := os.Stat(filepath.Join("E:/project/gin_plantfrom", filepath.FromSlash(rel)))
		require.NoError(t, err, rel)
	}
}

func TestBaseTemplateDriftSensitiveAreasStayVisible(t *testing.T) {
	highRisk := []string{
		"cmd/gorp/cmd/templates/project/cmd/app/main.go.tmpl",
		"cmd/gorp/cmd/templates/project/app/http/routes.go.tmpl",
		"cmd/gorp/cmd/templates/project/README.md.tmpl",
		"cmd/gorp/cmd/templates/project/docs/next-steps.md.tmpl",
		"cmd/gorp/cmd/templates/project/docs/deploy.md.tmpl",
		"cmd/gorp/cmd/templates/project/deploy/kubernetes/base/deployment.yaml.tmpl",
	}
	for _, rel := range highRisk {
		_, err := os.Stat(filepath.Join("E:/project/gin_plantfrom", filepath.FromSlash(rel)))
		require.NoError(t, err, rel)
	}
}
