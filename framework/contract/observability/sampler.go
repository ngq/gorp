// Package observability provides adaptive sampling capabilities for distributed tracing.
package observability

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// AdaptiveSamplerConfig defines parameters for tail-based adaptive trace sampling.
type AdaptiveSamplerConfig struct {
	BaseRatio     float64       // Baseline sampling ratio for normal fast requests (default: 0.01 / 1%)
	SlowThreshold time.Duration // Latency threshold for slow requests (default: 500ms)
	SampleOnError bool          // 100% force sample when status >= 500 or err != nil (default: true)
	SampleOnSlow  bool          // 100% force sample when latency >= SlowThreshold (default: true)
}

// DefaultAdaptiveSamplerConfig returns default production settings.
func DefaultAdaptiveSamplerConfig() AdaptiveSamplerConfig {
	return AdaptiveSamplerConfig{
		BaseRatio:     0.01,
		SlowThreshold: 500 * time.Millisecond,
		SampleOnError: true,
		SampleOnSlow:  true,
	}
}

// AdaptiveSampler determines whether a trace should be sampled based on request outcome and latency.
type AdaptiveSampler struct {
	config AdaptiveSamplerConfig
	mu     sync.Mutex
	rng    *rand.Rand
}

// NewAdaptiveSampler creates a new AdaptiveSampler instance.
func NewAdaptiveSampler(cfg ...AdaptiveSamplerConfig) *AdaptiveSampler {
	c := DefaultAdaptiveSamplerConfig()
	if len(cfg) > 0 {
		if cfg[0].BaseRatio > 0 {
			c.BaseRatio = cfg[0].BaseRatio
		}
		if cfg[0].SlowThreshold > 0 {
			c.SlowThreshold = cfg[0].SlowThreshold
		}
		c.SampleOnError = cfg[0].SampleOnError
		c.SampleOnSlow = cfg[0].SampleOnSlow
	}
	return &AdaptiveSampler{
		config: c,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ShouldSample decides if a trace span should be sampled and exported.
func (s *AdaptiveSampler) ShouldSample(ctx context.Context, statusCode int, latency time.Duration, err error) bool {
	// 1. 5xx Errors or Panics -> 100% Force Sample
	if s.config.SampleOnError && (statusCode >= 500 || err != nil) {
		return true
	}

	// 2. Slow Requests (Latency >= SlowThreshold) -> 100% Force Sample
	if s.config.SampleOnSlow && latency >= s.config.SlowThreshold {
		return true
	}

	// 3. Normal Requests -> Baseline Ratio Sampling
	if s.config.BaseRatio <= 0 {
		return false
	}
	if s.config.BaseRatio >= 1.0 {
		return true
	}

	s.mu.Lock()
	val := s.rng.Float64()
	s.mu.Unlock()

	return val < s.config.BaseRatio
}
