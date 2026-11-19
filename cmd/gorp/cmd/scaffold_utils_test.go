package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	frameworktesting "github.com/ngq/gorp/framework/testing"
	"github.com/stretchr/testify/require"
)

func TestPrintScaffoldNextMatchesStarterEntrypoint(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		want         []string
		notWant      []string
	}{
		{
			name:         "golayout",
			templateName: starterTemplateGoLayout,
			want:         []string{"go mod tidy", "go run ./cmd/app"},
		},
		{
			name:         "multi-flat-wire",
			templateName: starterTemplateMultiFlatWire,
			want:         []string{"go mod tidy", "make generate", "make run-user"},
			notWant:      []string{"go run ./cmd/app"},
		},
		{
			name:         "multi-independent",
			templateName: starterTemplateMultiIndependent,
			want:         []string{"go work sync", "make generate", "make run-user"},
			notWant:      []string{"go run ./cmd/app"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			printScaffoldNext(buf, "example", tt.templateName)
			out := buf.String()
			for _, want := range tt.want {
				require.Contains(t, out, want)
			}
			for _, notWant := range tt.notWant {
				require.NotContains(t, out, notWant)
			}
		})
	}
}

func TestMultiStarterGoModUsesContribVersionAndReplacesWhenLocal(t *testing.T) {
	require.NoError(t, frameworktesting.ChdirRepoRoot())

	t.Run("remote mode pins contrib version independently", func(t *testing.T) {
		projectDir := filepath.Join(t.TempDir(), "remote-mfw")
		data := buildScaffoldData(scaffoldInput{
			Name:             "remote-mfw",
			Module:           "example.com/remote-mfw",
			FrameworkModule:  "github.com/ngq/gorp",
			FrameworkVersion: defaultFrameworkVersion,
			Backend:          "gorm",
		})
		require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiFlatWire), projectDir, data))
		raw, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
		require.NoError(t, err)
		text := string(raw)
		require.Contains(t, text, "github.com/ngq/gorp "+defaultFrameworkVersion)
		require.Contains(t, text, "github.com/ngq/gorp/contrib/configsource/etcd "+defaultContribVersion)
		require.NotContains(t, text, "replace ")
	})

	t.Run("local mode replaces root and contrib submodules", func(t *testing.T) {
		projectDir := filepath.Join(t.TempDir(), "local-mi")
		data := buildScaffoldData(scaffoldInput{
			Name:             "local-mi",
			Module:           "example.com/local-mi",
			FrameworkModule:  "github.com/ngq/gorp",
			FrameworkPath:    ".",
			FrameworkVersion: "v0.0.0",
			Backend:          "gorm",
		})
		require.NoError(t, renderTemplateProject(projectTemplateFS, resolveOfflineTemplateRoot(starterTemplateMultiIndependent), projectDir, data))
		raw, err := os.ReadFile(filepath.Join(projectDir, "services", "user", "go.mod"))
		require.NoError(t, err)
		text := string(raw)
		require.Contains(t, text, "github.com/ngq/gorp/contrib/registry/etcd "+defaultContribVersion)
		require.Contains(t, text, "github.com/ngq/gorp/contrib/registry/etcd =>")
	})
}
