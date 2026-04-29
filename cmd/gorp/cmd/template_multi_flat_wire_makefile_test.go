package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestMultiFlatWireMakefileIncludesDeployTargets(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "mfwmake")
	data := buildScaffoldData(scaffoldInput{
		Name:            "mfwmake",
		Module:          "example.com/mfwmake",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiFlatWire), projectDir, data))

	makefile, err := os.ReadFile(filepath.Join(projectDir, "Makefile"))
	require.NoError(t, err)
	text := string(makefile)

	require.True(t, strings.Contains(text, ".PHONY: generate gen-wire gen-proto build clean run-user run-order run-product test tidy deploy-local deploy-local-down build-images deploy-local-build harbor-tag harbor-push"))
	require.Contains(t, text, "build-images:")
	require.Contains(t, text, "deploy-local-build: build-images deploy-local")
	require.Contains(t, text, "harbor-tag:")
	require.Contains(t, text, "harbor-push: build-images harbor-tag")
	require.Contains(t, text, "HARBOR_REGISTRY ?=")
	require.Contains(t, text, "HARBOR_NAMESPACE ?=")
	require.Contains(t, text, "IMAGE_TAG ?=")
}
