// Package cron provides cron scheduling service for gorp framework.
// Implements framework-level wrapper around robfig/cron with timeout control and Prometheus metrics.
// Supports second-level cron expressions with automatic panic recovery.
//
// Cron 调度服务包，提供 gorp 框架的定时任务调度能力。
// 对 robfig/cron 做 framework 级封装，包含超时控制和 Prometheus 指标采集。
// 支持秒级 cron 表达式，自动 panic 恢复。
//
// Eg:
//
//	// 注册 Provider
//	app.Register(cron.NewProvider())
//
//	// 添加定时任务
//	cronSvc := c.MustMake(runtimecontract.CronKey).(runtimecontract.Cron)
//	cronSvc.AddNamed("cleanup", "0 0 3 * * *", func(ctx context.Context) error {
//	    return cleanupOldData(ctx)
//	})
//	cronSvc.Start()
package cron

import (
	"context"
	"fmt"
	"sync"
	"time"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"

	"github.com/robfig/cron/v3"
)

// Provider registers the cron service contract.
// Exposes runtimecontract.CronKey for unified task registration.
// Core logic: Create Service instance with second-level parser, bind to container.
//
// Provider 注册 Cron 服务契约。
// 统一暴露 runtimecontract.CronKey，用于定时任务注册。
// 核心逻辑：创建支持秒级表达式的 Service 实例、绑定到容器。
type Provider struct{}

// NewProvider creates a new cron provider instance.
//
// NewProvider 创建新的 Cron Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "cron".
//
// Name 返回 Provider 名称 "cron"。
func (p *Provider) Name() string { return "cron" }

// IsDefer returns false, cron should be initialized immediately for service registration.
//
// IsDefer 返回 false，Cron 应立即初始化以便任务注册。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the cron service contract key.
//
// Provides 返回 Cron 服务契约键。
func (p *Provider) Provides() []string { return []string{runtimecontract.CronKey} }

// DependsOn returns the keys this provider depends on.
// Cron provider has no dependencies.
//
// DependsOn 返回该 provider 依赖的 key。
// Cron provider 无依赖。
func (p *Provider) DependsOn() []string { return nil }

// Register binds the cron service factory to the container.
// Core logic: Create Service with second-level parser, bind factory.
//
// Register 将 Cron 服务工厂绑定到容器。
// 核心逻辑：创建支持秒级表达式的 Service、绑定工厂。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(runtimecontract.CronKey, func(runtimecontract.Container) (any, error) {
		return NewService(), nil
	}, true)
	return nil
}

// Boot is a no-op for cron provider.
//
// Boot Cron Provider 无启动逻辑（启动由 Start 调用触发）。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// jobRecord tracks the execution history of a single cron job.
//
// jobRecord 追踪单个 cron 任务的执行历史。
type jobRecord struct {
	name    string
	spec    string
	status  runtimecontract.CronJobStatus
	lastRun time.Time
	nextRun time.Time
	err     string
	took    time.Duration
}

// Service wraps robfig/cron with framework-level enhancements.
// Features: timeout control, Prometheus metrics, panic recovery, job introspection.
// Core logic: Encapsulate robfig/cron, add timeout and metrics hooks, track execution history.
//
// Service 对 robfig/cron 做 framework 级封装。
// 特性：超时控制、Prometheus 指标采集、panic 恢复、任务内省。
// 核心逻辑：封装 robfig/cron，添加超时和指标 hooks，追踪执行历史。
type Service struct {
	c     *cron.Cron         // c is the underlying cron scheduler.
	mu    sync.RWMutex       // mu protects jobs and specs.
	jobs  map[int]*jobRecord // jobs maps entryID to job execution records.
	specs map[int]string     // specs maps entryID to the original cron spec expression.
}

// NewService creates a cron service with second-level expression support.
// Includes automatic panic recovery chain.
//
// NewService 创建支持秒级表达式的 Cron 调度器。
// 包含自动 panic 恢复链。
func NewService() *Service {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	c := cron.New(
		cron.WithParser(parser),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)
	return &Service{
		c:     c,
		jobs:  make(map[int]*jobRecord),
		specs: make(map[int]string),
	}
}

