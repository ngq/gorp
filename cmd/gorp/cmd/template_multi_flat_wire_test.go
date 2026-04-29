package cmd

import (
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestReleaseTemplateSupportForMultiFlatWire(t *testing.T) {
	require.NoError(t, validateReleaseStarterTemplate(starterTemplateMultiFlatWire))
	require.Equal(t, "templates/multi-flat-wire/project", resolveReleaseTemplateRoot(starterTemplateMultiFlatWire))
	require.Equal(t, "gorp-template-multi-flat-wire.zip", defaultReleaseTemplateAsset(starterTemplateMultiFlatWire))

	srcFS, srcRoot := releaseTemplateSource(starterTemplateMultiFlatWire)
	_, err := srcFS.Open(filepath.ToSlash(filepath.Join(srcRoot, "README.md.tmpl")))
	require.NoError(t, err)
}

func TestMultiFlatWireUserProtoDescriptorTracksModuleName(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())
	root := t.TempDir()
	projectDir := filepath.Join(root, "mfwproto")
	data := buildScaffoldData(scaffoldInput{
		Name:            "mfwproto",
		Module:          "example.com/mfwproto",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiFlatWire), projectDir, data))

	content, err := os.ReadFile(filepath.Join(projectDir, "proto", "user", "v1", "user.pb.go"))
	require.NoError(t, err)
	text := string(content)
	require.Contains(t, text, "example.com/mfwproto/proto/user/v1;userv1")
	require.NotContains(t, text, "{{.ModuleName}}")
	require.NotContains(t, text, "${{.ModuleName}}")
}

func TestDetectProjectTemplateTypeForMultiFlatWire(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())
	root := t.TempDir()
	projectDir := filepath.Join(root, "mfwdetect")
	data := buildScaffoldData(scaffoldInput{
		Name:            "mfwdetect",
		Module:          "example.com/mfwdetect",
		FrameworkModule: "github.com/ngq/gorp",
		FrameworkPath:   ".",
		Backend:         "gorm",
		WithDB:          true,
		WithSwagger:     true,
	})

	require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiFlatWire), projectDir, data))

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(projectDir))
	defer func() { _ = os.Chdir(oldWd) }()

	require.Equal(t, starterTemplateMultiFlatWire, detectProjectTemplateType())
}
