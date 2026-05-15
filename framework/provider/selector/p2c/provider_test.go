// Package p2c_test provides unit tests for the P2C (power of two choices) load balancer selector.
//
// 适用场景：
// - 验证 P2C Selector 的实例选择、负载估算和自适应行为。
package p2c

import (
	"context"
	"errors"
	"testing"
	"time"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// TestP2CSelectorSelectEmptyInstances verifies the selector rejects empty instance sets.
//
// TestP2CSelectorSelectEmptyInstances 验证 selector 会拒绝空实例集合。
func TestP2CSelectorSelectEmptyInstances(t *testing.T) {
	selector := NewP2CSelector()

	_, done, err := selector.Select(context.Background(), nil)
	if !errors.Is(err, discoverycontract.ErrNoAvailable) {
		t.Fatalf("expected ErrNoAvailable, got %v", err)
	}

	done(context.Background(), discoverycontract.DoneInfo{})
}

// TestP2CSelectorSelectSingleInstance verifies the selector returns the only healthy instance.
//
// TestP2CSelectorSelectSingleInstance 验证 selector 会返回唯一健康实例。
func TestP2CSelectorSelectSingleInstance(t *testing.T) {
	selector := NewP2CSelector()
	instances := []transportcontract.ServiceInstance{
		{ID: "1", Address: "inst1:8080", Healthy: true},
	}

	selected, done, err := selector.Select(context.Background(), instances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected.ID != "1" {
		t.Fatalf("expected instance 1, got %s", selected.ID)
	}

	done(context.Background(), discoverycontract.DoneInfo{})
}

// TestP2CSelectorDoneFuncUpdatesCounters verifies success and failure feedback update selector stats.
//
// TestP2CSelectorDoneFuncUpdatesCounters 验证成功与失败反馈会更新 selector 统计信息。
func TestP2CSelectorDoneFuncUpdatesCounters(t *testing.T) {
	selector := NewP2CSelector()
	instances := []transportcontract.ServiceInstance{
		{ID: "1", Address: "inst1:8080", Healthy: true},
		{ID: "2", Address: "inst2:8080", Healthy: true},
	}

	selected, done, err := selector.Select(context.Background(), instances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats := selector.instanceStats[selected.Address]
	if stats == nil || stats.pending != 1 {
		t.Fatalf("expected pending=1, got %+v", stats)
	}

	done(context.Background(), discoverycontract.DoneInfo{Latency: 20 * time.Millisecond})

	stats = selector.instanceStats[selected.Address]
	if stats.pending != 0 {
		t.Fatalf("expected pending=0, got %d", stats.pending)
	}
	if stats.successCount != 1 {
		t.Fatalf("expected successCount=1, got %d", stats.successCount)
	}
	if stats.latencyEWMA <= 0 {
		t.Fatalf("expected latency EWMA to be updated, got %f", stats.latencyEWMA)
	}

	selected, done, err = selector.Select(context.Background(), instances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	done(context.Background(), discoverycontract.DoneInfo{Err: errors.New("connection failed")})

	stats = selector.instanceStats[selected.Address]
	if stats.failCount != 1 {
		t.Fatalf("expected failCount=1, got %d", stats.failCount)
	}
}

// TestP2CSelectorPrefersLowerLoad verifies the selector prefers lower-load instances over time.
//
// TestP2CSelectorPrefersLowerLoad 验证 selector 会逐步偏向低负载实例。
func TestP2CSelectorPrefersLowerLoad(t *testing.T) {
	selector := NewP2CSelector()
	instances := []transportcontract.ServiceInstance{
		{ID: "high", Address: "high:8080", Healthy: true},
		{ID: "low", Address: "low:8080", Healthy: true},
	}

	selector.mu.Lock()
	selector.instanceStats["high:8080"] = &InstanceStats{
		pending:      10,
		successCount: 100,
		failCount:    20,
		latencyEWMA:  200,
	}
	selector.instanceStats["low:8080"] = &InstanceStats{
		pending:      2,
		successCount: 100,
		failCount:    0,
		latencyEWMA:  10,
	}
	selector.mu.Unlock()

	counts := make(map[string]int)
	for i := 0; i < 100; i++ {
		selected, done, err := selector.Select(context.Background(), instances)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		counts[selected.ID]++
		done(context.Background(), discoverycontract.DoneInfo{Latency: 5 * time.Millisecond})
	}

	if counts["low"] < counts["high"] {
		t.Fatalf("expected low-load instance selected more often, got high=%d low=%d", counts["high"], counts["low"])
	}
}

// TestP2CSelectorForceInstance verifies explicit force-routing still works.
//
// TestP2CSelectorForceInstance 验证显式强制路由仍然生效。
func TestP2CSelectorForceInstance(t *testing.T) {
	selector := NewP2CSelector()
	forced := transportcontract.ServiceInstance{ID: "forced", Address: "forced:8080", Healthy: true}

	selected, done, err := selector.Select(
		context.Background(),
		[]transportcontract.ServiceInstance{{ID: "1", Address: "inst1:8080", Healthy: true}},
		discoverycontract.WithForceInstance(forced),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected.ID != "forced" {
		t.Fatalf("expected forced instance, got %s", selected.ID)
	}

	stats := selector.instanceStats["forced:8080"]
	if stats == nil || stats.pending != 1 {
		t.Fatalf("expected forced instance pending=1, got %+v", stats)
	}

	done(context.Background(), discoverycontract.DoneInfo{})
}

// TestP2CSelectorLatencyEWMAAffectsScore verifies slower instances accumulate worse scores.
//
// TestP2CSelectorLatencyEWMAAffectsScore 验证更慢实例会积累更差的评分。
func TestP2CSelectorLatencyEWMAAffectsScore(t *testing.T) {
	selector := NewP2CSelector()

	selector.mu.Lock()
	selector.instanceStats["slow:8080"] = &InstanceStats{
		pending:        1,
		successCount:   50,
		failCount:      1,
		latencyEWMA:    300,
		latencySamples: 5,
	}
	selector.instanceStats["fast:8080"] = &InstanceStats{
		pending:        1,
		successCount:   50,
		failCount:      1,
		latencyEWMA:    20,
		latencySamples: 5,
	}
	selector.mu.Unlock()

	slowScore := selector.calculateScore(transportcontract.ServiceInstance{Address: "slow:8080"})
	fastScore := selector.calculateScore(transportcontract.ServiceInstance{Address: "fast:8080"})
	if slowScore <= fastScore {
		t.Fatalf("expected slow instance score > fast instance score, got slow=%f fast=%f", slowScore, fastScore)
	}
}
