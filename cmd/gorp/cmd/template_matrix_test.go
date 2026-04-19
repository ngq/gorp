package cmd

import (
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedStarterTemplateMatrixRenders(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	templates := []string{
		starterTemplateBase,
		starterTemplateGoLayout,
		starterTemplateGoLayoutWire,
		starterTemplateMultiFlat,
		starterTemplateMultiFlatWire,
	}

	for _, templateName := range templates {
		t.Run(templateName, func(t *testing.T) {
			root := t.TempDir()
			projectDir := filepath.Join(root, templateName)
			data := buildScaffoldData(scaffoldInput{
				Name:            templateName,
				Module:          "example.com/" + templateName,
				FrameworkModule: "github.com/ngq/gorp",
				FrameworkPath:   ".",
				Backend:         "gorm",
				WithDB:          true,
				WithSwagger:     true,
			})

			require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(templateName), projectDir, data))
			assertGeneratedProjectHasNoTemplateArtifacts(t, projectDir)
			marker := filepath.Join(projectDir, ".gorp-template.yml")
			if templateName != starterTemplateBase {
				require.FileExists(t, marker)
			}
		})
	}
}
