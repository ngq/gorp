package cmd

import (
	"os/exec"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

// TestGoLayoutTemplateCompiles verifies that the golayout template generates a project
// that compiles with go build after go mod tidy.
// This addresses verification requirement #6: "模板生成项目在最小配置下即可运行".
//
// TestGoLayoutTemplateCompiles 验证 golayout 模板生成的项目在 go mod tidy 后可通过 go build 编译。
// 对应验收要求 #6："模板生成项目在最小配置下即可运行"。
func TestGoLayoutTemplateCompiles(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "golayout-compile")
	data := buildScaffoldData(scaffoldInput{
		Name:             "golayout-compile",
		Module:           "example.com/golayout-compile",
		FrameworkModule:  "github.com/ngq/gorp",
		FrameworkPath:    ".",
		FrameworkVersion: "v0.0.0",
		Backend:          "gorm",
		WithDB:           true,
		WithSwagger:      true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateGoLayout), projectDir, data))

	// go mod tidy 拉齐依赖
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = projectDir
	tidyOut, err := tidyCmd.CombinedOutput()
	require.NoError(t, err, "go mod tidy failed: %s", string(tidyOut))

	// go build ./cmd/app 验证编译
	buildCmd := exec.Command("go", "build", "./cmd/app")
	buildCmd.Dir = projectDir
	buildOut, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "go build failed: %s", string(buildOut))
}

// TestMultiFlatWireTemplateCompiles verifies that the multi-flat-wire template generates a project
// whose user service compiles.
//
// TestMultiFlatWireTemplateCompiles 验证 multi-flat-wire 模板生成的项目 user 服务可编译。
func TestMultiFlatWireTemplateCompiles(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "multi-flat-wire-compile")
	data := buildScaffoldData(scaffoldInput{
		Name:             "multi-flat-wire-compile",
		Module:           "example.com/multi-flat-wire-compile",
		FrameworkModule:  "github.com/ngq/gorp",
		FrameworkPath:    ".",
		FrameworkVersion: "v0.0.0",
		Backend:          "gorm",
		WithDB:           true,
		WithSwagger:      true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiFlatWire), projectDir, data))

	userDir := filepath.Join(projectDir, "services", "user")

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = userDir
	tidyOut, err := tidyCmd.CombinedOutput()
	require.NoError(t, err, "go mod tidy failed: %s", string(tidyOut))

	buildCmd := exec.Command("go", "build", "./cmd")
	buildCmd.Dir = userDir
	buildOut, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "go build failed: %s", string(buildOut))
}
