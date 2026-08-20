package provider_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/ngq/gorp/framework/provider/loadshedding"
	"github.com/ngq/gorp/framework/provider/ratelimiter"
	"github.com/ngq/gorp/framework/provider/retry"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_DynamicHotReload(t *testing.T) {
	ctx := context.Background()

	// Initial config: 5 QPS, Burst 5
	cfg := resiliencecontract.RateLimiterConfig{
		Enabled: true,
		DefaultConfig: resiliencecontract.RateResourceConfig{
			QPS:   5,
			Burst: 5,
		},
	}
	rl := ratelimiter.NewTokenBucketRateLimiter(cfg)

	// Consume 5 tokens
	for i := 0; i < 5; i++ {
		require.NoError(t, rl.Allow(ctx, "api.checkout"))
	}
	// 6th should be rejected
	require.Error(t, rl.Allow(ctx, "api.checkout"))

	// Dynamic Hot Reload: Upgrade to 1000 QPS, Burst 1000
	newCfg := resiliencecontract.RateLimiterConfig{
		Enabled: true,
		DefaultConfig: resiliencecontract.RateResourceConfig{
			QPS:   1000,
			Burst: 1000,
		},
		ResourceConfigs: map[string]resiliencecontract.RateResourceConfig{
			"api.checkout": {
				QPS:   2000,
				Burst: 2000,
			},
		},
	}
	rl.UpdateConfig(newCfg)

	// Immediately retry: should succeed without restart!
	require.NoError(t, rl.Allow(ctx, "api.checkout"))
}

func TestLoadShedder_DynamicHotReload(t *testing.T) {
	ctx := context.Background()

	// Initial config: max concurrency 2
	cfg := resiliencecontract.LoadSheddingConfig{
		Enabled:        true,
		MaxConcurrency: 2,
	}
	prov := loadshedding.NewProvider()
	_ = prov

	// Test semaphore implementation directly through provider register / new
	// Concurrency test
	shedder := loadshedding.NewLoadShedderFromConfig(cfg)
	require.NoError(t, shedder.Allow(ctx, "order.create"))
	require.NoError(t, shedder.Allow(ctx, "order.create"))
	// 3rd should be shed
	require.Error(t, shedder.Allow(ctx, "order.create"))

	// Dynamic Hot Reload: expand capacity to 10
	shedder.UpdateConfig(resiliencecontract.LoadSheddingConfig{
		Enabled:        true,
		MaxConcurrency: 10,
	})

	// Now should allow
	require.NoError(t, shedder.Allow(ctx, "order.create"))
	shedder.Done(ctx, "order.create", nil)
}

func TestRetry_DynamicHotReload(t *testing.T) {
	ctx := context.Background()

	cfg := &resiliencecontract.RetryConfig{
		Enabled: true,
		DefaultPolicy: resiliencecontract.RetryPolicy{
			MaxAttempts:  1, // no retry
			InitialDelay: 1 * time.Millisecond,
		},
	}
	svc := retry.NewRetryService(cfg)

	attempts := 0
	err := svc.Do(ctx, func() error {
		attempts++
		return errors.New("connection reset by peer")
	})
	require.Error(t, err)
	require.Equal(t, 1, attempts)

	// Dynamic Hot Reload: enable 3 attempts
	svc.UpdateConfig(resiliencecontract.RetryConfig{
		Enabled: true,
		DefaultPolicy: resiliencecontract.RetryPolicy{
			MaxAttempts:  3,
			InitialDelay: 1 * time.Millisecond,
			Multiplier:   1.0,
		},
	})

	attempts = 0
	_ = svc.Do(ctx, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("connection reset by peer")
		}
		return nil
	})
	require.Equal(t, 3, attempts)
}

func TestRateLimiter_ConcurrentHotReloadSafety(t *testing.T) {
	ctx := context.Background()
	rl := ratelimiter.NewTokenBucketRateLimiter(resiliencecontract.RateLimiterConfig{
		Enabled: true,
		DefaultConfig: resiliencecontract.RateResourceConfig{
			QPS:   100,
			Burst: 200,
		},
	})

	var wg sync.WaitGroup
	// 50 goroutines making requests
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = rl.Allow(ctx, "resource.fast")
			}
		}()
	}

	// 5 goroutines dynamically updating config at the same time
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				rl.UpdateConfig(resiliencecontract.RateLimiterConfig{
					Enabled: true,
					DefaultConfig: resiliencecontract.RateResourceConfig{
						QPS:   float64(100 * (idx + 1)),
						Burst: 200 * (idx + 1),
					},
				})
			}
		}(i)
	}

	wg.Wait()
}
