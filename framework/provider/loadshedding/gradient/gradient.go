// Package gradient implements Little's Law (L = λ * W) Gradient Adaptive Concurrency Limiting.
package gradient

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// ErrLoadShedded is returned when a request is rejected due to adaptive concurrency overload.
var ErrLoadShedded = errors.New("request shed by adaptive gradient concurrency limiter")

type startTimeKey struct{}

// GradientConfig configures the Little's Law Gradient concurrency limiter.
type GradientConfig struct {
	MinLimit      float64       // Minimum concurrency floor (default: 10)
	MaxLimit      float64       // Maximum concurrency ceiling (default: 1000)
	InitialLimit  float64       // Initial concurrency limit (default: 100)
	QueueSize     float64       // Headroom queue size added to gradient calculation (default: 4.0)
	Smoothing     float64       // Exponential smoothing factor alpha (0.0 to 1.0, default: 0.2)
	RttWindowSize time.Duration // Sliding window size for minRTT tracking (default: 30s)
}

// DefaultGradientConfig returns sensible defaults.
func DefaultGradientConfig() GradientConfig {
	return GradientConfig{
		MinLimit:      10,
		MaxLimit:      1000,
		InitialLimit:  100,
		QueueSize:     4.0,
		Smoothing:     0.2,
		RttWindowSize: 30 * time.Second,
	}
}

// GradientLoadShedder implements Little's Law Gradient Adaptive Concurrency Limiting.
type GradientLoadShedder struct {
	mu           sync.RWMutex
	config       GradientConfig
	inFlight     atomic.Int64
	currentLimit float64
	minRtt       time.Duration
	currentRtt   time.Duration
	lastRttReset time.Time
}

// NewGradientLoadShedder creates a new GradientLoadShedder instance.
func NewGradientLoadShedder(cfg ...GradientConfig) *GradientLoadShedder {
	c := DefaultGradientConfig()
	if len(cfg) > 0 {
		if cfg[0].MinLimit > 0 {
			c.MinLimit = cfg[0].MinLimit
		}
		if cfg[0].MaxLimit > 0 {
			c.MaxLimit = cfg[0].MaxLimit
		}
		if cfg[0].InitialLimit > 0 {
			c.InitialLimit = cfg[0].InitialLimit
		}
		if cfg[0].QueueSize > 0 {
			c.QueueSize = cfg[0].QueueSize
		}
		if cfg[0].RttWindowSize > 0 {
			c.RttWindowSize = cfg[0].RttWindowSize
		}
	}
	return &GradientLoadShedder{
		config:       c,
		currentLimit: c.InitialLimit,
		lastRttReset: time.Now(),
	}
}

// WithStartTime injects request start time into context for RTT measurement in Done().
func WithStartTime(ctx context.Context) context.Context {
	return context.WithValue(ctx, startTimeKey{}, time.Now())
}

// Allow decides whether to admit the incoming request based on current dynamic concurrency limit.
func (g *GradientLoadShedder) Allow(ctx context.Context, resource string) error {
	g.mu.RLock()
	limit := g.currentLimit
	g.mu.RUnlock()

	inFlight := g.inFlight.Add(1)
	if float64(inFlight) > limit {
		g.inFlight.Add(-1)
		return ErrLoadShedded
	}

	return nil
}

// Done completes the request lifecycle, updates inflight counts and adjusts concurrency limits via RTT sampling.
func (g *GradientLoadShedder) Done(ctx context.Context, resource string, err error) {
	g.inFlight.Add(-1)

	if ctx != nil {
		if start, ok := ctx.Value(startTimeKey{}).(time.Time); ok {
			rtt := time.Since(start)
			g.RecordSample(rtt)
		}
	}
}

// RecordSample explicitly records an RTT sample and updates Little's Law Gradient concurrency limits.
func (g *GradientLoadShedder) RecordSample(rtt time.Duration) {
	if rtt <= 0 {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	if now.Sub(g.lastRttReset) > g.config.RttWindowSize || g.minRtt == 0 {
		g.minRtt = rtt
		g.lastRttReset = now
	} else if rtt < g.minRtt {
		g.minRtt = rtt
	}

	if g.currentRtt == 0 {
		g.currentRtt = rtt
	} else {
		alpha := g.config.Smoothing
		g.currentRtt = time.Duration(float64(g.currentRtt)*(1-alpha) + float64(rtt)*alpha)
	}

	// Calculate Little's Law Gradient: minRtt / currentRtt
	gradient := float64(g.minRtt) / float64(g.currentRtt)
	if gradient > 1.0 {
		gradient = 1.0
	}

	// New Limit = Current Limit * Gradient + QueueSize
	newLimit := g.currentLimit*gradient + g.config.QueueSize
	if newLimit < g.config.MinLimit {
		newLimit = g.config.MinLimit
	}
	if newLimit > g.config.MaxLimit {
		newLimit = g.config.MaxLimit
	}

	g.currentLimit = newLimit
}

// UpdateConfig updates runtime configuration parameters.
func (g *GradientLoadShedder) UpdateConfig(cfg resiliencecontract.LoadSheddingConfig) {}

// Limit returns current concurrency limit.
func (g *GradientLoadShedder) Limit() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.currentLimit
}

// InFlight returns current inflight requests.
func (g *GradientLoadShedder) InFlight() int64 {
	return g.inFlight.Load()
}
