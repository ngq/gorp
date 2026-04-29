package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestTemplateSourcesRenderProjectArtifactsWithoutTemplateResidue(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	tests := []struct {
		name     string
		template string
		mustExist []string
	}{
		{
			name:     "golayout",
			template: starterTemplateGoLayout,
			mustExist: []string{".gorp-template.yml", "cmd/app/main.go", "app/http/routes.go", "deploy/kubernetes/base/kustomization.yaml"},
		},
		{
			name:     "multi-flat-wire",
			template: starterTemplateMultiFlatWire,
			mustExist: []string{".gorp-template.yml", "services/user/cmd/main.go", "docs/deploy.md", "deploy/kubernetes/overlays/prod/kustomization.yaml"},
		},
		{
			name:     "multi-independent",
			template: starterTemplateMultiIndependent,
			mustExist: []string{".gorp-template.yml", "services/user/cmd/main.go", "services/user/deploy/Dockerfile", "go.work"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			projectDir := filepath.Join(root, tt.name)
			data := buildScaffoldData(scaffoldInput{
				Name:            tt.name,
				Module:          "example.com/" + tt.name,
				FrameworkModule: "github.com/ngq/gorp",
				FrameworkPath:   ".",
				Backend:         "gorm",
				WithDB:          true,
				WithSwagger:     true,
			})

			require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(tt.template), projectDir, data))
			assertGeneratedProjectHasNoTemplateArtifacts(t, projectDir)

			for _, rel := range tt.mustExist {
				_, err := os.Stat(filepath.Join(projectDir, filepath.FromSlash(rel)))
				require.NoError(t, err, rel)
			}
		})
	}
}

func TestReleaseTemplateSourcesRenderProjectArtifactsWithoutTemplateResidue(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	tests := []struct {
		name      string
		template  string
		mustExist []string
	}{
		{
			name:      "release-golayout",
			template:  starterTemplateGoLayout,
			mustExist: []string{"cmd/app/main.go", "app/http/routes.go", "README.md"},
		},
		{
			name:      "release-multi-flat-wire",
			template:  starterTemplateMultiFlatWire,
			mustExist: []string{"services/user/cmd/main.go", "docs/deploy.md", "deploy/kubernetes/base/kustomization.yaml"},
		},
		{
			name:      "release-multi-independent",
			template:  starterTemplateMultiIndependent,
			mustExist: []string{"README.md", "services/user/cmd/main.go", "services/user/deploy/Dockerfile"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			projectDir := filepath.Join(root, tt.name)
			data := buildScaffoldData(scaffoldInput{
				Name:            tt.name,
				Module:          "example.com/" + tt.name,
				FrameworkModule: "github.com/ngq/gorp",
				FrameworkPath:   ".",
				Backend:         "gorm",
				WithDB:          true,
				WithSwagger:     true,
			})

			srcFS, srcRoot := releaseTemplateSource(tt.template)
			require.NoError(t, renderTemplateProject(srcFS, srcRoot, projectDir, data))
			assertGeneratedProjectHasNoTemplateArtifacts(t, projectDir)

			for _, rel := range tt.mustExist {
				_, err := os.Stat(filepath.Join(projectDir, filepath.FromSlash(rel)))
				require.NoError(t, err, rel)
			}
		})
	}
}
