// Package bbr 提供 BBR 自适应过载保护实现。
package bbr

import (
	"sync"
	"time"
)

// --- 滑动窗口统计 ---
//
// 使用环形桶结构记录请求统计信息。
// 每个桶记录一个时间窗口内的：
// - 通过请求数（passCount）
// - 响应时间总和（rtSum）
// - 响应时间最小值（rtMin）
//
// 滑动窗口遍历所有桶，计算：
// - maxPASS：所有桶中最大的通过请求数
// - minRT：所有桶中最小的响应时间

// slidingWindow 是滑动窗口统计器。
type slidingWindow struct {
	mu          sync.RWMutex
	buckets     []*bucket     // 环形桶数组
	bucketCount int           // 桶数量
	bucketSize  time.Duration // 每个桶的时间窗口大小
	lastTime    time.Time     // 最近一次更新时间
	currentIdx  int           // 当前桶索引
}

// bucket 是单个时间窗口内的统计桶。
type bucket struct {
	mu         sync.Mutex
	passCount_ int64     // 通过请求数
	rtSum      int64     // 响应时间总和（纳秒）
	rtMin      int64     // 响应时间最小值（纳秒）
	rtCount    int64     // 响应时间采样次数
	resetTime  time.Time // 桶重置时间
}

// newSlidingWindow 创建滑动窗口统计器。
//
// 参数：
// - windowSize：滑动窗口总大小（如 10s）
// - bucketCount：桶数量（如 100）
//
// 每个桶的时间窗口 = windowSize / bucketCount
func newSlidingWindow(windowSize time.Duration, bucketCount int) *slidingWindow {
	bucketSize := windowSize / time.Duration(bucketCount)

	buckets := make([]*bucket, bucketCount)
	for i := range buckets {
		buckets[i] = &bucket{
			rtMin: int64(1<<63 - 1), // 初始化为最大值
		}
	}

	return &slidingWindow{
		buckets:     buckets,
		bucketCount: bucketCount,
		bucketSize:  bucketSize,
		lastTime:    time.Now(),
		currentIdx:  0,
	}
}

// Record 记录一次请求的响应时间。
func (w *slidingWindow) Record(resource string, rt time.Duration) {
	w.mu.Lock()

	// 检查是否需要前进到下一个桶
	now := time.Now()
	elapsed := now.Sub(w.lastTime)

	// 计算需要前进的桶数量
	advance := int(elapsed / w.bucketSize)
	if advance > 0 {
		// 前进桶，重置过期桶
		for i := 0; i < advance && i < w.bucketCount; i++ {
			w.currentIdx = (w.currentIdx + 1) % w.bucketCount
			w.buckets[w.currentIdx].reset()
		}
		w.lastTime = now
	}

	w.mu.Unlock()

	// 在当前桶中记录统计
	currentBucket := w.buckets[w.currentIdx]
	currentBucket.record(rt)
}

// MaxPass 返回滑动窗口内所有桶的最大通过请求数。
//
// 用于计算 maxInFlight = maxPASS * minRT * bucketPerSecond / 1000
func (w *slidingWindow) MaxPass() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var maxPass int64
	for _, b := range w.buckets {
		pass := b.passCount()
		if pass > maxPass {
			maxPass = pass
		}
	}

	// 如果没有数据，返回默认值 1
	if maxPass == 0 {
		return 1
	}

	return maxPass
}

// MinRT 返回滑动窗口内所有桶的最小响应时间。
//
// 用于计算 maxInFlight = maxPASS * minRT * bucketPerSecond / 1000
func (w *slidingWindow) MinRT() time.Duration {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var minRT int64 = int64(1<<63 - 1)
	var hasData bool

	for _, b := range w.buckets {
		rt := b.minRT()
		if rt > 0 && rt < minRT {
			minRT = rt
			hasData = true
		}
	}

	// 如果没有数据，返回默认值 1s
	if !hasData {
		return time.Second
	}

	return time.Duration(minRT)
}

// AverageRT 返回滑动窗口内的平均响应时间。
func (w *slidingWindow) AverageRT() time.Duration {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var rtSum int64
	var rtCount int64

	for _, b := range w.buckets {
		sum, count := b.rtStats()
		rtSum += sum
		rtCount += count
	}

	if rtCount == 0 {
		return 0
	}

	return time.Duration(rtSum / rtCount)
}

// TotalPass 返回滑动窗口内的总通过请求数。
func (w *slidingWindow) TotalPass() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var total int64
	for _, b := range w.buckets {
		total += b.passCount()
	}
	return total
}

// --- bucket 方法 ---
//
// bucket 是单个时间窗口内的统计容器。

// reset 重置桶的统计数据。
func (b *bucket) reset() {
	b.mu.Lock()
	b.passCount_ = 0
	b.rtSum = 0
	b.rtMin = int64(1<<63 - 1)
	b.rtCount = 0
	b.resetTime = time.Now()
	b.mu.Unlock()
}

// record 记录一次请求的响应时间。
func (b *bucket) record(rt time.Duration) {
	b.mu.Lock()
	b.passCount_++
	if rt > 0 {
		rtNs := int64(rt)
		b.rtSum += rtNs
		b.rtCount++
		if rtNs < b.rtMin {
			b.rtMin = rtNs
		}
	}
	b.mu.Unlock()
}

// passCount 返回桶的通过请求数。
func (b *bucket) passCount() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.passCount_
}

// minRT 返回桶的最小响应时间（纳秒）。
func (b *bucket) minRT() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.rtCount == 0 {
		return 0
	}
	return b.rtMin
}

// rtStats 返回桶的响应时间总和和采样次数。
func (b *bucket) rtStats() (int64, int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.rtSum, b.rtCount
}