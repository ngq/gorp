package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestDocGenWritesManualPagesNotVitepressDist(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectRoot := filepath.Join(root, "docgen-boundary")
	manualDir := filepath.Join(projectRoot, "docs", "manual")
	distDir := filepath.Join(projectRoot, "docs", ".vitepress", "dist")
	require.NoError(t, os.MkdirAll(manualDir, 0o755))
	require.NoError(t, os.MkdirAll(distDir, 0o755))

	sentinel := filepath.Join(distDir, "sentinel.txt")
	require.NoError(t, os.WriteFile(sentinel, []byte("keep"), 0o644))

	oldRoot := docProjectRoot
	oldOut := docOut
	oldCheck := docCheck
	oldStdout := docStdout
	defer func() {
		docProjectRoot = oldRoot
		docOut = oldOut
		docCheck = oldCheck
		docStdout = oldStdout
	}()

	docProjectRoot = projectRoot
	docOut = manualDir
	docCheck = false
	docStdout = false

	files, err := generateManualFiles(projectRoot, "", true)
	require.NoError(t, err)
	require.NoError(t, writeManualFiles(manualDir, files))

	_, err = os.Stat(filepath.Join(manualDir, "index.md"))
	require.NoError(t, err)
	content, err := os.ReadFile(sentinel)
	require.NoError(t, err)
	require.Equal(t, "keep", string(content))
}
