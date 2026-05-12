package noop

import (
	"context"
	"testing"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// TestNoopLoadShedder_AllowAlwaysSucceeds 验证 noop 实现总是允许请求通过。
func TestNoopLoadShedder_AllowAlwaysSucceeds(t *testing.T) {
	p := NewProvider()
	if p.Name() != "loadshedding.noop" {
		t.Fatalf("expected name loadshedding.noop, got %s", p.Name())
	}
	if !p.IsDefer() {
		t.Fatal("expected IsDefer to be true")
	}
	provides := p.Provides()
	if len(provides) != 1 || provides[0] != resiliencecontract.LoadShedderKey {
		t.Fatalf("expected Provides to return [%s], got %v", resiliencecontract.LoadShedderKey, provides)
	}
}

// TestNoopLoadShedder_AllowAlwaysSucceeds 验证 noop LoadShedder 总是允许请求通过。
func TestNoopLoadShedder_Integration(t *testing.T) {
	var ls noopLoadShedder
	ctx := context.Background()

	// Allow 应该总是返回 nil
	if err := ls.Allow(ctx, "any-resource"); err != nil {
		t.Fatalf("expected noop Allow to always succeed, got %v", err)
	}

	// Done 不应该 panic
	ls.Done(ctx, "any-resource", nil)
}
