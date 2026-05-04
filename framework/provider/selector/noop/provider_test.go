package noop

import (
	"context"
	"testing"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

func TestNoopSelector_Select_EmptyInstances(t *testing.T) {
	selector := &noopSelector{}
	ctx := context.Background()

	// 无实例时返回 ErrNoAvailable
	selected, done, err := selector.Select(ctx, nil)
	if err != discoverycontract.ErrNoAvailable {
		t.Errorf("expected ErrNoAvailable, got: %v", err)
	}
	if selected.Address != "" {
		t.Errorf("expected empty instance, got: %s", selected.Address)
	}

	// 调用 done 不应 panic
	done(ctx, discoverycontract.DoneInfo{})
}

func TestNoopSelector_Select_WithHealthyInstance(t *testing.T) {
	selector := &noopSelector{}
	ctx := context.Background()

	instances := []transportcontract.ServiceInstance{
		{ID: "1", Name: "svc", Address: "192.168.1.1:8080", Healthy: true},
		{ID: "2", Name: "svc", Address: "192.168.1.2:8080", Healthy: false},
	}

	// 返回第一个健康实例
	selected, done, err := selector.Select(ctx, instances)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if selected.Address != "192.168.1.1:8080" {
		t.Errorf("expected first healthy instance, got: %s", selected.Address)
	}

	done(ctx, discoverycontract.DoneInfo{})
}

func TestNoopSelector_Select_ForceInstance(t *testing.T) {
	selector := &noopSelector{}
	ctx := context.Background()

	forced := transportcontract.ServiceInstance{ID: "forced", Address: "forced:8080"}
	selected, done, err := selector.Select(
		ctx,
		nil,
		discoverycontract.WithForceInstance(forced),
	)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if selected.ID != "forced" {
		t.Errorf("expected forced instance, got: %s", selected.ID)
	}

	done(ctx, discoverycontract.DoneInfo{})
}

func TestNoopProvider_Register(t *testing.T) {
	// 测试 Provider 注册逻辑
	p := NewProvider()

	if p.Name() != "selector.noop" {
		t.Errorf("expected name 'selector.noop', got: %s", p.Name())
	}
	if !p.IsDefer() {
		t.Errorf("expected IsDefer=true")
	}
	if len(p.Provides()) != 2 {
		t.Errorf("expected 2 provides, got: %d", len(p.Provides()))
	}
}
