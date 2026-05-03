package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestToKubernetesName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "demo", want: "demo"},
		{in: "Demo_App", want: "demo-app"},
		{in: "  Demo App  ", want: "demo-app"},
		{in: "__", want: "gorp"},
		{in: "demo---svc", want: "demo-svc"},
	}

	for _, tt := range tests {
		require.Equal(t, tt.want, toKubernetesName(tt.in))
	}
}

func TestBuildScaffoldDataIncludesKubernetesNames(t *testing.T) {
	data := buildScaffoldData(scaffoldInput{
		Name:            "Demo_App",
		Module:          "example.com/demo-app",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.Equal(t, "demo-app", data["KubernetesName"])
	require.Equal(t, "demo-app-config", data["KubernetesConfigMapName"])
	require.Equal(t, "demo-app-secret", data["KubernetesSecretName"])
	require.Equal(t, "demo-app-dev", data["KubernetesNamespaceDev"])
	require.Equal(t, "demo-app-staging", data["KubernetesNamespaceStaging"])
	require.Equal(t, "demo-app-prod", data["KubernetesNamespaceProd"])
}

func TestPublicStartersRenderProjectScopedKubernetesNames(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	tests := []struct {
		name      string
		template  string
		project   string
		module    string
		namespace string
	}{
		{name: "golayout", template: starterTemplateGoLayout, project: "Demo_App", module: "example.com/demo-app", namespace: "demo-app-dev"},
		{name: "multi-flat-wire", template: starterTemplateMultiFlatWire, project: "Demo_App", module: "example.com/demo-app", namespace: "demo-app-dev"},
		{name: "multi-independent", template: starterTemplateMultiIndependent, project: "Demo_App", module: "example.com/demo-app", namespace: "demo-app-dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			projectDir := filepath.Join(root, "verify")
			data := buildScaffoldData(scaffoldInput{
				Name:            tt.project,
				Module:          tt.module,
				FrameworkModule: "github.com/ngq/gorp",
				FrameworkPath:   ".",
				Backend:         "gorm",
				WithDB:          true,
				WithSwagger:     true,
			})

			require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(tt.template), projectDir, data))

			var configMapPath, secretPath, deploymentPath, namespacePath string
			switch tt.template {
			case starterTemplateGoLayout:
				configMapPath = filepath.Join(projectDir, "deploy", "kubernetes", "base", "configmap.yaml")
				secretPath = filepath.Join(projectDir, "deploy", "kubernetes", "base", "secret.yaml")
				deploymentPath = filepath.Join(projectDir, "deploy", "kubernetes", "base", "deployment.yaml")
				namespacePath = filepath.Join(projectDir, "deploy", "kubernetes", "overlays", "dev", "kustomization.yaml")
			default:
				configMapPath = filepath.Join(projectDir, "deploy", "kubernetes", "base", "configmap.yaml")
				secretPath = filepath.Join(projectDir, "deploy", "kubernetes", "base", "secret.yaml")
				deploymentPath = filepath.Join(projectDir, "deploy", "kubernetes", "base", "deployment.yaml")
				namespacePath = filepath.Join(projectDir, "deploy", "kubernetes", "overlays", "dev", "kustomization.yaml")
			}

			configMap, err := os.ReadFile(configMapPath)
			require.NoError(t, err)
			require.Contains(t, string(configMap), "demo-app-config")

			secret, err := os.ReadFile(secretPath)
			require.NoError(t, err)
			require.Contains(t, string(secret), "demo-app-secret")

			deployment, err := os.ReadFile(deploymentPath)
			require.NoError(t, err)
			require.Contains(t, string(deployment), "demo-app-config")
			require.Contains(t, string(deployment), "demo-app-secret")

			namespaceDoc, err := os.ReadFile(namespacePath)
			require.NoError(t, err)
			require.Contains(t, string(namespaceDoc), "namespace: "+tt.namespace)

			if tt.template == starterTemplateGoLayout {
				deploymentText := string(deployment)
				require.Contains(t, deploymentText, "name: demo-app")
				require.Contains(t, deploymentText, "image: your-registry/demo-app:latest")

				serviceDoc, err := os.ReadFile(filepath.Join(projectDir, "deploy", "kubernetes", "base", "service.yaml"))
				require.NoError(t, err)
				require.Contains(t, string(serviceDoc), "name: demo-app")

				devKustomization, err := os.ReadFile(filepath.Join(projectDir, "deploy", "kubernetes", "overlays", "dev", "kustomization.yaml"))
				require.NoError(t, err)
				devText := string(devKustomization)
				require.Contains(t, devText, "name: your-registry/demo-app")
				require.Contains(t, devText, "newName: demo-app")

				stagingKustomization, err := os.ReadFile(filepath.Join(projectDir, "deploy", "kubernetes", "overlays", "staging", "kustomization.yaml"))
				require.NoError(t, err)
				stagingText := string(stagingKustomization)
				require.Contains(t, stagingText, "name: your-registry/demo-app")
				require.Contains(t, stagingText, "newName: your-registry/demo-app")
			} else {
				composeDoc, err := os.ReadFile(filepath.Join(projectDir, "deploy", "compose", "docker-compose.yaml"))
				require.NoError(t, err)
				composeText := string(composeDoc)
				require.Contains(t, composeText, "container_name: demo-app-redis")
				require.Contains(t, composeText, "container_name: demo-app-user")
				require.Contains(t, composeText, "container_name: demo-app-order")
				require.Contains(t, composeText, "container_name: demo-app-product")

				deploymentText := string(deployment)
				require.Contains(t, deploymentText, "image: your-registry/demo-app-user-service:latest")
				require.Contains(t, deploymentText, "image: your-registry/demo-app-order-service:latest")
				require.Contains(t, deploymentText, "image: your-registry/demo-app-product-service:latest")

				devKustomization, err := os.ReadFile(filepath.Join(projectDir, "deploy", "kubernetes", "overlays", "dev", "kustomization.yaml"))
				require.NoError(t, err)
				devText := string(devKustomization)
				require.Contains(t, devText, "name: your-registry/demo-app-user-service")
				require.Contains(t, devText, "newName: demo-app-user-service")
				require.Contains(t, devText, "name: your-registry/demo-app-order-service")
				require.Contains(t, devText, "newName: demo-app-order-service")
				require.Contains(t, devText, "name: your-registry/demo-app-product-service")
				require.Contains(t, devText, "newName: demo-app-product-service")
			}
		})
	}
}
