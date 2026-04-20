package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestRenderMultiFlatWireTemplateIncludesTests(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "mfwverify")
	data := buildScaffoldData(scaffoldInput{
		Name:            "mfwverify",
		Module:          "example.com/mfwverify",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	err := renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiFlatWire), projectDir, data)
	require.NoError(t, err)

	mustExist := []string{
		"services/user/internal/biz/user_test.go",
		"services/order/internal/biz/order_test.go",
		"services/product/internal/biz/product_test.go",
		"services/user/internal/server/http_test.go",
		"services/order/internal/server/http_test.go",
		"services/product/internal/server/http_test.go",
		"services/user/internal/service/grpc_smoke_test.go",
		"services/order/internal/service/grpc_smoke_test.go",
		"deployments/compose/docker-compose.yaml",
		"deployments/docker/Dockerfile.user",
		"deployments/docker/Dockerfile.order",
		"deployments/docker/Dockerfile.product",
		"docs/structure.md",
		"docs/component-tree.md",
		"docs/next-steps.md",
		"Makefile",
	}

	for _, rel := range mustExist {
		path := filepath.Join(projectDir, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %s: %v", rel, err)
		}
	}

	assertGeneratedProjectHasNoTemplateArtifacts(t, projectDir)
}
