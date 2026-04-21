package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestBaseTemplateUsesBootHTTPServiceEntrypoint(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "base-entry")
	data := buildScaffoldData(scaffoldInput{
		Name:            "base-entry",
		Module:          "example.com/base-entry",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateBase), projectDir, data))

	mainFile := filepath.Join(projectDir, "cmd", "app", "main.go")
	content, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "frameworkbootstrap.BootHTTPService(")
	require.NotContains(t, text, "cmd.Execute()")
	require.Contains(t, text, `apphttp "example.com/base-entry/app/http"`)
}

func TestBaseTemplateRoutesDoNotDuplicateHealthz(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "base-routes")
	data := buildScaffoldData(scaffoldInput{
		Name:            "base-routes",
		Module:          "example.com/base-routes",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateBase), projectDir, data))

	routesFile := filepath.Join(projectDir, "app", "http", "routes.go")
	content, err := os.ReadFile(routesFile)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, `engine.GET("/api/ping", handler.Ping)`)
	require.NotContains(t, text, `engine.GET("/healthz"`)
	require.Equal(t, 1, strings.Count(text, `engine.GET("/api/ping", handler.Ping)`))
}

func TestReleaseBaseTemplateUsesBootHTTPServiceEntrypoint(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "release-base-entry")
	data := buildScaffoldData(scaffoldInput{
		Name:            "release-base-entry",
		Module:          "example.com/release-base-entry",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	srcFS, srcRoot := releaseTemplateSource(starterTemplateBase)
	require.NoError(t, renderTemplateProject(srcFS, srcRoot, projectDir, data))

	mainFile := filepath.Join(projectDir, "cmd", "app", "main.go")
	content, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "frameworkbootstrap.BootHTTPService(")
	require.NotContains(t, text, "cmd.Execute()")
	require.Contains(t, text, `apphttp "example.com/release-base-entry/app/http"`)
}

func TestReleaseBaseTemplateRoutesDoNotDuplicateHealthz(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "release-base-routes")
	data := buildScaffoldData(scaffoldInput{
		Name:            "release-base-routes",
		Module:          "example.com/release-base-routes",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	srcFS, srcRoot := releaseTemplateSource(starterTemplateBase)
	require.NoError(t, renderTemplateProject(srcFS, srcRoot, projectDir, data))

	routesFile := filepath.Join(projectDir, "app", "http", "routes.go")
	content, err := os.ReadFile(routesFile)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, `engine.GET("/api/ping", handler.Ping)`)
	require.NotContains(t, text, `engine.GET("/healthz"`)
	require.Equal(t, 1, strings.Count(text, `engine.GET("/api/ping", handler.Ping)`))
}
