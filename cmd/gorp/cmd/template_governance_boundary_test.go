package cmd

import (
	"bytes"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestTemplateVersionCommandListsPublicStarterTemplatesOnly(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	buf := new(bytes.Buffer)
	templateVersionCmd.SetOut(buf)
	templateVersionCmd.SetErr(buf)

	require.NoError(t, templateVersionCmd.RunE(templateVersionCmd, nil))

	out := buf.String()
	require.NotContains(t, out, "base: minimal skeleton for custom structure")
	require.Contains(t, out, "golayout: default single-service starter")
	require.Contains(t, out, "multi-flat-wire: default microservice starter with Wire assembly")
	require.Contains(t, out, "multi-independent: stronger independently-governed multi-service starter")

	require.NotContains(t, out, "golayout-wire: advanced single-service starter with Wire assembly")
	require.NotContains(t, out, "multi-flat: default multi-service starter")
	require.Contains(t, out, "Release-pack / from-release currently supports the same public starter set: golayout, multi-flat-wire, multi-independent.")
}

func TestReleaseTemplateValidationAllowsMultiIndependent(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	err := validateReleaseStarterTemplate("multi-independent")
	require.NoError(t, err)
}
