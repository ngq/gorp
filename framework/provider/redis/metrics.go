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
)

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
