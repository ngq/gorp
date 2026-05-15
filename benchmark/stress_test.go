// Package benchmark provides stress tests for HTTP/gRPC request chains.
//
// 本包提供 HTTP/gRPC 请求链路的压力测试。
//
// 运行方式：
//
//	go test ./benchmark/... -run=Stress -v -timeout=30m
//
// 目标：
// - 验证高并发下无 OOM、无内存泄漏、无异常抖动
// - 测量吞吐量、延迟分布、资源使用
package benchmark

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	discoverycontract "github.com/ngq/gorp/framework/contract/discovery"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/ngq/gorp/framework/provider/selector/p2c"
	"github.com/ngq/gorp/framework/provider/selector/random"
	"github.com/ngq/gorp/framework/provider/selector/wrr"
)

// ============================================================
// Stress Test Configuration
// 压力测试配置
// ============================================================

const (
	stressDuration    = 10 * time.Second // 每个测试持续时间
	stressConcurrency = 1000             // 并发数
	stressInstanceCount = 100            // 服务实例数
)

// ============================================================
// Selector Stress Tests
// 选择器压力测试
// ============================================================

// StressRandomSelector tests Random selector under high concurrency.
//
// StressRandomSelector 测试 Random 选择器在高并发下的表现。
func StressRandomSelector(t *testing.T) {
	selector := random.NewRandomSelector()
	instances := makeInstances(stressInstanceCount)

	var successCount atomic.Int64
	var errorCount atomic.Int64
	var stop atomic.Bool

	ctx := context.Background()

	startTime := time.Now()

	// Start concurrent workers
	// 启动并发 worker
	var wg sync.WaitGroup
	for i := 0; i < stressConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for !stop.Load() {
				_, _, err := selector.Select(ctx, instances)
				if err != nil {
					errorCount.Add(1)
				} else {
					successCount.Add(1)
				}
			}
		}(i)
	}

	// Wait for duration
	// 等待测试时长
	time.Sleep(stressDuration)
	stop.Store(true)
	wg.Wait()

	duration := time.Since(startTime).Seconds()
	throughput := float64(successCount.Load()) / duration

	t.Logf("Random Selector Stress Test:")
	t.Logf("  Duration: %.2f seconds", duration)
	t.Logf("  Concurrency: %d workers", stressConcurrency)
	t.Logf("  Success: %d operations", successCount.Load())
	t.Logf("  Errors: %d operations", errorCount.Load())
	t.Logf("  Throughput: %.2f ops/sec", throughput)
	t.Logf("  Memory: %d MB", getMemoryMB())

	// Verify no errors occurred
	// 验证没有错误发生
	if errorCount.Load() > 0 {
		t.Errorf("Unexpected errors: %d", errorCount.Load())
	}

	// Verify reasonable throughput
	// 验证合理的吞吐量
	if throughput < 100000 {
		t.Logf("WARNING: Low throughput (< 100k ops/sec), may indicate performance issue")
	}
}

// StressWRRSelector tests WRR selector under high concurrency.
//
// StressWRRSelector 测试 WRR 选择器在高并发下的表现。
func StressWRRSelector(t *testing.T) {
	selector := wrr.NewWRRSelector()
	instances := makeInstancesWithWeight(stressInstanceCount)

	var successCount atomic.Int64
	var errorCount atomic.Int64
	var stop atomic.Bool

	ctx := context.Background()

	startTime := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < stressConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for !stop.Load() {
				_, done, err := selector.Select(ctx, instances)
				if err != nil {
					errorCount.Add(1)
				} else {
					successCount.Add(1)
					done(ctx, discoverycontract.DoneInfo{Err: nil})
				}
			}
		}(i)
	}

	time.Sleep(stressDuration)
	stop.Store(true)
	wg.Wait()

	duration := time.Since(startTime).Seconds()
	throughput := float64(successCount.Load()) / duration

	t.Logf("WRR Selector Stress Test:")
	t.Logf("  Duration: %.2f seconds", duration)
	t.Logf("  Concurrency: %d workers", stressConcurrency)
	t.Logf("  Success: %d operations", successCount.Load())
	t.Logf("  Errors: %d operations", errorCount.Load())
	t.Logf("  Throughput: %.2f ops/sec", throughput)
	t.Logf("  Memory: %d MB", getMemoryMB())

	if errorCount.Load() > 0 {
		t.Errorf("Unexpected errors: %d", errorCount.Load())
	}
}

