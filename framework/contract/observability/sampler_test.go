package observability_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract/observability"
	"github.com/stretchr/testify/require"
)

func TestAdaptiveSampler_SamplingDecisions(t *testing.T) {
	sampler := observability.NewAdaptiveSampler(observability.AdaptiveSamplerConfig{
		BaseRatio:     0.0, // Base ratio 0 to ensure normal requests are not sampled
		SlowThreshold: 500 * time.Millisecond,
		SampleOnError: true,
		SampleOnSlow:  true,
	})

	ctx := t.Context()

	// 1. Normal Fast Request (200 OK, 10ms) -> Not Sampled (BaseRatio = 0)
	sampled1 := sampler.ShouldSample(ctx, 200, 10*time.Millisecond, nil)
	require.False(t, sampled1)

	// 2. Error Request (500 Internal Error) -> 100% Sampled
	sampled2 := sampler.ShouldSample(ctx, 500, 10*time.Millisecond, nil)
	require.True(t, sampled2)

	// 3. Application Error -> 100% Sampled
	sampled3 := sampler.ShouldSample(ctx, 200, 10*time.Millisecond, errors.New("db timeout"))
	require.True(t, sampled3)

	// 4. Slow Request (Latency 600ms >= 500ms) -> 100% Sampled
	sampled4 := sampler.ShouldSample(ctx, 200, 600*time.Millisecond, nil)
	require.True(t, sampled4)
}
