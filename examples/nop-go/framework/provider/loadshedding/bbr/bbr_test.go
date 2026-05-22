// Package bbr 提供 BBR 自适应过载保护实现。
package bbr

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.Equal(t, true, cfg.Enabled)
	require.Equal(t, 0.8, cfg.CPUThreshold)
	require.Equal(t, 10*time.Second, cfg.WindowSize)
	require.Equal(t, 100, cfg.BucketCount)
	require.Equal(t, 1*time.Second, cfg.CoolDown)
	require.Equal(t, 1*time.Millisecond, cfg.MinRTThreshold)
}

func TestNewLoadShedder(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		ls := NewLoadShedder(nil)
		require.NotNil(t, ls)
		require.NotNil(t, ls.cfg)
		require.NotNil(t, ls.cpuMonitor)
		require.NotNil(t, ls.window)
	})

	t.Run("custom config", func(t *testing.T) {
		cfg := &Config{
			Enabled:        true,
			CPUThreshold:   0.9,
			WindowSize:     5 * time.Second,
			BucketCount:    50,
			CoolDown:       2 * time.Second,
			MinRTThreshold: 2 * time.Millisecond,
		}
		ls := NewLoadShedder(cfg)
		require.NotNil(t, ls)
		require.Equal(t, 0.9, ls.cfg.CPUThreshold)
		require.Equal(t, 5*time.Second, ls.cfg.WindowSize)
	})
}

func TestLoadShedder_Allow_Done(t *testing.T) {
	ls := NewLoadShedder(&Config{
		Enabled:        true,
		CPUThreshold:   0.8,
		WindowSize:     1 * time.Second,
		BucketCount:    10,
		CoolDown:       100 * time.Millisecond,
		MinRTThreshold: 1 * time.Millisecond,
	})
	defer ls.cpuMonitor.Stop()

	ctx := context.Background()

	t.Run("allow when cpu not overloaded", func(t *testing.T) {
		err := ls.Allow(ctx, "test-resource")
		require.NoError(t, err)

		// 释放
		ls.Done(ctx, "test-resource", nil)
	})
}

func TestLoadShedder_ConcurrentAccess(t *testing.T) {
	ls := NewLoadShedder(&Config{
		Enabled:        true,
		CPUThreshold:   0.8,
		WindowSize:     1 * time.Second,
		BucketCount:    10,
		CoolDown:       100 * time.Millisecond,
		MinRTThreshold: 1 * time.Millisecond,
	})
	defer ls.cpuMonitor.Stop()

	const n = 100
	var allowed, denied atomic.Int32
	var wg sync.WaitGroup

	ctx := context.Background()

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := ls.Allow(ctx, "test-resource")
			if err == nil {
				allowed.Add(1)
				// 模拟处理
				time.Sleep(10 * time.Millisecond)
				ls.Done(ctx, "test-resource", nil)
			} else {
				denied.Add(1)
			}
		}()
	}

	wg.Wait()

	// 应该有请求被允许
	require.Greater(t, allowed.Load(), int32(0))
}

func TestSlidingWindow(t *testing.T) {
	w := newSlidingWindow(1*time.Second, 10)
	require.NotNil(t, w)

	t.Run("record and max pass", func(t *testing.T) {
		// 记录一些请求
		w.Record("test", 10*time.Millisecond)
		w.Record("test", 20*time.Millisecond)
		w.Record("test", 15*time.Millisecond)

		maxPass := w.MaxPass()
		require.Greater(t, maxPass, int64(0))
	})

	t.Run("min rt", func(t *testing.T) {
		w.Record("test", 5*time.Millisecond)
		w.Record("test", 50*time.Millisecond)

		minRT := w.MinRT()
		require.Greater(t, minRT, time.Duration(0))
	})
}

func TestCPUMonitor(t *testing.T) {
	m := newCPUMonitor(0.8)
	require.NotNil(t, m)

	// 等待采样
	time.Sleep(600 * time.Millisecond)

	// 检查 CPU 使用率
	usage := m.cpuUsage()
	require.GreaterOrEqual(t, usage, float64(0))
	require.LessOrEqual(t, usage, float64(1))

	// 停止监控
	m.Stop()
}

func TestWithStartTime(t *testing.T) {
	ctx := context.Background()
	ctx = WithStartTime(ctx)

	startTime, ok := ctx.Value(startTimeKey).(time.Time)
	require.True(t, ok)
	require.False(t, startTime.IsZero())
}

func TestCalculateMaxInFlight(t *testing.T) {
	ls := NewLoadShedder(&Config{
		Enabled:        true,
		CPUThreshold:   0.8,
		WindowSize:     10 * time.Second,
		BucketCount:    100,
		CoolDown:       1 * time.Second,
		MinRTThreshold: 1 * time.Millisecond,
	})
	defer ls.cpuMonitor.Stop()

	// 记录一些数据
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		ls.Allow(ctx, "test")
		ls.window.Record("test", 10*time.Millisecond)
		ls.Done(ctx, "test", nil)
	}

	// 计算 maxInFlight
	maxInFlight := ls.calculateMaxInFlight()
	require.Greater(t, maxInFlight, int64(0))
}

func TestResourceIsolation(t *testing.T) {
	ls := NewLoadShedder(&Config{
		Enabled:        true,
		CPUThreshold:   0.8,
		WindowSize:     1 * time.Second,
		BucketCount:    10,
		CoolDown:       100 * time.Millisecond,
		MinRTThreshold: 1 * time.Millisecond,
	})
	defer ls.cpuMonitor.Stop()

	ctx := context.Background()

	// 不同资源应该有独立的统计
	err1 := ls.Allow(ctx, "resource-1")
	err2 := ls.Allow(ctx, "resource-2")

	require.NoError(t, err1)
	require.NoError(t, err2)

	ls.Done(ctx, "resource-1", nil)
	ls.Done(ctx, "resource-2", nil)
}
