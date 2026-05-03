package cmd

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestCurrentTemplateTreesDoNotUsePlaceholderOrGitkeepFiles(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	roots := []string{
		"cmd/gorp/cmd/templates/project",
		"cmd/gorp/cmd/templates/golayout/project",
		"cmd/gorp/cmd/templates/release/golayout/project",
		"cmd/gorp/cmd/templates/multi-flat-wire/project",
		"cmd/gorp/cmd/templates/multi-independent/project",
		"cmd/gorp/cmd/templates/release/project",
	}

	for _, root := range roots {
		var forbidden []string
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			switch filepath.Base(path) {
			case "placeholder.txt", ".gitkeep":
				forbidden = append(forbidden, path)
			}
			return nil
		})
		require.NoError(t, err, root)
		require.Empty(t, forbidden, root)
	}
}

func TestCurrentTemplateTreesDoNotKeepEmptyDirectories(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	roots := []string{
		"cmd/gorp/cmd/templates/project",
		"cmd/gorp/cmd/templates/golayout/project",
		"cmd/gorp/cmd/templates/release/golayout/project",
		"cmd/gorp/cmd/templates/multi-flat-wire/project",
		"cmd/gorp/cmd/templates/multi-independent/project",
		"cmd/gorp/cmd/templates/release/project",
	}

	for _, root := range roots {
		var emptyDirs []string
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				return nil
			}
			entries, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				emptyDirs = append(emptyDirs, path)
			}
			return nil
		})
		require.NoError(t, err, root)
		require.Empty(t, emptyDirs, root)
	}
}

func TestReleaseProjectHandlerTemplatesUseRootFacadeHelpers(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	files := []string{
		"cmd/gorp/cmd/templates/release/project/app/http/handler/ping.go.tmpl",
		"cmd/gorp/cmd/templates/release/project/app/http/handler/health.go.tmpl",
	}

	for _, path := range files {
		content, err := os.ReadFile(filepath.Clean(path))
		require.NoError(t, err, path)
		text := string(content)
		require.NotContains(t, text, "framework/provider/gin", path)
		require.NotContains(t, text, "ginprovider.", path)
		require.Contains(t, text, `gorp "{{.FrameworkModule}}"`, path)
		require.Contains(t, text, "gorp.Success(", path)
	}
}
