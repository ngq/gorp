package p2c

import (
	"context"
	"errors"
	"testing"

	"github.com/ngq/gorp/framework/contract"
)

func TestP2CSelector_Select_EmptyInstances(t *testing.T) {
	selector := NewP2CSelector()
	ctx := context.Background()

	_, done, err := selector.Select(ctx, nil)
	if err != contract.ErrNoAvailable {
		t.Errorf("expected ErrNoAvailable, got: %v", err)
	}

	done(ctx, contract.DoneInfo{})
}

func TestP2CSelector_Select_SingleInstance(t *testing.T) {
	selector := NewP2CSelector()
	ctx := context.Background()

	instances := []contract.ServiceInstance{
		{ID: "1", Address: "inst1:8080", Healthy: true},
	}

	selected, done, err := selector.Select(ctx, instances)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if selected.ID != "1" {
		t.Errorf("expected instance 1, got: %s", selected.ID)
	}

	done(ctx, contract.DoneInfo{})
}

func TestP2CSelector_Select_DoneFuncUpdatesStats(t *testing.T) {
	selector := NewP2CSelector()
	ctx := context.Background()

	instances := []contract.ServiceInstance{
		{ID: "1", Address: "inst1:8080", Healthy: true},
		{ID: "2", Address: "inst2:8080", Healthy: true},
	}

	// 选择实例
	selected, done, err := selector.Select(ctx, instances)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 验证 pending 增加
	stats := selector.instanceStats[selected.Address]
	if stats == nil || stats.pending != 1 {
		t.Errorf("expected pending=1, got: %v", stats)
	}

	// 调用 DoneFunc（成功）
	done(ctx, contract.DoneInfo{Err: nil})

	// 验证 pending 减少，成功计数增加
	stats = selector.instanceStats[selected.Address]
	if stats.pending != 0 {
		t.Errorf("expected pending=0, got: %d", stats.pending)
	}
	if stats.successCount != 1 {
		t.Errorf("expected successCount=1, got: %d", stats.successCount)
	}
}

func TestP2CSelector_Select_DoneFuncWithError(t *testing.T) {
	selector := NewP2CSelector()
	ctx := context.Background()

	instances := []contract.ServiceInstance{
		{ID: "1", Address: "inst1:8080", Healthy: true},
		{ID: "2", Address: "inst2:8080", Healthy: true},
	}

	selected, done, err := selector.Select(ctx, instances)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 调用 DoneFunc（失败）
	done(ctx, contract.DoneInfo{Err: errors.New("connection failed")})

	// 验证失败计数增加
	stats := selector.instanceStats[selected.Address]
	if stats.failCount != 1 {
		t.Errorf("expected failCount=1, got: %d", stats.failCount)
	}
}

func TestP2CSelector_Select_LowerLoadInstance(t *testing.T) {
	selector := NewP2CSelector()
	ctx := context.Background()

	instances := []contract.ServiceInstance{
		{ID: "high", Address: "high:8080", Healthy: true},
		{ID: "low", Address: "low:8080", Healthy: true},
	}

	// 手动设置负载：high 实例高负载
	selector.mu.Lock()
	selector.instanceStats["high:8080"] = &InstanceStats{
		pending:      10,
		successCount: 100,
		failCount:    20, // 高失败率
	}
	selector.instanceStats["low:8080"] = &InstanceStats{
		pending:      2,
		successCount: 100,
		failCount:    0, // 低失败率
	}
	selector.mu.Unlock()

	// 多次选择，低负载实例应被选中更多
	counts := make(map[string]int)
	for i := 0; i < 100; i++ {
		selected, done, err := selector.Select(ctx, instances)
		if err != nil {
			continue
		}
		counts[selected.ID]++
		done(ctx, contract.DoneInfo{Err: nil})
	}

	// 低负载实例应被选中更多
	if counts["low"] < counts["high"] {
		t.Errorf("expected low-load instance selected more, got high=%d, low=%d", counts["high"], counts["low"])
	}
}

func TestP2CSelector_Select_ForceInstance(t *testing.T) {
	selector := NewP2CSelector()
	ctx := context.Background()

	instances := []contract.ServiceInstance{
		{ID: "1", Address: "inst1:8080", Healthy: true},
	}

	forced := contract.ServiceInstance{ID: "forced", Address: "forced:8080"}
	selected, done, err := selector.Select(
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

	// 验证 pending 增加
	stats := selector.instanceStats["forced:8080"]
	if stats == nil || stats.pending != 1 {
		t.Errorf("expected pending=1 for forced instance")
	}

	done(ctx, contract.DoneInfo{})
}

func TestP2CSelector_CalculateScore(t *testing.T) {
	selector := NewP2CSelector()

	// 测试无统计的实例（新实例）
	inst := contract.ServiceInstance{Address: "new:8080"}
	score := selector.calculateScore(inst)
	if score != 0.0 {
		t.Errorf("expected score=0 for new instance, got: %f", score)
	}

	// 测试高负载实例
	selector.mu.Lock()
	selector.instanceStats["high:8080"] = &InstanceStats{
		pending:      10,
		successCount: 90,
		failCount:    10, // 10% 失败率
		totalLatency: 9000, // 100ms 平均延迟
	}
	selector.mu.Unlock()

	instHigh := contract.ServiceInstance{Address: "high:8080"}
	scoreHigh := selector.calculateScore(instHigh)

	// 测试低负载实例
	selector.mu.Lock()
	selector.instanceStats["low:8080"] = &InstanceStats{
		pending:      2,
		successCount: 100,
		failCount:    0,
		totalLatency: 1000, // 10ms 平均延迟
	}
	selector.mu.Unlock()

	instLow := contract.ServiceInstance{Address: "low:8080"}
	scoreLow := selector.calculateScore(instLow)

	// 高负载实例评分应更高
	if scoreHigh <= scoreLow {
		t.Errorf("expected high-load score > low-load score, got high=%f, low=%f", scoreHigh, scoreLow)
	}
}