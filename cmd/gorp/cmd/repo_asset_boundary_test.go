package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestTemplateDirectoriesStayAsGeneratorInputs(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	templateRoots := []string{
		"cmd/gorp/cmd/templates/project",
		"cmd/gorp/cmd/templates/golayout/project",
		"cmd/gorp/cmd/templates/multi-flat-wire/project",
		"cmd/gorp/cmd/templates/multi-independent/project",
		"cmd/gorp/cmd/templates/release",
	}

	for _, rel := range templateRoots {
		info, err := os.Stat(filepath.Join("E:/project/gin_plantfrom", filepath.FromSlash(rel)))
		require.NoError(t, err, rel)
		require.True(t, info.IsDir(), rel)
	}
}

func TestDeployDirectoryStaysPrimaryDeliveryAssetRoot(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	mustExist := []string{
		"deploy/kubernetes/base/deployment.yaml",
		"deploy/kubernetes/base/service.yaml",
		"deploy/kubernetes/base/kustomization.yaml",
		"deploy/kubernetes/overlays/dev/kustomization.yaml",
		"deploy/kubernetes/overlays/staging/kustomization.yaml",
		"deploy/kubernetes/overlays/prod/kustomization.yaml",
	}

	for _, rel := range mustExist {
		_, err := os.Stat(filepath.Join("E:/project/gin_plantfrom", filepath.FromSlash(rel)))
		require.NoError(t, err, rel)
	}

	_, err := os.Stat(filepath.Join("E:/project/gin_plantfrom", "deployments"))
	require.Error(t, err)
}
