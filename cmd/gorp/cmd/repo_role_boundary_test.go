package cmd

import (
	"bytes"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestRootCommandKeepsToolchainBoundary(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	require.NoError(t, rootCmd.Execute())

	out := buf.String()
	require.Contains(t, out, "new")
	require.Contains(t, out, "template")
	require.Contains(t, out, "proto")
	require.Contains(t, out, "model")
	require.NotContains(t, out, " app ")
	require.NotContains(t, out, " grpc ")
	require.NotContains(t, out, " cron ")
	require.NotContains(t, out, " deploy ")
}

func TestPublicStarterMatrixMatchesCurrentBoundary(t *testing.T) {
	allowed := []string{starterTemplateGoLayout, starterTemplateMultiFlatWire, starterTemplateMultiIndependent}
	for _, name := range allowed {
		require.NoError(t, validateStarterTemplate(name))
		require.NoError(t, validateReleaseStarterTemplate(name))
	}

	require.Error(t, validateStarterTemplate(starterTemplateBase))
	require.Error(t, validateReleaseStarterTemplate(starterTemplateBase))
}
