package wrr

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
)

func TestWRRSelector_Select_EmptyInstances(t *testing.T) {
	selector := NewWRRSelector()
	ctx := context.Background()

	_, done, err := selector.Select(ctx, nil)
	if err != contract.ErrNoAvailable {
		t.Errorf("expected ErrNoAvailable, got: %v", err)
	}

	done(ctx, contract.DoneInfo{})
}

func TestWRRSelector_Select_WeightDistribution(t *testing.T) {
	selector := NewWRRSelector()
	ctx := context.Background()

	// 创建不同权重的实例
	instances := []contract.ServiceInstance{
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

func TestWRRSelector_Select_AllSameWeight(t *testing.T) {
	selector := NewWRRSelector()
	ctx := context.Background()

	// 所有实例权重相同
	instances := []contract.ServiceInstance{
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

func TestWRRSelector_Select_ForceInstance(t *testing.T) {
	selector := NewWRRSelector()
	ctx := context.Background()

	instances := []contract.ServiceInstance{
		{ID: "1", Address: "inst1:8080", Healthy: true},
		{ID: "2", Address: "inst2:8080", Healthy: true},
	}

	forced := contract.ServiceInstance{ID: "forced", Address: "forced:8080"}
	selected, _, err := selector.Select(
		ctx,
		instances,
		contract.WithForceInstance(forced),
	)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if selected.ID != "forced" {
		t.Errorf("expected forced instance, got: %s", selected.ID)
	}
}

func TestWRRSelector_CleanupStaleWeights(t *testing.T) {
	selector := NewWRRSelector()
	ctx := context.Background()

	// 先选择一些实例，建立权重状态
	oldInstances := []contract.ServiceInstance{
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
	newInstances := []contract.ServiceInstance{
		{ID: "1", Address: "inst1:8080", Healthy: true, Metadata: map[string]string{"weight": "100"}},
		{ID: "3", Address: "inst3:8080", Healthy: true, Metadata: map[string]string{"weight": "50"}},
	}

	selector.Select(ctx, newInstances)

	// 验证废弃权重已清理（inst2 权重应被删除）
	if _, exists := selector.currentWeight["inst2:8080"]; exists {
		t.Errorf("stale weight for inst2 should be cleaned")
	}
}