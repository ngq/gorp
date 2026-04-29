package contract

import "testing"

type providerContractStub struct{}

func (providerContractStub) Name() string                 { return "stub" }
func (providerContractStub) Register(Container) error     { return nil }
func (providerContractStub) Boot(Container) error         { return nil }
func (providerContractStub) IsDefer() bool                { return true }
func (providerContractStub) Provides() []string           { return []string{"a", "b"} }

func TestServiceProviderContractSurface(t *testing.T) {
	var p ServiceProvider = providerContractStub{}

	if p.Name() != "stub" {
		t.Fatalf("expected provider name stub, got %q", p.Name())
	}
	if !p.IsDefer() {
		t.Fatalf("expected IsDefer to return true")
	}
	provides := p.Provides()
	if len(provides) != 2 || provides[0] != "a" || provides[1] != "b" {
		t.Fatalf("unexpected Provides result: %#v", provides)
	}
}
