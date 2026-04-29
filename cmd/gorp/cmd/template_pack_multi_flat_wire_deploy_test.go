package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestMultiFlatWireTemplateIncludesDeployDocsAndProdOverlay(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "mfwdeploy")
	data := buildScaffoldData(scaffoldInput{
		Name:            "mfwdeploy",
		Module:          "example.com/mfwdeploy",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiFlatWire), projectDir, data))

	mustExist := []string{
		"docs/deploy.md",
		"deploy/kubernetes/base/deployment.yaml",
		"deploy/kubernetes/base/service.yaml",
		"deploy/kubernetes/base/configmap.yaml",
		"deploy/kubernetes/base/secret.yaml",
		"deploy/kubernetes/base/kustomization.yaml",
		"deploy/kubernetes/overlays/dev/kustomization.yaml",
		"deploy/kubernetes/overlays/staging/kustomization.yaml",
		"deploy/kubernetes/overlays/prod/kustomization.yaml",
		"deploy/kubernetes/overlays/prod/hpa.yaml",
		"deploy/kubernetes/overlays/prod/ingress.yaml",
	}

	for _, rel := range mustExist {
		path := filepath.Join(projectDir, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %s: %v", rel, err)
		}
	}
}
