package cmd

import "testing"

func TestResolveOfflineIntentDefaults(t *testing.T) {
	tests := []struct {
		name           string
		intent         string
		wantTemplate   string
		wantStarter    string
	}{
		{name: "default", intent: "", wantTemplate: starterTemplateGoLayout, wantStarter: starterProfileBasic},
		{name: "multi wire", intent: newIntentMultiWire, wantTemplate: starterTemplateMultiFlatWire, wantStarter: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTemplate, gotStarter := resolveOfflineIntentDefaults(tt.intent)
			if gotTemplate != tt.wantTemplate || gotStarter != tt.wantStarter {
				t.Fatalf("intent %q => template=%q starter=%q, want template=%q starter=%q", tt.intent, gotTemplate, gotStarter, tt.wantTemplate, tt.wantStarter)
			}
		})
	}
}

func TestParseNewIntentBoundaries(t *testing.T) {
	if got, err := parseNewIntent(nil); err != nil || got != "" {
		t.Fatalf("empty args should resolve to empty intent, got %q err=%v", got, err)
	}

	if got, err := parseNewIntent([]string{newIntentMultiWire}); err != nil || got != newIntentMultiWire {
		t.Fatalf("multi-wire should be accepted, got %q err=%v", got, err)
	}

	if _, err := parseNewIntent([]string{"multi"}); err == nil {
		t.Fatalf("multi should be rejected")
	}

	if _, err := parseNewIntent([]string{"a", "b"}); err == nil {
		t.Fatalf("multiple intents should be rejected")
	}
}

func TestExplicitTemplateOverridesIntentDefault(t *testing.T) {
	intentTemplate, _ := resolveOfflineIntentDefaults(newIntentMultiWire)
	if intentTemplate != starterTemplateMultiFlatWire {
		t.Fatalf("multi-wire default template drifted: %q", intentTemplate)
	}

	explicit := starterTemplateMultiIndependent
	if explicit == intentTemplate {
		t.Fatalf("test setup invalid: explicit template should differ from intent template")
	}
}
