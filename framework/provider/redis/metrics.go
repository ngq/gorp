// Package redis provides Redis service for gorp framework.
// This file defines Prometheus metrics hook for Redis client monitoring.
// Includes command count, duration, and connection count metrics.
//
// Redis 包提供 Redis 服务，用于 gorp 框架。
// 本文件定义用于 Redis 客户端监控的 Prometheus 指标 hook。
// 包括命令次数、耗时和连接数指标。
package redis

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
)

var (
	// redisCommandsTotal Redis 命令执行总数
	redisCommandsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gorp_redis_commands_total",
		Help: "Total number of Redis commands executed.",
	}, []string{"command", "status"})

	// redisCommandDuration Redis 命令执行耗时
	redisCommandDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gorp_redis_command_duration_seconds",
		Help:    "Redis command latency in seconds.",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"command", "status"})

	// redisPoolHitsTotal 连接池命中累计
	redisPoolHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gorp_redis_pool_hits_total",
		Help: "The number of times a connection was found in the pool.",
	}, []string{"addr"})

	// redisPoolMissesTotal 连接池未命中累计
	redisPoolMissesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gorp_redis_pool_misses_total",
		Help: "The number of times a connection was not found in the pool.",
	}, []string{"addr"})

	// redisPoolTimeoutsTotal 获取连接超时累计
	redisPoolTimeoutsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gorp_redis_pool_timeouts_total",
		Help: "The number of times a wait for a connection timed out.",
	}, []string{"addr"})

	// redisPoolConns 当前连接总数
	redisPoolConns = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gorp_redis_pool_connections",
		Help: "The current number of connections in the pool.",
	}, []string{"addr"})

	// redisPoolIdleConns 当前空闲连接数
	redisPoolIdleConns = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gorp_redis_pool_idle_connections",
		Help: "The current number of idle connections in the pool.",
	}, []string{"addr"})

	// redisPoolStaleConns 当前失效连接数
	redisPoolStaleConns = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gorp_redis_pool_stale_connections",
		Help: "The current number of stale connections in the pool.",
	}, []string{"addr"})
)

// RedisPoolCollector 定期从 client.PoolStats() 采集连接池指标。
// go-redis 的 Dial/Close 没有成对的连接生命周期钩子，在 Hook 层统计连接数
// 只会得到只增不减的失真 Gauge；PoolStats 是官方给出的连接池状态来源。
type RedisPoolCollector struct {
	client *redis.Client
	addr   string
	stopCh chan struct{}
	once   sync.Once
	// 上次采集的累计值；Hits/Misses/Timeouts 是单调累计，Counter 只加增量。
	lastHits    uint32
	lastMisses  uint32
	lastTimeouts uint32
}

// NewRedisPoolCollector 创建连接池指标采集器。
func NewRedisPoolCollector(client *redis.Client, addr string) *RedisPoolCollector {
	return &RedisPoolCollector{
		client: client,
		addr:   addr,
		stopCh: make(chan struct{}),
	}
}

// Start 启动后台采集 goroutine（每 5 秒一次）。
func (c *RedisPoolCollector) Start() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-c.stopCh:
				return
			case <-ticker.C:
				c.collect()
			}
		}
	}()
}

// Stop 停止采集 goroutine。幂等。
func (c *RedisPoolCollector) Stop() error {
	c.once.Do(func() { close(c.stopCh) })
	return nil
}

// Close 实现 io.Closer，便于 RegisterCloser 接线。
func (c *RedisPoolCollector) Close() error { return c.Stop() }

func (c *RedisPoolCollector) collect() {
	if c.client == nil {
		return
	}
	stats := c.client.PoolStats()
	redisPoolConns.WithLabelValues(c.addr).Set(float64(stats.TotalConns))
	redisPoolIdleConns.WithLabelValues(c.addr).Set(float64(stats.IdleConns))
	redisPoolStaleConns.WithLabelValues(c.addr).Set(float64(stats.StaleConns))

	// 累计值只加增量。
	hits := uint32(stats.Hits)
	if hits >= c.lastHits {
		redisPoolHitsTotal.WithLabelValues(c.addr).Add(float64(hits - c.lastHits))
	}
	misses := uint32(stats.Misses)
	if misses >= c.lastMisses {
		redisPoolMissesTotal.WithLabelValues(c.addr).Add(float64(misses - c.lastMisses))
	}
	timeouts := uint32(stats.Timeouts)
	if timeouts >= c.lastTimeouts {
		redisPoolTimeoutsTotal.WithLabelValues(c.addr).Add(float64(timeouts - c.lastTimeouts))
	}
	c.lastHits = hits
	c.lastMisses = misses
	c.lastTimeouts = timeouts
}

// RedisMetricsHook 为 Redis 客户端添加指标收集 hook。
//
// 中文说明：
// - 通过 go-redis 的 Hook 接口拦截每次命令执行；
// - 记录命令执行次数和耗时；
// - 区分成功和失败的命令（status 标签）；
// - 用于监控 Redis 性能和识别慢命令。
type RedisMetricsHook struct{}

// NewRedisMetricsHook 创建 Redis 指标收集 hook。
func NewRedisMetricsHook() *RedisMetricsHook {
	return &RedisMetricsHook{}
}

// DialHook 实现 redis.Hook 接口，透传连接建立。
// go-redis 的 Hook 没有连接关闭回调，在 Dial 成功时 Inc 的 Gauge 无人 Dec，
// 只增不减与真实连接数脱节——曾因此误报连接泄漏告警，故不再在此统计连接数；
// 需要连接池指标时请从 client.PoolStats() 采集。
func (h *RedisMetricsHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

// ProcessHook 实现 redis.Hook 接口，在命令执行前后记录指标。
func (h *RedisMetricsHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)

		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}

		command := cmd.Name()
		redisCommandsTotal.WithLabelValues(command, status).Inc()
		redisCommandDuration.WithLabelValues(command, status).Observe(duration)

		return err
	}
}

// ProcessPipelineHook 实现 redis.Hook 接口，处理 pipeline 命令。
func (h *RedisMetricsHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)

		duration := time.Since(start).Seconds()
		status := "success"
		if err != nil {
			status = "error"
		}

		// Pipeline 命令统一记录为 "pipeline"
		redisCommandsTotal.WithLabelValues("pipeline", status).Inc()
		redisCommandDuration.WithLabelValues("pipeline", status).Observe(duration)

		return err
	}
}
