package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestGoLayoutRuntimeProviderStaysCompatibilityExtension(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "golayout-runtime-extension")
	data := buildScaffoldData(scaffoldInput{
		Name:            "golayout-runtime-extension",
		Module:          "example.com/golayout-runtime-extension",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateGoLayout), projectDir, data))

	readme, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	require.NoError(t, err)
	text := string(readme)
	require.NotContains(t, text, "app/provider")
	require.Contains(t, text, "cmd/app/main.go")
	require.Contains(t, text, "typed runtime + direct constructor")

	_, err = os.Stat(filepath.Join(projectDir, "app", "provider", "service", "provider.go"))
	require.ErrorIs(t, err, os.ErrNotExist)

	_, err = os.Stat(filepath.Join(projectDir, "app", "provider", "runtime_provider", "provider.go"))
	require.ErrorIs(t, err, os.ErrNotExist)
}
