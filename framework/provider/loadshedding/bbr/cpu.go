// Package bbr 提供 BBR 自适应过载保护实现。
package bbr

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// --- CPU 监控器 ---
//
// 独立线程定期采样 CPU 使用率，使用 EMA（指数移动平均）平滑处理。
// EMA 公式：ema = ema * decay + sample * (1 - decay)
// - decay：衰减系数，越大越平滑
// - 采样间隔：500ms
// - 衰减系数：0.95
//
// 采样来源：Linux 上读取 /proc/stat 获取真实系统 CPU 使用率（见
// cpu_proc.go）；其他平台退化为 goroutine 数代理指标（见 cpu_fallback.go），
// 代理路径不做任何 STW 级别的系统调用。

// cpuMonitor 监控系统 CPU 使用率。
type cpuMonitor struct {
	threshold   float64       // CPU 阈值（0.0-1.0）
	ema         atomic.Value  // 存储 float64，EMA 平滑后的 CPU 使用率
	overloaded_ atomic.Bool   // 当前是否过载
	stopCh      chan struct{} // 停止信号
	stopOnce    sync.Once
	wg          sync.WaitGroup
}

// cpuSample 表示一次 CPU 采样结果。
type cpuSample struct {
	user   uint64
	system uint64
	idle   uint64
	total  uint64
}

// newCPUMonitor 创建 CPU 监控器并启动后台采样线程。
func newCPUMonitor(threshold float64) *cpuMonitor {
	m := &cpuMonitor{
		threshold: threshold,
		stopCh:    make(chan struct{}),
	}
	m.ema.Store(float64(0))

	// 启动后台采样线程
	m.wg.Add(1)
	go m.run()

	return m
}

// Stop 停止 CPU 监控。幂等，可安全多次调用。
func (m *cpuMonitor) Stop() {
	m.stopOnce.Do(func() { close(m.stopCh) })
	m.wg.Wait()
}

// run 是后台采样线程的主循环。
func (m *cpuMonitor) run() {
	defer m.wg.Done()

	// 采样间隔
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	// 获取初始采样
	lastSample, lastOK := m.sampleCPU()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			// 获取新采样
			currentSample, currentOK := m.sampleCPU()

			// 计算 CPU 使用率
			usage := m.calculateUsage(lastSample, currentSample, lastOK && currentOK)
			lastSample, lastOK = currentSample, currentOK

			// EMA 平滑
			oldEMA := m.ema.Load().(float64)
			newEMA := oldEMA*0.95 + usage*(1-0.95)
			m.ema.Store(newEMA)

			// 更新过载状态
			m.overloaded_.Store(newEMA >= m.threshold)
		}
	}
}

// sampleCPU 采样当前系统 CPU 使用情况。
// 第二个返回值表示采样是否来自真实系统计数器；false 时调用方应使用代理指标。
func (m *cpuMonitor) sampleCPU() (cpuSample, bool) {
	return sampleSystemCPU()
}

// calculateUsage 计算两次采样之间的 CPU 使用率。
// real 为 true 时基于系统计数器计算：usage = 1 - Δidle/Δtotal。
// real 为 false 时（非 Linux 平台回退）使用 goroutine 数代理指标。
func (m *cpuMonitor) calculateUsage(last, current cpuSample, real bool) float64 {
	if real {
		dTotal := current.total - last.total
		if dTotal == 0 {
			return 0
		}
		dIdle := current.idle - last.idle
		usage := 1 - float64(dIdle)/float64(dTotal)
		if usage < 0 {
			usage = 0
		}
		if usage > 1 {
			usage = 1
		}
		return usage
	}

	// 回退代理：goroutine 数 / (GOMAXPROCS * 100)。
	// 这不是 CPU 使用率——只是调度压力的粗略代理，仅在无法读取
	// 系统计数器的平台上避免 BBR 完全失效。
	goroutineCount := runtime.NumGoroutine()
	maxProcs := runtime.GOMAXPROCS(0)
	threshold := float64(maxProcs * 100)
	usage := float64(goroutineCount) / threshold
	if usage > 1.0 {
		usage = 1.0
	}
	return usage
}

// overloaded 返回当前是否 CPU 过载。
func (m *cpuMonitor) overloaded() bool {
	return m.overloaded_.Load()
}

// cpuUsage 返回当前 EMA 平滑后的 CPU 使用率。
func (m *cpuMonitor) cpuUsage() float64 {
	return m.ema.Load().(float64)
}
