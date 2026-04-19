package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func assertGeneratedProjectHasNoTemplateArtifacts(t *testing.T, projectDir string) {
	t.Helper()

	var files []string
	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	require.NoError(t, err)

	for _, path := range files {
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		text := string(content)
		require.NotContains(t, text, "<no value>", "unexpected unresolved placeholder in %s", path)
		require.NotContains(t, text, "{{.", "unexpected template expression in %s", path)
	}
}
