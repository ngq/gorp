package gorp

import (
	"testing"

	"github.com/ngq/gorp/framework/application"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

func TestGovernanceModeAliasesExposeStableValues(t *testing.T) {
	if GovernanceModeMonolith != resiliencecontract.GovernanceModeMonolith {
		t.Fatalf("expected GovernanceModeMonolith to match contract, got %q", GovernanceModeMonolith)
	}
	if GovernanceModeGinFirst != resiliencecontract.GovernanceModeGinFirst {
		t.Fatalf("expected GovernanceModeGinFirst to match contract, got %q", GovernanceModeGinFirst)
	}
	if GovernanceModeMicroservice != resiliencecontract.GovernanceModeMicroservice {
		t.Fatalf("expected GovernanceModeMicroservice to match contract, got %q", GovernanceModeMicroservice)
	}
}

func TestTopLevelHTTPServiceOptionsAliasSupportsGovernanceMode(t *testing.T) {
	opts := HTTPServiceOptions{GovernanceMode: GovernanceModeMicroservice}
	if opts.GovernanceMode != resiliencecontract.GovernanceModeMicroservice {
		t.Fatalf("expected HTTPServiceOptions.GovernanceMode to keep root alias value, got %q", opts.GovernanceMode)
	}
}

func TestTopLevelGovernanceOptionsWireIntoApplicationOptions(t *testing.T) {
	opts := []application.Option{
		HTTP(HTTPServiceOptions{GovernanceMode: GovernanceModeMonolith}),
		WithGovernanceMode(GovernanceModeMicroservice),
		WithGovernanceDisabled("tracing"),
		WithGovernanceProvider("serviceauth", "mtls"),
		WithMonolithMode(),
		WithGinFirstMode(),
		WithMicroserviceMode(),
	}

	if len(opts) != 7 {
		t.Fatalf("expected seven top-level governance helpers, got %d", len(opts))
	}
	for i, opt := range opts {
		if opt == nil {
			t.Fatalf("expected top-level option %d to be non-nil", i)
		}
	}
}
