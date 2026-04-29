package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestRenderMultiIndependentTemplateIncludesStarterMarkers(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "miverify")
	data := buildScaffoldData(scaffoldInput{
		Name:            "miverify",
		Module:          "example.com/miverify",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiIndependent), projectDir, data))

	mustExist := []string{
		".gorp-template.yml",
		"README.md",
		"Makefile",
		"go.work",
		"shared/go.mod",
		"services/user/go.mod",
		"services/order/go.mod",
		"services/product/go.mod",
		"services/user/cmd/main.go",
		"services/order/cmd/main.go",
		"services/product/cmd/main.go",
	}

	for _, rel := range mustExist {
		path := filepath.Join(projectDir, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %s: %v", rel, err)
		}
	}

	readme, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	require.NoError(t, err)
	readmeText := string(readme)
	require.Contains(t, readmeText, "## 特点")
	require.Contains(t, readmeText, "默认入口直接复用 framework/bootstrap")
	require.NotContains(t, readmeText, "保留扩展空间")
	require.Contains(t, readmeText, "先把服务跑通，再按业务需要补日志字段。")

	mainFile := filepath.Join(projectDir, "services", "user", "cmd", "main.go")
	mainContent, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	mainText := string(mainContent)
	require.Contains(t, mainText, "frameworkbootstrap.BootHTTPService(")
	require.Contains(t, mainText, "frameworkbootstrap.AutoMigrateModels(rt, &userdata.UserPO{})")
	require.NotContains(t, mainText, "默认直接通过 framework/bootstrap")

	productStartFile := filepath.Join(projectDir, "services", "product", "start.go")
	productStartContent, err := os.ReadFile(productStartFile)
	require.NoError(t, err)
	productStartText := string(productStartContent)
	require.Contains(t, productStartText, "暴露最小启动入口")
	require.NotContains(t, productStartText, "对外暴露的最小启动入口")

	assertGeneratedProjectHasNoTemplateArtifacts(t, projectDir)
}
