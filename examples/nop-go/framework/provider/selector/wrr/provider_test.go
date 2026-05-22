// Package wrr_test provides unit tests for the weighted round-robin load balancer selector.
//
// 适用场景：
// - 验证 WRR（加权轮询）Selector 的权重选择与轮询行为。
package wrr

import (
	"context"
	"errors"
	"fmt"
	"testing"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// TestWRRSelector_Select_EmptyInstances verifies that selecting from empty instances returns ErrNoAvailable.
//
// TestWRRSelector_Select_EmptyInstances 验证从空实例中选择会返回 ErrNoAvailable。
func TestWRRSelector_Select_EmptyInstances(t *testing.T) {
	selector := NewWRRSelector()
	ctx := context.Background()

	_, done, err := selector.Select(ctx, nil)
	if !errors.Is(err, discoverycontract.ErrNoAvailable) {
		t.Errorf("expected ErrNoAvailable, got: %v", err)
	}

	done(ctx, discoverycontract.DoneInfo{})
}

// TestWRRSelector_Select_WeightDistribution verifies that instances are selected according to their weight.
//
// TestWRRSelector_Select_WeightDistribution 验证实例根据其权重被选中。
func TestWRRSelector_Select_WeightDistribution(t *testing.T) {
	selector := NewWRRSelector()
	ctx := context.Background()

	// 创建不同权重的实例
	instances := []transportcontract.ServiceInstance{
		{ID: "1", Address: "inst1:8080", Healthy: true, Metadata: map[string]string{"weight": "100"}},
		{ID: "2", Address: "inst2:8080", Healthy: true, Metadata: map[string]string{"weight": "50"}},
		{ID: "3", Address: "inst3:8080", Healthy: true, Metadata: map[string]string{"weight": "50"}},
	}

	// 多次选择，验证权重分布
	counts := make(map[string]int)
	for i := 0; i < 100; i++ {
		selected, _, err := selector.Select(ctx, instances)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		counts[selected.ID]++
	}

	// 权重 100 的实例应该被选中最多
	if counts["1"] < 30 {
		t.Errorf("weight-100 instance selected too few: %d", counts["1"])
	}
	// 权重 50 的实例应该被选中较少
	if counts["2"] > counts["1"] {
		t.Errorf("weight-50 instance selected more than weight-100")
	}
}

// TestWRRSelector_Select_AllSameWeight verifies that instances with equal weight are selected evenly.
//
// TestWRRSelector_Select_AllSameWeight 验证具有相同权重的实例被均匀选中。
func TestWRRSelector_Select_AllSameWeight(t *testing.T) {
	selector := NewWRRSelector()
	ctx := context.Background()

	// 所有实例权重相同
	instances := []transportcontract.ServiceInstance{
		{ID: "1", Address: "inst1:8080", Healthy: true},
		{ID: "2", Address: "inst2:8080", Healthy: true},
		{ID: "3", Address: "inst3:8080", Healthy: true},
	}

	counts := make(map[string]int)
	for i := 0; i < 30; i++ {
		selected, _, err := selector.Select(ctx, instances)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		counts[selected.ID]++
	}

	// 相同权重应该均匀分布（每个约 10 次）
	for id, count := range counts {
		if count < 5 || count > 15 {
			t.Errorf("instance %s count out of range: %d", id, count)
		}
	}
}

// TestWRRSelector_Select_ForceInstance verifies that the selector respects the ForceInstance option.
//
// TestWRRSelector_Select_ForceInstance 验证 selector 遵循 ForceInstance 选项。
func TestWRRSelector_Select_ForceInstance(t *testing.T) {
	selector := NewWRRSelector()
	ctx := context.Background()

	instances := []transportcontract.ServiceInstance{
		{ID: "1", Address: "inst1:8080", Healthy: true},
		{ID: "2", Address: "inst2:8080", Healthy: true},
	}

	forced := transportcontract.ServiceInstance{ID: "forced", Address: "forced:8080"}
	selected, _, err := selector.Select(
		ctx,
		instances,
		discoverycontract.WithForceInstance(forced),
	)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if selected.ID != "forced" {
		t.Errorf("expected forced instance, got: %s", selected.ID)
	}
}

// TestWRRSelector_CleanupStaleWeights verifies that stale weight entries are cleaned up when instance set changes.
//
// TestWRRSelector_CleanupStaleWeights 验证当实例集变更时，陈旧的权重条目被清理。
func TestWRRSelector_CleanupStaleWeights(t *testing.T) {
	selector := NewWRRSelector()
	ctx := context.Background()

	// 先选择一些实例，建立权重状态
	oldInstances := []transportcontract.ServiceInstance{
		{ID: "1", Address: "inst1:8080", Healthy: true, Metadata: map[string]string{"weight": "100"}},
		{ID: "2", Address: "inst2:8080", Healthy: true, Metadata: map[string]string{"weight": "50"}},
	}

	for i := 0; i < 10; i++ {
		selector.Select(ctx, oldInstances)
	}

	// 验证权重状态已建立
	if len(selector.currentWeight) != 2 {
		t.Errorf("expected 2 weights, got: %d", len(selector.currentWeight))
	}

	// 切换到新实例列表（不含 inst2）
	newInstances := []transportcontract.ServiceInstance{
		{ID: "1", Address: "inst1:8080", Healthy: true, Metadata: map[string]string{"weight": "100"}},
		{ID: "3", Address: "inst3:8080", Healthy: true, Metadata: map[string]string{"weight": "50"}},
	}

	selector.Select(ctx, newInstances)

	// 验证废弃权重已清理（inst2 权重应被删除）
	if _, exists := selector.currentWeight["inst2:8080"]; exists {
		t.Errorf("stale weight for inst2 should be cleaned")
	}
}

// TestWRRSelector_FallbackToP2CForLargeInstanceSet verifies that WRR falls back
// to P2C when the instance count exceeds wrrP2CFallbackThreshold.
//
// TestWRRSelector_FallbackToP2CForLargeInstanceSet 验证实例数量超过阈值后
// WRR 自动降级到 P2C，不再使用 WRR 权重计算。
func TestWRRSelector_FallbackToP2CForLargeInstanceSet(t *testing.T) {
	selector := NewWRRSelector()
	ctx := context.Background()

	// 构造超过阈值的实例集
	instances := make([]transportcontract.ServiceInstance, wrrP2CFallbackThreshold+10)
	for i := range instances {
		instances[i] = transportcontract.ServiceInstance{
			ID:       fmt.Sprintf("inst-%d", i),
			Address:  fmt.Sprintf("inst-%d:8080", i),
			Healthy:  true,
			Metadata: map[string]string{"weight": "100"},
		}
	}

	// 大实例集应正常返回实例（降级到 P2C）
	selected, _, err := selector.Select(ctx, instances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected.ID == "" {
		t.Error("expected non-empty selected instance")
	}

	// 降级到 P2C 后，currentWeight 不应被更新
	if len(selector.currentWeight) != 0 {
		t.Errorf("expected no WRR weight updates when fallback to P2C, got %d entries", len(selector.currentWeight))
	}
}