// Add registers a cron job without Prometheus metrics labels.
// Suitable for simple scheduled tasks.
// Core logic: Wrap function with 5-minute timeout, add to scheduler, track execution history.
//
// Add 注册不带指标标签的定时任务。
// 适用于简单定时任务场景。
// 核心逻辑：用 5 分钟超时包装函数、添加到调度器、追踪执行历史。
func (s *Service) Add(spec string, fn func(ctx context.Context) error) (int, error) {
	var id cron.EntryID
	var err error
	id, err = s.c.AddFunc(spec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		start := time.Now()
		runErr := fn(ctx)
		took := time.Since(start)
		s.recordExecution(int(id), runErr, start, took)
	})
	if err != nil {
		return int(id), err
	}
	s.mu.Lock()
	s.jobs[int(id)] = &jobRecord{
		name:   "",
		spec:   spec,
		status: runtimecontract.CronJobStatusPending,
	}
	s.specs[int(id)] = spec
	s.mu.Unlock()
	cronJobsScheduled.Inc()
	return int(id), nil
}

// AddNamed registers a cron job with Prometheus metrics labels.
// Suitable for tasks requiring execution monitoring by name.
// Core logic: Wrap function with timeout, add metrics recording, add to scheduler, track execution history.
//
// AddNamed 注册带 Prometheus 指标标签的定时任务。
// 适用于需要按任务名观测执行情况的场景。
// 核心逻辑：用超时包装函数、添加指标记录、添加到调度器、追踪执行历史。
func (s *Service) AddNamed(name, spec string, fn func(ctx context.Context) error) (int, error) {
	var id cron.EntryID
	var err error
	id, err = s.c.AddFunc(spec, func() {
		start := time.Now()
		cronJobsInProgress.WithLabelValues(name).Inc()
		defer cronJobsInProgress.WithLabelValues(name).Dec()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		runErr := fn(ctx)
		took := time.Since(start)

		status := "success"
		if runErr != nil {
			status = "error"
		}
		duration := took.Seconds()
		cronJobsTotal.WithLabelValues(name, status).Inc()
		cronJobDuration.WithLabelValues(name).Observe(duration)

		s.recordExecution(int(id), runErr, start, took)
	})
	if err != nil {
		return int(id), err
	}
	s.mu.Lock()
	s.jobs[int(id)] = &jobRecord{
		name:   name,
		spec:   spec,
		status: runtimecontract.CronJobStatusPending,
	}
	s.specs[int(id)] = spec
	s.mu.Unlock()
	cronJobsScheduled.Inc()
	return int(id), nil
}

// Start begins the cron scheduler.
// Registered jobs will start executing according to their schedules.
//
// Start 启动 Cron 调度器。
// 已注册的任务将按调度表达式开始执行。
func (s *Service) Start() {
	s.c.Start()
}

// Stop stops the cron scheduler and returns context for waiting on running jobs.
//
// Stop 停止 Cron 调度器，返回 context 用于等待正在执行的任务完成。
func (s *Service) Stop() context.Context {
	return s.c.Stop()
}

// Jobs returns information about all registered cron jobs, including
// schedule, last/next run time, and execution status.
//
// Jobs 返回所有已注册 cron 任务的信息，包括调度表达式、上次/下次执行时间及执行状态。
func (s *Service) Jobs() []runtimecontract.CronJobEntry {
	entries := s.c.Entries()
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]runtimecontract.CronJobEntry, 0, len(entries))
	for _, e := range entries {
		entryID := int(e.ID)
		rec, hasRec := s.jobs[entryID]

		ent := runtimecontract.CronJobEntry{
			ID:          entryID,
			NextRunTime: e.Next,
		}

		if hasRec {
			ent.Name = rec.name
			ent.Spec = rec.spec
			ent.Status = rec.status
			ent.LastRunTime = rec.lastRun
			ent.LastError = rec.err
			ent.LastDuration = rec.took
		} else {
			// Fallback: use spec from specs map or synthesize from entry
			ent.Spec = s.specs[entryID]
			ent.Status = runtimecontract.CronJobStatusPending
			if !e.Prev.IsZero() {
				ent.LastRunTime = e.Prev
				ent.Status = runtimecontract.CronJobStatusSuccess
			}
		}

		result = append(result, ent)
	}
	return result
}

// recordExecution updates the execution record for a job after each run.
//
// recordExecution 在每次执行后更新任务的执行记录。
func (s *Service) recordExecution(entryID int, err error, start time.Time, took time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, ok := s.jobs[entryID]
	if !ok {
		rec = &jobRecord{}
		s.jobs[entryID] = rec
	}

	rec.lastRun = start
	rec.took = took

	if err != nil {
		rec.status = runtimecontract.CronJobStatusError
		rec.err = fmt.Sprintf("%v", err)
	} else {
		rec.status = runtimecontract.CronJobStatusSuccess
		rec.err = ""
	}
}
