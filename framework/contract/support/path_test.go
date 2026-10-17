package support

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveResourcePath(t *testing.T) {
	// 1. Empty path returns empty
	require.Equal(t, "", ResolveResourcePath(""))

	// 2. Existing file in current dir
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.conf")
	require.NoError(t, os.WriteFile(testFile, []byte("data"), 0o644))

	// Direct absolute path
	resolved := ResolveResourcePath(testFile)
	require.Equal(t, testFile, resolved)

	// 3. Candidate search dirs
	resolvedFromCandidates := ResolveResourcePath("test.conf", tmpDir)
	require.Equal(t, testFile, resolvedFromCandidates)

	// 4. APP_BASE_PATH
	t.Setenv("APP_BASE_PATH", tmpDir)
	resolvedFromEnv := ResolveResourcePath("test.conf")
	require.Equal(t, testFile, resolvedFromEnv)

	// 5. Non-existent file returns empty string
	require.Equal(t, "", ResolveResourcePath("non_existent_file_xyz_123456.txt"))
}
