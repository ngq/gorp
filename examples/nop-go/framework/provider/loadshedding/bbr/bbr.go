// Package bbr 提供 BBR 自适应过载保护实现。
//
// BBR（Bottleneck Bandwidth and RTT）是一种基于系统负载的自适应限流算法。
// 核心思想：当 CPU 过载时，根据历史统计动态调整最大并发数。
//
// 公式：maxInFlight = maxPASS * minRT * bucketPerSecond / 1000
// - maxPASS：滑动窗口内单个桶的最大通过请求数
// - minRT：滑动窗口内的最小响应时间
// - bucketPerSecond：每秒的桶数量
//
// 判断逻辑：
// - CPU < 阈值：允许请求通过
// - CPU >= 阈值 且 inflight > maxInFlight：拒绝请求
// - 冷却期：CPU 下降后仍有 1 秒冷却时间
package bbr

import (
	"context"
	"sync"
	"time"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// ErrLoadShedded 表示请求因过载保护被丢弃。
var ErrLoadShedded = resiliencecontract.ServiceUnavailable("server is busy: bbr load shedding active")

// --- BBR 自适应过载保护实现 ---
//
// BBR（Bottleneck Bandwidth and RTT）是一种基于系统负载的自适应限流算法。
// 核心思想：当 CPU 过载时，根据历史统计动态调整最大并发数。
//
// 公式：maxInFlight = maxPASS * minRT * bucketPerSecond / 1000
// - maxPASS：滑动窗口内单个桶的最大通过请求数
// - minRT：滑动窗口内的最小响应时间
// - bucketPerSecond：每秒的桶数量
//
// 判断逻辑：
// - CPU < 阈值：允许请求通过
// - CPU >= 阈值 且 inflight > maxInFlight：拒绝请求
// - 冷却期：CPU 下降后仍有 1 秒冷却时间

// LoadShedder 是 BBR 自适应过载保护实现，实现 resiliencecontract.LoadShedder 契约。
type LoadShedder struct {
	cfg        *Config
	cpuMonitor *cpuMonitor
	window     *slidingWindow
	stats      sync.Map // resource -> *resourceStats
}

// resourceStats 记录单个资源的运行时状态。
type resourceStats struct {
	mu          sync.Mutex
	inFlight    int64     // 当前正在处理的请求数
	droppedTime time.Time // 最近一次拒绝的时间（用于冷却判断）
	passCount   int64     // 通过请求数累计
	rtSum       int64     // 响应时间累计（纳秒）
	rtCount     int64     // 响应时间采样次数
}

// Config 是 BBR 策略的配置参数。
type Config struct {
	Enabled        bool          // 是否启用
	CPUThreshold   float64       // CPU 阈值（0.0-1.0），默认 0.8
	WindowSize     time.Duration // 滑动窗口大小，默认 10s
	BucketCount    int           // 桶数量，默认 100
	CoolDown       time.Duration // 冷却时间，默认 1s
	MinRTThreshold time.Duration // 最小 RT 阈值，低于此值不限流，默认 1ms
}

// DefaultConfig 返回默认 BBR 配置。
func DefaultConfig() *Config {
	return &Config{
		Enabled:        true,
		CPUThreshold:   0.8,
		WindowSize:     10 * time.Second,
		BucketCount:    100,
		CoolDown:       1 * time.Second,
		MinRTThreshold: 1 * time.Millisecond,
	}
}

// NewLoadShedder 创建 BBR 自适应过载保护器。
func NewLoadShedder(cfg *Config) *LoadShedder {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &LoadShedder{
		cfg:        cfg,
		cpuMonitor: newCPUMonitor(cfg.CPUThreshold),
		window:     newSlidingWindow(cfg.WindowSize, cfg.BucketCount),
	}
}

// Allow 判断是否允许请求通过。
//
// 判断逻辑：
// 1. CPU 未过载 → 允许
// 2. CPU 过载但处于冷却期 → 允许
// 3. CPU 过载且 inflight <= maxInFlight → 允许
// 4. CPU 过载且 inflight > maxInFlight → 拒绝
func (b *LoadShedder) Allow(ctx context.Context, resource string) error {
	stats := b.getOrCreateStats(resource)

	// 快速路径：CPU 未过载
	if !b.cpuMonitor.overloaded() {
		stats.mu.Lock()
		stats.inFlight++
		stats.mu.Unlock()
		return nil
	}

	// CPU 过载，检查是否在冷却期
	stats.mu.Lock()
	if time.Since(stats.droppedTime) < b.cfg.CoolDown {
		// 冷却期内，允许请求（让系统恢复）
		stats.inFlight++
		stats.mu.Unlock()
		return nil
	}

	// 计算最大允许的 inflight
	maxInFlight := b.calculateMaxInFlight()

	// 判断是否超过上限
	if stats.inFlight < maxInFlight {
		stats.inFlight++
		stats.mu.Unlock()
		return nil
	}

	// 拒绝请求
	stats.droppedTime = time.Now()
	stats.mu.Unlock()
	return ErrLoadShedded
}

// Done 完成请求处理，更新统计信息。
func (b *LoadShedder) Done(ctx context.Context, resource string, err error) {
	stats := b.getOrCreateStats(resource)

	// 记录响应时间（从 context 获取开始时间，或使用当前时间估算）
	startTime, ok := ctx.Value(startTimeKey).(time.Time)
	var rt time.Duration
	if ok {
		rt = time.Since(startTime)
	} else {
		rt = 0 // 无法获取开始时间，不记录
	}

	stats.mu.Lock()
	stats.inFlight--
	if rt > 0 {
		stats.rtSum += int64(rt)
		stats.rtCount++
	}
	stats.passCount++
	stats.mu.Unlock()

	// 更新滑动窗口统计
	b.window.Record(resource, rt)
}

// calculateMaxInFlight 计算当前最大允许的 inflight 数。
//
// 公式：maxInFlight = maxPASS * minRT * bucketPerSecond / 1000
func (b *LoadShedder) calculateMaxInFlight() int64 {
	maxPASS := b.window.MaxPass()
	minRT := b.window.MinRT()

	// minRT 太小（< 1ms）说明系统极快，不需要限流
	if minRT < b.cfg.MinRTThreshold {
		return int64(1 << 30) // 返回极大值，相当于不限流
	}

	// 每秒桶数量
	bucketPerSecond := float64(b.cfg.BucketCount) / b.cfg.WindowSize.Seconds()

	// 计算 maxInFlight
	maxInFlight := float64(maxPASS) * minRT.Seconds() * bucketPerSecond

	// 至少允许 1 个请求
	if maxInFlight < 1 {
		return 1
	}

	return int64(maxInFlight)
}

// getOrCreateStats 获取或创建资源对应的统计状态。
func (b *LoadShedder) getOrCreateStats(resource string) *resourceStats {
	if v, ok := b.stats.Load(resource); ok {
		return v.(*resourceStats)
	}
	stats := &resourceStats{}
	actual, _ := b.stats.LoadOrStore(resource, stats)
	return actual.(*resourceStats)
}

// --- Context Key ---
//
// 用于在 context 中传递请求开始时间，以便精确计算响应时间。

// startTimeKey 是 context 中存储请求开始时间的 key。
type startTimeKeyType struct{}

var startTimeKey = startTimeKeyType{}

// WithStartTime 在 context 中设置请求开始时间。
// 用于 BBR 精确计算响应时间。
func WithStartTime(ctx context.Context) context.Context {
	return context.WithValue(ctx, startTimeKey, time.Now())
}