package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestGoLayoutTemplateUsesGorpBootHTTPServiceEntrypoint(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "golayout-entry")
	data := buildScaffoldData(scaffoldInput{
		Name:            "golayout-entry",
		Module:          "example.com/golayout-entry",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateGoLayout), projectDir, data))

	mainFile := filepath.Join(projectDir, "cmd", "app", "main.go")
	content, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "gorp.BootHTTPService(")
	require.Contains(t, text, "gorp.AutoMigrateModels(rt, &data.DemoPO{})")
	require.NotContains(t, text, "cmd.Execute()")
	require.Contains(t, text, `apphttp "example.com/golayout-entry/app/http"`)
}

func TestGoLayoutTemplateRoutesDoNotDuplicateHealthz(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "golayout-routes")
	data := buildScaffoldData(scaffoldInput{
		Name:            "golayout-routes",
		Module:          "example.com/golayout-routes",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateGoLayout), projectDir, data))

	routesFile := filepath.Join(projectDir, "app", "http", "routes.go")
	content, err := os.ReadFile(routesFile)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, `api := r.Group("/api/v1")`)
	require.NotContains(t, text, `engine.GET("/healthz"`)
}

func TestMultiFlatWireTemplateUsesBootstrapEntrypointAndKeepsWireInCmdLayer(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "multi-flat-wire-entry")
	data := buildScaffoldData(scaffoldInput{
		Name:            "multi-flat-wire-entry",
		Module:          "example.com/multi-flat-wire-entry",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiFlatWire), projectDir, data))

	mainFile := filepath.Join(projectDir, "services", "user", "cmd", "main.go")
	content, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	text := string(content)
	require.Contains(t, text, "frameworkbootstrap.BootHTTPService(")
	require.Contains(t, text, "frameworkbootstrap.AutoMigrateModels(rt, &userdata.UserPO{})")

	wireFile := filepath.Join(projectDir, "services", "user", "cmd", "wire.go")
	wireContent, err := os.ReadFile(wireFile)
	require.NoError(t, err)
	wireText := string(wireContent)
	require.Contains(t, wireText, "package main")
	require.Contains(t, wireText, "func wireUserServices(db *gorm.DB)")
}

func TestReleaseGoLayoutTemplateUsesGorpBootHTTPServiceEntrypoint(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "release-golayout-entry")
	data := buildScaffoldData(scaffoldInput{
		Name:            "release-golayout-entry",
		Module:          "example.com/release-golayout-entry",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	srcFS, srcRoot := releaseTemplateSource(starterTemplateGoLayout)
	require.NoError(t, renderTemplateProject(srcFS, srcRoot, projectDir, data))

	mainFile := filepath.Join(projectDir, "cmd", "app", "main.go")
	content, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "frameworkbootstrap.BootHTTPService(")
	require.Contains(t, text, "frameworkbootstrap.AutoMigrateModels(rt, &data.DemoPO{})")
	require.NotContains(t, text, "cmd.Execute()")
	require.Contains(t, text, `apphttp "example.com/release-golayout-entry/app/http"`)
}

func TestReleaseGoLayoutTemplateRoutesDoNotDuplicateHealthz(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "release-golayout-routes")
	data := buildScaffoldData(scaffoldInput{
		Name:            "release-golayout-routes",
		Module:          "example.com/release-golayout-routes",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	srcFS, srcRoot := releaseTemplateSource(starterTemplateGoLayout)
	require.NoError(t, renderTemplateProject(srcFS, srcRoot, projectDir, data))

	routesFile := filepath.Join(projectDir, "app", "http", "routes.go")
	content, err := os.ReadFile(routesFile)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, `api := r.Group("/api/v1")`)
	require.NotContains(t, text, `engine.GET("/healthz"`)
}
