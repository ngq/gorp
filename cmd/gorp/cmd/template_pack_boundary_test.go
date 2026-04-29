package cmd

import "testing"

func TestDefaultReleaseTemplateAssetMapping(t *testing.T) {
	cases := map[string]string{
		"":                          "gorp-template-golayout.zip",
		starterTemplateGoLayout:         "gorp-template-golayout.zip",
		starterTemplateMultiFlatWire:    "gorp-template-multi-flat-wire.zip",
		starterTemplateMultiIndependent: "gorp-template-multi-independent.zip",
	}

	for input, expected := range cases {
		if got := defaultReleaseTemplateAsset(input); got != expected {
			t.Fatalf("template %q expected asset %q, got %q", input, expected, got)
		}
	}
}

func TestReleaseTemplateSourceMatchesTemplateType(t *testing.T) {
	srcFS, srcRoot := releaseTemplateSource(starterTemplateGoLayout)
	if srcFS == nil {
		t.Fatalf("expected release FS for golayout template")
	}
	if srcRoot != "templates/release/golayout/project" {
		t.Fatalf("expected golayout release root templates/release/golayout/project, got %q", srcRoot)
	}

	_, multiRoot := releaseTemplateSource(starterTemplateMultiFlatWire)
	if multiRoot != "templates/multi-flat-wire/project" {
		t.Fatalf("expected multi-flat-wire release root templates/multi-flat-wire/project, got %q", multiRoot)
	}

	_, independentRoot := releaseTemplateSource(starterTemplateMultiIndependent)
	if independentRoot != "templates/multi-independent/project" {
		t.Fatalf("expected multi-independent release root templates/multi-independent/project, got %q", independentRoot)
	}
}
