package gradient_test

import (
	"testing"
	"time"

	"github.com/ngq/gorp/framework/provider/loadshedding/gradient"
	"github.com/stretchr/testify/require"
)

func TestGradientLoadShedder_NormalTraffic(t *testing.T) {
	shedder := gradient.NewGradientLoadShedder(gradient.GradientConfig{
		MinLimit:     10,
		MaxLimit:     500,
		InitialLimit: 50,
	})

	ctx := t.Context()

	// 1. Normal requests under limit should be allowed
	for i := 0; i < 20; i++ {
		err := shedder.Allow(ctx, "api")
		require.NoError(t, err)
	}

	require.Equal(t, int64(20), shedder.InFlight())
}

func TestGradientLoadShedder_RTTOverloadAdjustment(t *testing.T) {
	shedder := gradient.NewGradientLoadShedder(gradient.GradientConfig{
		MinLimit:     10,
		MaxLimit:     100,
		InitialLimit: 100,
		Smoothing:    0.5,
	})

	// Establish Baseline minRTT = 10ms
	shedder.RecordSample(10 * time.Millisecond)

	initialLimit := shedder.Limit()

	// Simulate latency spike: currentRTT jumps from 10ms to 100ms (10x congestion)
	for i := 0; i < 10; i++ {
		shedder.RecordSample(100 * time.Millisecond)
	}

	newLimit := shedder.Limit()
	// Gradient = minRTT / currentRTT = 10 / 100 = 0.1 -> Limit drops significantly!
	require.Less(t, newLimit, initialLimit)
	require.GreaterOrEqual(t, newLimit, 10.0) // MinLimit floor respected
}
