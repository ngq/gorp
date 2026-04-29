package cmd

import "testing"

func TestOfflineAndReleaseTemplateSupportBoundaries(t *testing.T) {
	offlineAllowed := []string{"", starterTemplateGoLayout, starterTemplateMultiFlatWire, starterTemplateMultiIndependent}
	for _, name := range offlineAllowed {
		if err := validateStarterTemplate(name); err != nil {
			t.Fatalf("offline template %q should be allowed: %v", name, err)
		}
	}

	if err := validateStarterTemplate("base"); err == nil {
		t.Fatalf("base should not be treated as public starter template")
	}

	releaseAllowed := []string{"", starterTemplateGoLayout, starterTemplateMultiFlatWire, starterTemplateMultiIndependent}
	for _, name := range releaseAllowed {
		if err := validateReleaseStarterTemplate(name); err != nil {
			t.Fatalf("release template %q should be allowed: %v", name, err)
		}
	}

	if err := validateReleaseStarterTemplate("base"); err == nil {
		t.Fatalf("base should not be allowed for public release templates")
	}
}

func TestResolveReleaseTemplateRootMatchesReleaseAssetPolicy(t *testing.T) {
	cases := map[string]string{
		starterTemplateGoLayout:         "templates/release/golayout/project",
		starterTemplateMultiFlatWire:    "templates/multi-flat-wire/project",
		starterTemplateMultiIndependent: "templates/multi-independent/project",
	}

	for name, expected := range cases {
		if got := resolveReleaseTemplateRoot(name); got != expected {
			t.Fatalf("template %q expected release root %q, got %q", name, expected, got)
		}
	}
}
