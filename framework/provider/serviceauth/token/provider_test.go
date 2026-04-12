package token

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
)

func TestProviderProvidesServiceAuthKeysOnly(t *testing.T) {
	p := NewProvider()

	if p.Name() != "serviceauth.token" {
		t.Fatalf("unexpected provider name: %s", p.Name())
	}
	if !p.IsDefer() {
		t.Fatal("serviceauth.token provider should be defer")
	}

	provides := p.Provides()
	if len(provides) != 2 {
		t.Fatalf("unexpected provides len: %d, values: %#v", len(provides), provides)
	}

	expected := map[string]bool{
		contract.ServiceAuthKey:     true,
		contract.ServiceIdentityKey: true,
	}
	for _, key := range provides {
		if !expected[key] {
			t.Fatalf("unexpected provided key: %s", key)
		}
	}
}
