package cron

import (
	"context"
	"time"

	"github.com/ngq/gorp/framework/contract"

	"github.com/robfig/cron/v3"
)

// Provider registers the cron service into the container.
type Provider struct{}

// NewProvider creates a new cron provider.
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name.
func (p *Provider) Name() string { return "cron" }

// IsDefer returns false because cron is not lazy loaded.
func (p *Provider) IsDefer() bool { return false }

// Provides returns the keys this provider provides.
func (p *Provider) Provides() []string { return []string{contract.CronKey} }

// Register binds the cron service to the container.
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.CronKey, func(contract.Container) (any, error) {
		return NewService(), nil
	}, true)
	return nil
}

// Boot does nothing.
func (p *Provider) Boot(contract.Container) error { return nil }

// Service wraps robfig/cron with additional capabilities.
type Service struct {
	c *cron.Cron
}

// NewService creates a cron scheduler with second-level support.
func NewService() *Service {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	c := cron.New(
		cron.WithParser(parser),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)
	return &Service{c: c}
}

// Add registers a cron job without metrics collection.
func (s *Service) Add(spec string, fn func(ctx context.Context) error) (int, error) {
	id, err := s.c.AddFunc(spec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		_ = fn(ctx)
	})
	cronJobsScheduled.Inc()
	return int(id), err
}

// AddNamed registers a named cron job with Prometheus metrics.
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

// Start starts the cron scheduler.
func (s *Service) Start() {
	s.c.Start()
}

// Stop stops the cron scheduler.
func (s *Service) Stop() context.Context {
	return s.c.Stop()
}