package application

import (
	"context"
	"testing"

	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

func TestServiceIdentityContextHelpers(t *testing.T) {
	identity := &securitycontract.ServiceIdentity{
		ServiceID:   "svc-1",
		ServiceName: "demo",
	}
	ctx := WithServiceIdentity(context.Background(), identity)

	got, ok := FromServiceIdentity(ctx)
	if !ok {
		t.Fatal("expected service identity from context")
	}
	if got.ServiceID != "svc-1" || got.ServiceName != "demo" {
		t.Fatalf("unexpected service identity: %#v", got)
	}
}
