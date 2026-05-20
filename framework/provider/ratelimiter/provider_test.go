// Package ratelimiter_test provides unit tests for rate limiter provider.
package ratelimiter_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/ngq/gorp/framework/provider/ratelimiter"
	"github.com/stretchr/testify/require"
)

// TestRateLimiter_Allow_Basic 验证基本限流行为。
func TestRateLimiter_Allow_Basic(t *testing.T) {
	cfg := resiliencecontract.RateLimiterConfig{
		Enabled: true,
		DefaultConfig: resiliencecontract.RateResourceConfig{
			QPS:   10, // 10 QPS
			Burst: 2,  // 突发 2 个
		},
	}

	limiter := ratelimiter.NewTokenBucketRateLimiter(cfg)

	// 前两个请求应该成功（burst）
	require.NoError(t, limiter.Allow(context.Background(), "test"))
	require.NoError(t, limiter.Allow(context.Background(), "test"))

	// 第三个应该被限流
	err := limiter.Allow(context.Background(), "test")
	require.Error(t, err)
	require.True(t, errors.Is(err, ratelimiter.ErrRateLimited))
}

// TestRateLimiter_Allow_DifferentResources 验证不同资源独立限流。
func TestRateLimiter_Allow_DifferentResources(t *testing.T) {
	cfg := resiliencecontract.RateLimiterConfig{
		Enabled: true,
		ResourceConfigs: map[string]resiliencecontract.RateResourceConfig{
			"resource-a": {QPS: 1, Burst: 1},
			"resource-b": {QPS: 1, Burst: 1},
		},
	}

	limiter := ratelimiter.NewTokenBucketRateLimiter(cfg)

	// 两个资源各消耗一个令牌
	require.NoError(t, limiter.Allow(context.Background(), "resource-a"))
	require.NoError(t, limiter.Allow(context.Background(), "resource-b"))

	// 两个资源都应该被限流
	require.Error(t, limiter.Allow(context.Background(), "resource-a"))
	require.Error(t, limiter.Allow(context.Background(), "resource-b"))
}

// TestRateLimiter_Wait_BlocksUntilTokenAvailable 验证 Wait 会阻塞等待令牌。
func TestRateLimiter_Wait_BlocksUntilTokenAvailable(t *testing.T) {
	cfg := resiliencecontract.RateLimiterConfig{
		Enabled: true,
		DefaultConfig: resiliencecontract.RateResourceConfig{
			QPS:   100, // 100 QPS = 10ms 一个令牌
			Burst: 1,
		},
	}

	limiter := ratelimiter.NewTokenBucketRateLimiter(cfg)

	// 消耗掉 burst
	require.NoError(t, limiter.Allow(context.Background(), "test"))

	// Wait 应该阻塞直到令牌可用
	start := time.Now()
	err := limiter.Wait(context.Background(), "test")
	elapsed := time.Since(start)

	require.NoError(t, err)
	// 应该等待大约 10ms（100 QPS）
	require.GreaterOrEqual(t, elapsed.Milliseconds(), int64(8))
}

// TestRateLimiter_WaitTimeout_RespectsTimeout 验证 WaitTimeout 超时行为。
func TestRateLimiter_WaitTimeout_RespectsTimeout(t *testing.T) {
	cfg := resiliencecontract.RateLimiterConfig{
		Enabled: true,
		DefaultConfig: resiliencecontract.RateResourceConfig{
			QPS:   1, // 1 QPS = 1s 一个令牌
			Burst: 1,
		},
	}

	limiter := ratelimiter.NewTokenBucketRateLimiter(cfg)

	// 消耗掉 burst
	require.NoError(t, limiter.Allow(context.Background(), "test"))

	// WaitTimeout 设置 50ms 超时，应该超时
	ctx := context.Background()
	err := limiter.WaitTimeout(ctx, "test", 50*time.Millisecond)
	require.Error(t, err) // context deadline exceeded
}

// TestRateLimiter_Reserve_AdvancedUsage 验证 Reserve 高级用法。
func TestRateLimiter_Reserve_AdvancedUsage(t *testing.T) {
	cfg := resiliencecontract.RateLimiterConfig{
		Enabled: true,
		DefaultConfig: resiliencecontract.RateResourceConfig{
			QPS:   10,
			Burst: 1,
		},
	}

	limiter := ratelimiter.NewTokenBucketRateLimiter(cfg)

	// 消耗掉 burst
	require.NoError(t, limiter.Allow(context.Background(), "test"))

	// Reserve 应该返回延迟信息
	r := limiter.Reserve(context.Background(), "test")
	require.True(t, r.OK())
	require.Greater(t, r.Delay(), time.Duration(0))

	// 取消预留
	r.Cancel()
}

// TestRateLimiter_ConcurrentAccess 验证并发安全。
func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	cfg := resiliencecontract.RateLimiterConfig{
		Enabled: true,
		DefaultConfig: resiliencecontract.RateResourceConfig{
			QPS:   1, // 低 QPS，减少令牌生成影响
			Burst: 100,
		},
	}

	limiter := ratelimiter.NewTokenBucketRateLimiter(cfg)

	const n = 1000
	var allowed, denied atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := limiter.Allow(context.Background(), "test")
			if err == nil {
				allowed.Add(1)
			} else {
				denied.Add(1)
			}
		}()
	}
	wg.Wait()

	// burst=100，QPS=1，在短时间内最多允许约 100 个（可能有少量新令牌生成）
	// 主要验证：1) 有请求被拒绝 2) 允许数在合理范围内
	require.Greater(t, denied.Load(), int32(0), "应该有请求被限流")
	require.LessOrEqual(t, allowed.Load(), int32(110), "允许数应该在合理范围内（burst + 少量新令牌）")
}

// TestRateLimiter_AllowN_MultipleTokens 验证 AllowN 多令牌请求。
func TestRateLimiter_AllowN_MultipleTokens(t *testing.T) {
	cfg := resiliencecontract.RateLimiterConfig{
		Enabled: true,
		DefaultConfig: resiliencecontract.RateResourceConfig{
			QPS:   10,
			Burst: 5,
		},
	}

	limiter := ratelimiter.NewTokenBucketRateLimiter(cfg)

	// 请求 3 个令牌应该成功
	require.NoError(t, limiter.AllowN(context.Background(), "test", 3))

	// 再请求 3 个应该失败（只剩 2 个）
	require.Error(t, limiter.AllowN(context.Background(), "test", 3))
}
