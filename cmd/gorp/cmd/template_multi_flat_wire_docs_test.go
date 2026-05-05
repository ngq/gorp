package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestMultiFlatWireDocsLinkToDeployAssets(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "mfwdocs")
	data := buildScaffoldData(scaffoldInput{
		Name:            "mfwdocs",
		Module:          "example.com/mfwdocs",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiFlatWire), projectDir, data))

	readme, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	require.NoError(t, err)
	deployDoc, err := os.ReadFile(filepath.Join(projectDir, "docs", "deploy.md"))
	require.NoError(t, err)
	nextSteps, err := os.ReadFile(filepath.Join(projectDir, "docs", "next-steps.md"))
	require.NoError(t, err)
	structure, err := os.ReadFile(filepath.Join(projectDir, "docs", "structure.md"))
	require.NoError(t, err)
	componentTree, err := os.ReadFile(filepath.Join(projectDir, "docs", "component-tree.md"))
	require.NoError(t, err)

	readmeText := string(readme)
	deployText := string(deployDoc)
	nextText := string(nextSteps)
	structureText := string(structure)
	componentTreeText := string(componentTree)

	require.Contains(t, readmeText, "docs/deploy.md")
	require.Contains(t, readmeText, "gorp new multi-wire")
	require.Contains(t, readmeText, "make generate")
	require.Contains(t, readmeText, "make test")
	require.Contains(t, readmeText, "make deploy-local")
	require.Contains(t, readmeText, "`internal/server/http`")
	require.Contains(t, readmeText, "`internal/server/grpc`")
	require.Contains(t, readmeText, "gorp.Run(...)")
	require.Contains(t, readmeText, "mfwdocs-user-service")
	require.Contains(t, readmeText, "mfwdocs-order-service")
	require.Contains(t, readmeText, "mfwdocs-product-service")
	require.Contains(t, readmeText, "`mfwdocs-*`")
	require.Contains(t, readmeText, "GRPCServerRegistrar + pb.RegisterXxxServer(...)")
	require.Contains(t, readmeText, "GRPCConnFactory + pb.NewXxxClient(conn)")
	require.NotContains(t, readmeText, "third_party/")
	require.NotContains(t, readmeText, "framework/bootstrap")
	require.Contains(t, structureText, "`internal/server/http/`")
	require.Contains(t, structureText, "`internal/server/grpc/`")
	require.Contains(t, structureText, "framework capability / application")
	require.NotContains(t, structureText, "shared/infrastructure")

	require.Contains(t, componentTreeText, "`services/<service>/internal/server/http/`")
	require.Contains(t, componentTreeText, "`services/<service>/internal/server/grpc/`")
	require.NotContains(t, componentTreeText, "`internal/server/` |")

	require.Contains(t, nextText, "docs/deploy.md")
	require.Contains(t, nextText, "`internal/server/http/`")
	require.Contains(t, nextText, "`internal/server/grpc/`")
	require.Contains(t, nextText, "`make generate`")
	require.Contains(t, deployText, "make deploy-local")
	require.Contains(t, deployText, "make harbor-push")
	require.Contains(t, deployText, "mfwdocs-user-service")
	require.Contains(t, deployText, "mfwdocs-order-service")
	require.Contains(t, deployText, "mfwdocs-product-service")
	require.Contains(t, deployText, "`mfwdocs-*`")
	require.Contains(t, deployText, "deploy/kubernetes/overlays/staging/kustomization.yaml")
	require.Contains(t, deployText, "deploy/kubernetes/overlays/prod/kustomization.yaml")
}
