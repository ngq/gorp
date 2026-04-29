package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestSmokeScriptExistsAndCoversCurrentMainline(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	path := filepath.Join("scripts", "smoke.sh")
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	text := string(content)

	require.Contains(t, text, "go test ./framework/container/...")
	require.Contains(t, text, "go test ./framework/contract/...")
	require.Contains(t, text, "go test ./framework/provider/...")
	require.Contains(t, text, "go test ./framework/bootstrap/...")
	require.Contains(t, text, "go test ./contrib/...")
	require.Contains(t, text, "go test ./cmd/gorp/cmd")
	require.Contains(t, text, "kubectl kustomize deploy/kubernetes/overlays/dev")
	require.Contains(t, text, "kubectl kustomize deploy/kubernetes/overlays/staging")
	require.Contains(t, text, "kubectl kustomize deploy/kubernetes/overlays/prod")
	require.True(t, strings.HasPrefix(text, "#!/usr/bin/env bash"))
}
