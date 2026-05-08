package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestRenderMultiIndependentTemplateIncludesStarterMarkers(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	root := t.TempDir()
	projectDir := filepath.Join(root, "miverify")
	data := buildScaffoldData(scaffoldInput{
		Name:            "miverify",
		Module:          "example.com/miverify",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiIndependent), projectDir, data))

	mustExist := []string{
		".gorp-template.yml",
		"README.md",
		"Makefile",
		"go.work",
		"shared/README.md",
		"shared/go.mod",
		"shared/config/config.go",
		"shared/db/db.go",
		"shared/logger/logger.go",
		"docs/structure.md",
		"docs/component-tree.md",
		"docs/deploy.md",
		"docs/next-steps.md",
		"deploy/compose/docker-compose.yaml",
		"deploy/docker/Dockerfile.user",
		"deploy/docker/Dockerfile.order",
		"deploy/docker/Dockerfile.product",
		"services/user/go.mod",
		"services/order/go.mod",
		"services/product/go.mod",
		"services/user/cmd/main.go",
		"services/user/cmd/wire.go",
		"services/user/cmd/wire_gen.go",
		"services/order/cmd/main.go",
		"services/order/cmd/wire.go",
		"services/order/cmd/wire_gen.go",
		"services/product/cmd/main.go",
		"services/product/cmd/wire.go",
		"services/product/cmd/wire_gen.go",
		"services/user/internal/server/http/routes.go",
		"services/order/internal/server/http/routes.go",
		"services/product/internal/server/http/routes.go",
		"services/user/internal/server/http/handler/user.go",
		"services/order/internal/server/http/handler/order.go",
		"services/product/internal/server/http/handler/product.go",
		"services/user/internal/server/grpc/README.md",
		"services/order/internal/server/grpc/README.md",
		"services/product/internal/server/grpc/README.md",
		"services/user/internal/server/http/middleware/README.md",
		"services/order/internal/server/http/middleware/README.md",
		"services/product/internal/server/http/middleware/README.md",
	}

	for _, rel := range mustExist {
		path := filepath.Join(projectDir, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %s: %v", rel, err)
		}
	}

	mustNotExist := []string{
		"services/user/start.go",
		"services/order/start.go",
		"services/product/start.go",
		"services/gateway",
		"shared/infrastructure",
		"deployments",
		"deploy/k8s",
	}
	for _, rel := range mustNotExist {
		path := filepath.Join(projectDir, filepath.FromSlash(rel))
		_, err := os.Stat(path)
		require.Error(t, err, rel)
		require.True(t, os.IsNotExist(err), rel)
	}

	readme, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	require.NoError(t, err)
	readmeText := string(readme)
	require.Contains(t, readmeText, "make generate")
	require.Contains(t, readmeText, "make deploy-local")
	require.Contains(t, readmeText, "go work sync")
	require.Contains(t, readmeText, "`shared/`")
	require.Contains(t, readmeText, "`internal/server/http/`")
	require.Contains(t, readmeText, "`internal/server/grpc/`")
	require.Contains(t, readmeText, "gorp.Run(...)")
	require.Contains(t, readmeText, "WithMicroserviceMode")
	require.Contains(t, readmeText, "miverify-user-service")
	require.Contains(t, readmeText, "miverify-order-service")
	require.Contains(t, readmeText, "miverify-product-service")
	require.Contains(t, readmeText, "`miverify-*`")
	require.NotContains(t, readmeText, "framework/bootstrap")
	require.NotContains(t, readmeText, "shared/infrastructure")
	require.NotContains(t, readmeText, "start.go")

	makefile, err := os.ReadFile(filepath.Join(projectDir, "Makefile"))
	require.NoError(t, err)
	makeText := string(makefile)
	require.Contains(t, makeText, "USER_IMAGE ?= miverify-user-service")
	require.Contains(t, makeText, "ORDER_IMAGE ?= miverify-order-service")
	require.Contains(t, makeText, "PRODUCT_IMAGE ?= miverify-product-service")
	require.NotContains(t, makeText, "docker build -f deploy/docker/Dockerfile.user -t user-service:$(IMAGE_TAG) .")

	compose, err := os.ReadFile(filepath.Join(projectDir, "deploy", "compose", "docker-compose.yaml"))
	require.NoError(t, err)
	composeText := string(compose)
	require.Contains(t, composeText, "container_name: miverify-redis")
	require.Contains(t, composeText, "container_name: miverify-user")
	require.Contains(t, composeText, "container_name: miverify-order")
	require.Contains(t, composeText, "container_name: miverify-product")

	mainFile := filepath.Join(projectDir, "services", "user", "cmd", "main.go")
	mainContent, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	mainText := string(mainContent)
	require.Contains(t, mainText, "gorp.Run(")
	require.Contains(t, mainText, "gorp.HTTP()")
	require.Contains(t, mainText, "gorp.WithMicroserviceMode()")
	require.Contains(t, mainText, "gorp.WithMigrate(migrate)")
	require.Contains(t, mainText, "gorp.WithSetup(setup)")
	require.NotContains(t, mainText, "frameworkbootstrap.BootHTTPService(")

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

	structure, err := os.ReadFile(filepath.Join(projectDir, "docs", "structure.md"))
	require.NoError(t, err)
	structureText := string(structure)
	require.Contains(t, structureText, "`internal/server/http/`")
	require.Contains(t, structureText, "`internal/server/grpc/`")
	require.Contains(t, structureText, "`shared/`")
	require.NotContains(t, structureText, "`shared/infrastructure/`")

	componentTree, err := os.ReadFile(filepath.Join(projectDir, "docs", "component-tree.md"))
	require.NoError(t, err)
	componentTreeText := string(componentTree)
	require.Contains(t, componentTreeText, "`services/<service>/internal/server/http/`")
	require.Contains(t, componentTreeText, "`services/<service>/internal/server/grpc/`")
	require.Contains(t, componentTreeText, "`shared/`")

	grpcReadme, err := os.ReadFile(filepath.Join(projectDir, "services", "user", "internal", "server", "grpc", "README.md"))
	require.NoError(t, err)
	require.Contains(t, string(grpcReadme), "future Proto-first gRPC register / adapter code")

	middlewareReadme, err := os.ReadFile(filepath.Join(projectDir, "services", "user", "internal", "server", "http", "middleware", "README.md"))
	require.NoError(t, err)
	require.Contains(t, string(middlewareReadme), "Service-local HTTP middleware")

	deployDoc, err := os.ReadFile(filepath.Join(projectDir, "docs", "deploy.md"))
	require.NoError(t, err)
	deployText := string(deployDoc)
	require.Contains(t, deployText, "make deploy-local")
	require.Contains(t, deployText, "make harbor-push")
	require.Contains(t, deployText, "miverify-user-service")
	require.Contains(t, deployText, "miverify-order-service")
	require.Contains(t, deployText, "miverify-product-service")
	require.Contains(t, deployText, "`miverify-*`")
	require.Contains(t, deployText, "deploy/kubernetes/overlays/staging/kustomization.yaml")
	require.Contains(t, deployText, "deploy/kubernetes/overlays/prod/kustomization.yaml")

	configMap, err := os.ReadFile(filepath.Join(projectDir, "deploy", "kubernetes", "base", "configmap.yaml"))
	require.NoError(t, err)
	require.Contains(t, string(configMap), "miverify-config")
	require.NotContains(t, string(configMap), "multi-flat-wire-config")

	secret, err := os.ReadFile(filepath.Join(projectDir, "deploy", "kubernetes", "base", "secret.yaml"))
	require.NoError(t, err)
	require.Contains(t, string(secret), "miverify-secret")
	require.NotContains(t, string(secret), "multi-flat-wire-secret")

	deployment, err := os.ReadFile(filepath.Join(projectDir, "deploy", "kubernetes", "base", "deployment.yaml"))
	require.NoError(t, err)
	deploymentText := string(deployment)
	require.Contains(t, deploymentText, "miverify-config")
	require.Contains(t, deploymentText, "miverify-secret")
	require.Contains(t, deploymentText, "image: your-registry/miverify-user-service:latest")
	require.Contains(t, deploymentText, "image: your-registry/miverify-order-service:latest")
	require.Contains(t, deploymentText, "image: your-registry/miverify-product-service:latest")
	require.NotContains(t, deploymentText, "multi-flat-wire-config")
	require.NotContains(t, deploymentText, "multi-flat-wire-secret")

	devKustomization, err := os.ReadFile(filepath.Join(projectDir, "deploy", "kubernetes", "overlays", "dev", "kustomization.yaml"))
	require.NoError(t, err)
	devText := string(devKustomization)
	require.Contains(t, devText, "name: your-registry/miverify-user-service")
	require.Contains(t, devText, "newName: miverify-user-service")
	require.Contains(t, devText, "name: your-registry/miverify-order-service")
	require.Contains(t, devText, "newName: miverify-order-service")
	require.Contains(t, devText, "name: your-registry/miverify-product-service")
	require.Contains(t, devText, "newName: miverify-product-service")

	assertGeneratedProjectHasNoTemplateArtifacts(t, projectDir)
}
