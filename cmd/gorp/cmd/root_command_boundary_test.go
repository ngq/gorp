package cmd

import (
	"bytes"
	"strings"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestRootCommandFocusesOnToolchainAndScaffolding(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	children := rootCmd.Commands()
	names := make([]string, 0, len(children))
	for _, c := range children {
		if c.Hidden {
			continue
		}
		names = append(names, c.Name())
	}

	require.Contains(t, names, "new")
	require.Contains(t, names, "template")
	require.Contains(t, names, "proto")
	require.Contains(t, names, "model")
	require.Contains(t, names, "provider")
	require.Contains(t, names, "middleware")
	require.Contains(t, names, "command")
	require.Contains(t, names, "doc")
	require.Contains(t, names, "swagger")
	require.Contains(t, names, "openapi")
}

func TestRootCommandDoesNotExposeLegacyRuntimeCommands(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	children := rootCmd.Commands()
	names := make([]string, 0, len(children))
	for _, c := range children {
		if c.Hidden {
			continue
		}
		names = append(names, c.Name())
	}

	require.NotContains(t, names, "app")
	require.NotContains(t, names, "grpc")
	require.NotContains(t, names, "cron")
	require.NotContains(t, names, "build")
	require.NotContains(t, names, "dev")
	require.NotContains(t, names, "deploy")
}

func TestRootHelpDescribesNonRuntimeCLIModel(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, "Framework, starter templates, and developer tooling for gorp")
	require.NotContains(t, strings.ToLower(out), "gorp app")
	require.NotContains(t, strings.ToLower(out), "gorp grpc")
	require.NotContains(t, strings.ToLower(out), "gorp cron")
}
