package cron

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// cronJobsTotal cron 任务执行总数
	cronJobsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gorp_cron_jobs_total",
		Help: "Total number of cron job executions.",
	}, []string{"job", "status"})

	// cronJobDuration cron 任务执行耗时
	cronJobDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gorp_cron_job_duration_seconds",
		Help:    "Cron job execution duration in seconds.",
		Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 120, 300},
	}, []string{"job"})

	// cronJobsInProgress 当前正在执行的 cron 任务数
	cronJobsInProgress = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gorp_cron_jobs_in_progress",
		Help: "Current number of cron jobs being executed.",
	}, []string{"job"})

	// cronJobsScheduled 已注册的 cron 任务数
	cronJobsScheduled = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gorp_cron_jobs_scheduled",
		Help: "Number of cron jobs currently scheduled.",
	})
)

// CronJobMetrics 包装 cron 任务并收集指标。
//
// 中文说明：
// - jobName: 任务名称，用于标识不同的任务；
// - 记录任务执行次数、耗时、当前执行数；
// - 区分成功和失败的任务（status 标签）。
type CronJobMetrics struct {
	JobName string
}