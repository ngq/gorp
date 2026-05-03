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
		".gorp-template.yml",
		"go.mod",
		"Makefile",
		"pkg/README.md",
		"services/user/internal/biz/user_test.go",
		"services/order/internal/biz/order_test.go",
		"services/product/internal/biz/product_test.go",
		"services/user/internal/server/http/routes_test.go",
		"services/order/internal/server/http/routes_test.go",
		"services/product/internal/server/http/routes_test.go",
		"services/user/internal/server/http/routes.go",
		"services/order/internal/server/http/routes.go",
		"services/product/internal/server/http/routes.go",
		"services/user/internal/server/http/handler/user.go",
		"services/order/internal/server/http/handler/order.go",
		"services/product/internal/server/http/handler/product.go",
		"services/user/internal/server/grpc/register.go",
		"services/order/internal/server/grpc/README.md",
		"services/product/internal/server/grpc/README.md",
		"services/user/internal/server/http/middleware/README.md",
		"services/order/internal/server/http/middleware/README.md",
		"services/product/internal/server/http/middleware/README.md",
		"services/user/internal/service/grpc_smoke_test.go",
		"services/order/internal/service/grpc_smoke_test.go",
		"deploy/compose/docker-compose.yaml",
		"deploy/docker/Dockerfile.user",
		"deploy/docker/Dockerfile.order",
		"deploy/docker/Dockerfile.product",
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

	mustNotExist := []string{
		"cmd",
		"scripts",
		"deploy/k8s",
		"services/gateway",
		"shared/infrastructure",
		"third_party",
	}
	for _, rel := range mustNotExist {
		_, err := os.Stat(filepath.Join(projectDir, filepath.FromSlash(rel)))
		require.True(t, os.IsNotExist(err), rel)
	}

	readme, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	require.NoError(t, err)
	readmeText := string(readme)
	require.Contains(t, readmeText, "gorp new multi-wire")
	require.Contains(t, readmeText, "make generate")
	require.Contains(t, readmeText, "make test")
	require.Contains(t, readmeText, "make deploy-local")
	require.Contains(t, readmeText, "`internal/server/http`")
	require.Contains(t, readmeText, "`internal/server/grpc`")
	require.Contains(t, readmeText, "gorp.Run(...)")
	require.Contains(t, readmeText, "GRPCServerRegistrar + pb.RegisterXxxServer(...)")
	require.Contains(t, readmeText, "GRPCConnFactory + pb.NewXxxClient(conn)")
	require.NotContains(t, readmeText, "services/*/start.go")
	require.NotContains(t, readmeText, "third_party/")

	grpcReadme, err := os.ReadFile(filepath.Join(projectDir, "services", "order", "internal", "server", "grpc", "README.md"))
	require.NoError(t, err)
	require.Contains(t, string(grpcReadme), "Proto-first gRPC register / adapter")

	middlewareReadme, err := os.ReadFile(filepath.Join(projectDir, "services", "order", "internal", "server", "http", "middleware", "README.md"))
	require.NoError(t, err)
	require.Contains(t, string(middlewareReadme), "Service-local HTTP middleware")

	pkgReadme, err := os.ReadFile(filepath.Join(projectDir, "pkg", "README.md"))
	require.NoError(t, err)
	pkgReadmeText := string(pkgReadme)
	require.Contains(t, pkgReadmeText, "stable helpers shared by multiple services")
	require.Contains(t, pkgReadmeText, "reused by at least two services")
	require.Contains(t, pkgReadmeText, "infrastructure SDK initialization")

	makefile, err := os.ReadFile(filepath.Join(projectDir, "Makefile"))
	require.NoError(t, err)
	makeText := string(makefile)
	require.Contains(t, makeText, "USER_IMAGE ?= mfwverify-user-service")
	require.Contains(t, makeText, "ORDER_IMAGE ?= mfwverify-order-service")
	require.Contains(t, makeText, "PRODUCT_IMAGE ?= mfwverify-product-service")
	require.NotContains(t, makeText, "docker build -f deploy/docker/Dockerfile.user -t user-service:$(IMAGE_TAG) .")

	compose, err := os.ReadFile(filepath.Join(projectDir, "deploy", "compose", "docker-compose.yaml"))
	require.NoError(t, err)
	composeText := string(compose)
	require.Contains(t, composeText, "container_name: mfwverify-redis")
	require.Contains(t, composeText, "container_name: mfwverify-user")
	require.Contains(t, composeText, "container_name: mfwverify-order")
	require.Contains(t, composeText, "container_name: mfwverify-product")

	deployment, err := os.ReadFile(filepath.Join(projectDir, "deploy", "kubernetes", "base", "deployment.yaml"))
	require.NoError(t, err)
	deploymentText := string(deployment)
	require.Contains(t, deploymentText, "image: your-registry/mfwverify-user-service:latest")
	require.Contains(t, deploymentText, "image: your-registry/mfwverify-order-service:latest")
	require.Contains(t, deploymentText, "image: your-registry/mfwverify-product-service:latest")

	devKustomization, err := os.ReadFile(filepath.Join(projectDir, "deploy", "kubernetes", "overlays", "dev", "kustomization.yaml"))
	require.NoError(t, err)
	devText := string(devKustomization)
	require.Contains(t, devText, "name: your-registry/mfwverify-user-service")
	require.Contains(t, devText, "newName: mfwverify-user-service")
	require.Contains(t, devText, "name: your-registry/mfwverify-order-service")
	require.Contains(t, devText, "newName: mfwverify-order-service")
	require.Contains(t, devText, "name: your-registry/mfwverify-product-service")
	require.Contains(t, devText, "newName: mfwverify-product-service")

	userSvc, err := os.ReadFile(filepath.Join(projectDir, "services", "user", "internal", "service", "service.go"))
	require.NoError(t, err)
	require.Contains(t, string(userSvc), "gorp.GetGRPCTraceID")
	require.Contains(t, string(userSvc), "gorp.FromServerContext")
	require.NotContains(t, string(userSvc), "framework/provider/grpc")
	require.NotContains(t, string(userSvc), "github.com/ngq/gorp/app/grpc")

	serverFiles := []string{
		"services/user/internal/server/http/routes.go",
		"services/order/internal/server/http/routes.go",
		"services/product/internal/server/http/routes.go",
		"services/user/internal/server/http/handler/user.go",
		"services/order/internal/server/http/handler/order.go",
		"services/product/internal/server/http/handler/product.go",
	}
	for _, rel := range serverFiles {
		content, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(rel)))
		require.NoError(t, err, rel)
		require.NotContains(t, string(content), "framework/provider/gin", rel)
		require.NotContains(t, string(content), "ginprovider", rel)
	}

	handlerFiles := []string{
		"services/user/internal/server/http/handler/user.go",
		"services/order/internal/server/http/handler/order.go",
		"services/product/internal/server/http/handler/product.go",
	}
	for _, rel := range handlerFiles {
		content, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(rel)))
		require.NoError(t, err, rel)
		require.Contains(t, string(content), "c.JSON(", rel)
	}

	structure, err := os.ReadFile(filepath.Join(projectDir, "docs", "structure.md"))
	require.NoError(t, err)
	structureText := string(structure)
	require.Contains(t, structureText, "`internal/server/http/`")
	require.Contains(t, structureText, "`internal/server/grpc/`")

	componentTree, err := os.ReadFile(filepath.Join(projectDir, "docs", "component-tree.md"))
	require.NoError(t, err)
	componentTreeText := string(componentTree)
	require.Contains(t, componentTreeText, "`services/<service>/internal/server/http/`")
	require.Contains(t, componentTreeText, "`services/<service>/internal/server/grpc/`")

	userPB, err := os.ReadFile(filepath.Join(projectDir, "proto", "user", "v1", "user.pb.go"))
	require.NoError(t, err)
	userPBText := string(userPB)
	require.NotContains(t, userPBText, "{{.ModuleName}}")
	require.Contains(t, userPBText, "example.com/mfwverify/proto/user/v1;userv1")

	assertGeneratedProjectHasNoTemplateArtifacts(t, projectDir)
}