// StressP2CSelector tests P2C selector under high concurrency.
//
// StressP2CSelector 测试 P2C 选择器在高并发下的表现。
func StressP2CSelector(t *testing.T) {
	selector := p2c.NewP2CSelector()
	instances := makeInstances(stressInstanceCount)

	var successCount atomic.Int64
	var errorCount atomic.Int64
	var stop atomic.Bool

	ctx := context.Background()

	startTime := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < stressConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for !stop.Load() {
				_, done, err := selector.Select(ctx, instances)
				if err != nil {
					errorCount.Add(1)
				} else {
					successCount.Add(1)
					done(ctx, discoverycontract.DoneInfo{Err: nil})
				}
			}
		}(i)
	}

	time.Sleep(stressDuration)
	stop.Store(true)
	wg.Wait()

	duration := time.Since(startTime).Seconds()
	throughput := float64(successCount.Load()) / duration

	t.Logf("P2C Selector Stress Test:")
	t.Logf("  Duration: %.2f seconds", duration)
	t.Logf("  Concurrency: %d workers", stressConcurrency)
	t.Logf("  Success: %d operations", successCount.Load())
	t.Logf("  Errors: %d operations", errorCount.Load())
	t.Logf("  Throughput: %.2f ops/sec", throughput)
	t.Logf("  Memory: %d MB", getMemoryMB())

	if errorCount.Load() > 0 {
		t.Errorf("Unexpected errors: %d", errorCount.Load())
	}
}

// ============================================================
// Metadata Stress Tests
// 元数据压力测试
// ============================================================

// StressMetadata tests metadata operations under high concurrency.
//
// StressMetadata 测试元数据操作在高并发下的表现。
func StressMetadata(t *testing.T) {
	var successCount atomic.Int64
	var stop atomic.Bool

	startTime := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < stressConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for !stop.Load() {
				md := transportcontract.NewMetadata()
				md.Set("x-request-id", fmt.Sprintf("req-%d-%d", workerID, time.Now().UnixNano()))
				md.Set("x-trace-id", fmt.Sprintf("trace-%d", workerID))
				md.Add("x-custom-header", "value1")
				md.Add("x-custom-header", "value2")
				_ = md.Get("x-request-id")
				_ = md.Values("x-custom-header")
				_ = md.Clone()
				successCount.Add(1)
			}
		}(i)
	}

	time.Sleep(stressDuration)
	stop.Store(true)
	wg.Wait()

	duration := time.Since(startTime).Seconds()
	throughput := float64(successCount.Load()) / duration

	t.Logf("Metadata Stress Test:")
	t.Logf("  Duration: %.2f seconds", duration)
	t.Logf("  Concurrency: %d workers", stressConcurrency)
	t.Logf("  Success: %d operations", successCount.Load())
	t.Logf("  Throughput: %.2f ops/sec", throughput)
	t.Logf("  Memory: %d MB", getMemoryMB())
}

// ============================================================
// Memory Leak Detection
// 内存泄漏检测
// ============================================================

// StressMemoryLeak tests for memory leaks under sustained load.
//
// StressMemoryLeak 测试持续负载下的内存泄漏。
func StressMemoryLeak(t *testing.T) {
	selector := random.NewRandomSelector()
	instances := makeInstances(stressInstanceCount)

	// Record initial memory
	// 记录初始内存
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	initialMB := m1.Alloc / 1024 / 1024

	t.Logf("Initial memory: %d MB", initialMB)

	ctx := context.Background()

	// Run sustained load
	// 运行持续负载
	for i := 0; i < 1000000; i++ {
		_, _, _ = selector.Select(ctx, instances)
	}

	// Force GC and check memory
	// 强制 GC 并检查内存
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	finalMB := m2.Alloc / 1024 / 1024

	t.Logf("Final memory: %d MB", finalMB)
	t.Logf("Memory delta: %d MB", finalMB-initialMB)

	// Allow some variance, but significant growth indicates leak
	// 允许一些波动，但显著增长表示泄漏
	if finalMB > initialMB+50 {
		t.Errorf("Potential memory leak: memory grew from %d MB to %d MB", initialMB, finalMB)
	}
}

// ============================================================
// Helper Functions
// 辅助函数
// ============================================================

func getMemoryMB() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc / 1024 / 1024
}

// ============================================================
// Main Stress Test Runner
// 主压力测试运行器
// ============================================================

// TestStressSuite runs all stress tests.
//
// TestStressSuite 运行所有压力测试。
func TestStressSuite(t *testing.T) {
	t.Run("RandomSelector", StressRandomSelector)
	t.Run("WRRSelector", StressWRRSelector)
	t.Run("P2CSelector", StressP2CSelector)
	t.Run("Metadata", StressMetadata)
	t.Run("MemoryLeak", StressMemoryLeak)
}