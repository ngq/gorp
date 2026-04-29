package cmd

import (
	"path/filepath"
	"testing"
)

func TestNewFromReleaseTemplateAssetAndRootMapping(t *testing.T) {
	tests := []struct {
		name      string
		template  string
		wantAsset string
		wantRoot  string
	}{
		{name: "golayout", template: starterTemplateGoLayout, wantAsset: "gorp-template-golayout.zip", wantRoot: "templates/release/golayout/project"},
		{name: "multi-flat-wire", template: starterTemplateMultiFlatWire, wantAsset: "gorp-template-multi-flat-wire.zip", wantRoot: "templates/multi-flat-wire/project"},
		{name: "multi-independent", template: starterTemplateMultiIndependent, wantAsset: "gorp-template-multi-independent.zip", wantRoot: "templates/multi-independent/project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultReleaseTemplateAsset(tt.template); got != tt.wantAsset {
				t.Fatalf("template %q asset=%q, want %q", tt.template, got, tt.wantAsset)
			}
			if got := resolveReleaseTemplateRoot(tt.template); got != tt.wantRoot {
				t.Fatalf("template %q root=%q, want %q", tt.template, got, tt.wantRoot)
			}
		})
	}
}

func TestReleaseTemplateSourceProvidesProjectRoot(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		entryFile  string
	}{
		{name: "golayout", template: starterTemplateGoLayout, entryFile: "cmd/app/main.go.tmpl"},
		{name: "multi-flat-wire", template: starterTemplateMultiFlatWire, entryFile: ".gorp-template.yml.tmpl"},
		{name: "multi-independent", template: starterTemplateMultiIndependent, entryFile: ".gorp-template.yml.tmpl"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcFS, srcRoot := releaseTemplateSource(tt.template)
			if srcFS == nil {
				t.Fatalf("template %q returned nil fs", tt.template)
			}
			if _, err := srcFS.Open(filepath.ToSlash(filepath.Join(srcRoot, tt.entryFile))); err != nil {
				t.Fatalf("template %q missing entry file %s at %s: %v", tt.template, tt.entryFile, srcRoot, err)
			}
		})
	}
}

func TestReleaseTemplateValidationRejectsInternalBase(t *testing.T) {
	if err := validateReleaseStarterTemplate(starterTemplateBase); err == nil {
		t.Fatalf("internal base should not be accepted as public release template")
	}
}
