package gorp

import (
	"testing"

	"github.com/ngq/gorp/framework/application"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

func TestGovernanceModeAliasesExposeStableValues(t *testing.T) {
	if GovernanceModeMono != resiliencecontract.GovernanceModeMono {
		t.Fatalf("expected GovernanceModeMono to match contract, got %q", GovernanceModeMono)
	}
	if GovernanceModeMicro != resiliencecontract.GovernanceModeMicro {
		t.Fatalf("expected GovernanceModeMicro to match contract, got %q", GovernanceModeMicro)
	}
}

func TestHTTPModeAliasesExposeStableValues(t *testing.T) {
	if HTTPModeContract != resiliencecontract.HTTPModeContract {
		t.Fatalf("expected HTTPModeContract to match contract, got %q", HTTPModeContract)
	}
	if HTTPModeGin != resiliencecontract.HTTPModeGin {
		t.Fatalf("expected HTTPModeGin to match contract, got %q", HTTPModeGin)
	}
}

func TestTopLevelHTTPServiceOptionsAliasSupportsGovernanceMode(t *testing.T) {
	opts := HTTPServiceOptions{GovernanceMode: GovernanceModeMicro}
	if opts.GovernanceMode != resiliencecontract.GovernanceModeMicro {
		t.Fatalf("expected HTTPServiceOptions.GovernanceMode to keep root alias value, got %q", opts.GovernanceMode)
	}
}

func TestTopLevelGovernanceOptionsWireIntoApplicationOptions(t *testing.T) {
	opts := []application.Option{
		HTTP(HTTPServiceOptions{GovernanceMode: GovernanceModeMono}),
		WithGovernanceMode(GovernanceModeMicro),
		WithGovernanceDisabled("tracing"),
		WithGovernanceProvider("serviceauth", "mtls"),
		WithMonoMode(),
		WithMicroMode(),
		WithMicroGovernance(),
		WithMonoGovernance(),
		WithHTTPMode(HTTPModeGin),
		WithGinHTTPMode(),
		WithContractHTTPMode(),
	}

	if len(opts) != 11 {
		t.Fatalf("expected eleven top-level governance helpers, got %d", len(opts))
	}
	for i, opt := range opts {
		if opt == nil {
			t.Fatalf("expected top-level option %d to be non-nil", i)
		}
	}
}
