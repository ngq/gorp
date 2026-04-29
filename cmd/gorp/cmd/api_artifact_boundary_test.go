package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIGenRequiresDocsSwaggerJSON(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "docs"), 0o755))
	require.NoError(t, os.Chdir(root))
	defer func() { require.NoError(t, frameworktesting.ChdirRepoRoot()) }()

	buf := new(bytes.Buffer)
	openapiGenCmd.SetOut(buf)
	openapiGenCmd.SetErr(buf)
	err := openapiGenCmd.RunE(openapiGenCmd, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), filepath.Join("docs", "swagger.json"))
}

func TestSwaggerGenWritesIntoDocsDirectory(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	buf := new(bytes.Buffer)
	swaggerGenCmd.SetOut(buf)
	swaggerGenCmd.SetErr(buf)

	usage := swaggerGenCmd.Use + " :: " + swaggerGenCmd.Short
	require.True(t, strings.Contains(usage, "gen"))
	require.True(t, strings.Contains(usage, "swagger2"))

	content := string(mustReadFile(t, filepath.Join("cmd", "gorp", "cmd", "swagger.go")))
	require.Contains(t, content, `filepath.Join("docs")`)
	require.Contains(t, content, `filepath.Join("cmd", "gorp", "main.go")`)
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return b
}
