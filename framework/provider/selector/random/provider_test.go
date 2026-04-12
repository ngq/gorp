package random

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
)

// TestRandomSelector_Select_EmptyInstances 测试空实例列表。
//
// 中文说明：
// - 当实例列表为空时，应返回 ErrNoAvailable 错误。
func TestRandomSelector_Select_EmptyInstances(t *testing.T) {
	selector := NewRandomSelector()

	_, _, err := selector.Select(context.Background(), nil)
	if err != contract.ErrNoAvailable {
		t.Errorf("expected ErrNoAvailable, got: %v", err)
	}

	_, _, err = selector.Select(context.Background(), []contract.ServiceInstance{})
	if err != contract.ErrNoAvailable {
		t.Errorf("expected ErrNoAvailable for empty slice, got: %v", err)
	}
}

// TestRandomSelector_Select_SingleInstance 测试单个实例。
//
// 中文说明：
// - 只有一个健康实例时，应始终选择该实例。
func TestRandomSelector_Select_SingleInstance(t *testing.T) {
	selector := NewRandomSelector()

	instances := []contract.ServiceInstance{
		{Name: "test", Address: "localhost:8080", Healthy: true},
	}

	selected, _, err := selector.Select(context.Background(), instances)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if selected.Address != "localhost:8080" {
		t.Errorf("expected localhost:8080, got: %s", selected.Address)
	}
}

// TestRandomSelector_Select_MultipleInstances 测试多个实例。
//
// 中文说明：
// - 多个健康实例时，应随机选择一个。
func TestRandomSelector_Select_MultipleInstances(t *testing.T) {
	selector := NewRandomSelector()

	instances := []contract.ServiceInstance{
		{Name: "test", Address: "localhost:8080", Healthy: true},
		{Name: "test", Address: "localhost:8081", Healthy: true},
		{Name: "test", Address: "localhost:8082", Healthy: true},
	}

	// 执行多次选择，验证随机性
	selectedAddrs := make(map[string]int)
	for i := 0; i < 100; i++ {
		selected, _, err := selector.Select(context.Background(), instances)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		selectedAddrs[selected.Address]++
	}

	// 验证所有实例都被选择过（随机性检查）
	if len(selectedAddrs) < 2 {
		t.Errorf("random selector should select multiple instances, got: %v", selectedAddrs)
	}
}

// TestRandomSelector_Select_UnhealthyInstances 测试不健康实例过滤。
//
// 中文说明：
// - 不健康的实例应被过滤掉。
func TestRandomSelector_Select_UnhealthyInstances(t *testing.T) {
	selector := NewRandomSelector()

	instances := []contract.ServiceInstance{
		{Name: "test", Address: "localhost:8080", Healthy: false},
		{Name: "test", Address: "localhost:8081", Healthy: true},
	}

	selected, _, err := selector.Select(context.Background(), instances)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if selected.Address != "localhost:8081" {
		t.Errorf("expected localhost:8081, got: %s", selected.Address)
	}
}

// TestRandomSelector_Select_AllUnhealthy 测试全部不健康实例。
//
// 中文说明：
// - 当所有实例都不健康时，应返回 ErrNoAvailable。
func TestRandomSelector_Select_AllUnhealthy(t *testing.T) {
	selector := NewRandomSelector()

	instances := []contract.ServiceInstance{
		{Name: "test", Address: "localhost:8080", Healthy: false},
		{Name: "test", Address: "localhost:8081", Healthy: false},
	}

	_, _, err := selector.Select(context.Background(), instances)
	if err != contract.ErrNoAvailable {
		t.Errorf("expected ErrNoAvailable, got: %v", err)
	}
}

// TestRandomSelector_Select_ForceInstance 测试强制指定实例。
//
// 中文说明：
// - 当 ForceInstance 选项指定时，应直接返回该实例。
func TestRandomSelector_Select_ForceInstance(t *testing.T) {
	selector := NewRandomSelector()

	instances := []contract.ServiceInstance{
		{Name: "test", Address: "localhost:8080", Healthy: true},
		{Name: "test", Address: "localhost:8081", Healthy: true},
	}

	forced := contract.ServiceInstance{Name: "test", Address: "forced:9999", Healthy: true}

	selected, _, err := selector.Select(context.Background(), instances,
		contract.WithForceInstance(forced),
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if selected.Address != "forced:9999" {
		t.Errorf("expected forced:9999, got: %s", selected.Address)
	}
}

// TestRandomSelector_Select_WithFilter 测试过滤器。
//
// 中文说明：
// - 过滤器应用于过滤实例列表。
func TestRandomSelector_Select_WithFilter(t *testing.T) {
	selector := NewRandomSelector()

	instances := []contract.ServiceInstance{
		{Name: "test", Address: "localhost:8080", Healthy: true, Metadata: map[string]string{"zone": "a"}},
		{Name: "test", Address: "localhost:8081", Healthy: true, Metadata: map[string]string{"zone": "b"}},
	}

	// 只选择 zone=a 的实例
	filter := func(instance contract.ServiceInstance) bool {
		return instance.Metadata["zone"] == "a"
	}

	selected, _, err := selector.Select(context.Background(), instances,
		contract.WithFilter(filter),
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if selected.Address != "localhost:8080" {
		t.Errorf("expected localhost:8080, got: %s", selected.Address)
	}
}

// TestRandomProvider_Register 测试 Provider 注册。
//
// 中文说明：
// - 验证 Provider 名称和提供的服务键正确。
func TestRandomProvider_Register(t *testing.T) {
	p := NewProvider()

	if p.Name() != "selector.random" {
		t.Errorf("expected name selector.random, got: %s", p.Name())
	}

	if !p.IsDefer() {
		t.Errorf("expected IsDefer to be true")
	}

	provides := p.Provides()
	if len(provides) != 2 {
		t.Errorf("expected 2 provided keys, got: %v", provides)
	}
	expected := map[string]bool{
		contract.SelectorKey:        true,
		contract.SelectorBuilderKey: true,
	}
	for _, key := range provides {
		if !expected[key] {
			t.Errorf("unexpected provided key: %s", key)
		}
	}
}