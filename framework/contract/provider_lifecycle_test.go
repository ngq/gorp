package contract

import "testing"

func TestServiceProviderLifecycleContractMeaning(t *testing.T) {
	var p ServiceProvider = providerContractStub{}

	if p.Name() == "" {
		t.Fatalf("provider name must not be empty")
	}
	if err := p.Register(nil); err != nil {
		t.Fatalf("register should be callable in contract test: %v", err)
	}
	if err := p.Boot(nil); err != nil {
		t.Fatalf("boot should be callable in contract test: %v", err)
	}
	if !p.IsDefer() {
		t.Fatalf("stub should represent deferred provider semantics")
	}
	if len(p.Provides()) == 0 {
		t.Fatalf("deferred provider should expose provided keys")
	}
}
