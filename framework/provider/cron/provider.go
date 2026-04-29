package cron

import (
	"context"
	"time"

	"github.com/ngq/gorp/framework/contract"

	"github.com/robfig/cron/v3"
)

// Provider 把 Cron 服务注册进容器。
//
// 中文说明：
// - 统一暴露 contract.CronKey；
// - 让业务和 starter 通过同一个调度入口注册定时任务；
// - 底层使用 robfig/cron，但上层不需要直接依赖其装配细节。
type Provider struct{}

// NewProvider 创建 cron provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 provider 名称。
func (p *Provider) Name() string { return "cron" }

// IsDefer 表示 cron provider 不走延迟加载。
func (p *Provider) IsDefer() bool { return false }

// Provides 返回 cron provider 暴露的能力 key。
func (p *Provider) Provides() []string { return []string{contract.CronKey} }

// Register 绑定 Cron 服务。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.CronKey, func(contract.Container) (any, error) {
		return NewService(), nil
	}, true)
	return nil
}

// Boot cron provider 无额外启动逻辑。
func (p *Provider) Boot(contract.Container) error { return nil }

// Service 对 robfig/cron 做 framework 级封装。
//
// 中文说明：
// - 对外收口 framework 自己的定时任务注册心智；
// - 内部补了超时控制与 Prometheus 指标采集；
// - 让业务侧不必在每个 cron job 外层重复写这些样板。
type Service struct {
	c *cron.Cron
}

// NewService 创建支持秒级表达式的 Cron 调度器。
func NewService() *Service {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	c := cron.New(
		cron.WithParser(parser),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)
	return &Service{c: c}
}

// Add 注册一个不带指标标签的定时任务。
//
// 中文说明：
// - 适合简单定时任务场景；
// - 统一补 5 分钟超时，避免任务无限悬挂。
func (s *Service) Add(spec string, fn func(ctx context.Context) error) (int, error) {
	id, err := s.c.AddFunc(spec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		_ = fn(ctx)
	})
	cronJobsScheduled.Inc()
	return int(id), err
}

// AddNamed 注册一个带 Prometheus 指标标签的定时任务。
//
// 中文说明：
// - 适合需要按任务名观测执行情况的场景；
// - 会记录执行中数量、总次数、耗时和成功/失败状态。
func (s *Service) AddNamed(name, spec string, fn func(ctx context.Context) error) (int, error) {
	id, err := s.c.AddFunc(spec, func() {
		start := time.Now()
		cronJobsInProgress.WithLabelValues(name).Inc()
		defer cronJobsInProgress.WithLabelValues(name).Dec()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		err := fn(ctx)

		status := "success"
		if err != nil {
			status = "error"
		}
		duration := time.Since(start).Seconds()
		cronJobsTotal.WithLabelValues(name, status).Inc()
		cronJobDuration.WithLabelValues(name).Observe(duration)
	})
	cronJobsScheduled.Inc()
	return int(id), err
}

// Start 启动 Cron 调度器。
//
// 中文说明：
// - 由宿主或业务启动流程调用后，已注册任务才会开始调度。
func (s *Service) Start() {
	s.c.Start()
}

// Stop 停止 Cron 调度器。
//
// 中文说明：
// - 返回底层停止上下文，便于上层等待正在执行中的任务完成。
func (s *Service) Stop() context.Context {
	return s.c.Stop()
}