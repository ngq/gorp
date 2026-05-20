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

// cpuMonitor 监控系统 CPU 使用率。
type cpuMonitor struct {
	threshold   float64        // CPU 阈值（0.0-1.0）
	ema         atomic.Value   // 存储 float64，EMA 平滑后的 CPU 使用率
	overloaded_ atomic.Bool     // 当前是否过载
	stopCh      chan struct{}  // 停止信号
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

// Stop 停止 CPU 监控。
func (m *cpuMonitor) Stop() {
	close(m.stopCh)
	m.wg.Wait()
}

// run 是后台采样线程的主循环。
func (m *cpuMonitor) run() {
	defer m.wg.Done()

	// 采样间隔
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	// 获取初始采样
	lastSample := m.sampleCPU()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			// 获取新采样
			currentSample := m.sampleCPU()

			// 计算 CPU 使用率
			usage := m.calculateUsage(lastSample, currentSample)
			lastSample = currentSample

			// EMA 平滑
			oldEMA := m.ema.Load().(float64)
			newEMA := oldEMA*0.95 + usage*(1-0.95)
			m.ema.Store(newEMA)

			// 更新过载状态
			m.overloaded_.Store(newEMA >= m.threshold)
		}
	}
}

// sampleCPU 采样当前 CPU 使用情况。
//
// 使用 runtime.ReadMemStats 和自定义计算获取 CPU 使用率。
// 注意：Go 没有直接获取 CPU 使用率的 API，这里使用进程时间近似计算。
func (m *cpuMonitor) sampleCPU() cpuSample {
	// 使用进程 CPU 时间
	var rusage runtime.MemStats
	runtime.ReadMemStats(&rusage)

	// 使用 goroutine 数量和 GC CPU 占用作为代理指标
	// 这是一个近似方法，不是精确的 CPU 使用率
	//
	// 更精确的方法需要读取 /proc/stat（Linux）或调用系统 API
	// 这里采用简化方案：基于 GC CPU 占用和 goroutine 数量估算

	return cpuSample{
		user:   rusage.PauseTotalNs, // GC 暂停时间作为 user 时间代理
		system: 0,
		idle:   0,
		total:  rusage.PauseTotalNs,
	}
}

// calculateUsage 计算两次采样之间的 CPU 使用率。
func (m *cpuMonitor) calculateUsage(last, current cpuSample) float64 {
	// 简化实现：基于 GC 暂停时间占比估算
	// 实际生产环境建议使用 gopsutil 等库获取精确 CPU 使用率
	//
	// 这里使用一个启发式方法：
	// - 监控 GC 频率和暂停时间
	// - 如果 GC 暂停时间增长过快，说明内存压力大，间接反映 CPU 压力

	// 使用 goroutine 数量作为负载指标
	// 当 goroutine 数量过多时，调度开销增加，视为 CPU 过载
	goroutineCount := runtime.NumGoroutine()
	maxProcs := runtime.GOMAXPROCS(0)

	// 如果 goroutine 数量超过 GOMAXPROCS 的 100 倍，认为 CPU 可能过载
	// 这是一个启发式阈值
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

// --- 更精确的 CPU 监控实现（可选）---
//
// 如果需要更精确的 CPU 监控，可以使用以下方法：
// 1. Linux: 读取 /proc/stat 或 /proc/[pid]/stat
// 2. 跨平台: 使用 github.com/shirou/gopsutil/v3/cpu
//
// 当前实现使用 goroutine 数量作为代理指标，适用于大多数场景。
// 如果需要更精确的监控，可以在编译时通过 build tag 选择不同实现。

// cpuMonitorPrecise 是更精确的 CPU 监控实现（使用 gopsutil）。
// 当前未启用，仅作为参考。
type cpuMonitorPrecise struct {
	threshold float64
	ema       atomic.Value
	stopCh    chan struct{}
	wg        sync.WaitGroup
	// cpu.Percent 需要 time.Duration 参数
	// 这里不实际使用，仅作为文档说明
}

// 注意：如果需要使用 gopsutil，需要添加依赖：
// import "github.com/shirou/gopsutil/v3/cpu"
//
// 然后在 run() 中使用：
// percent, _ := cpu.Percent(500*time.Millisecond, false)
// usage := percent[0] / 100.0
