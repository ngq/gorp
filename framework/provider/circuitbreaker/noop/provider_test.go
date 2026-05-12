// Package noop_test provides unit tests for the circuit breaker noop provider.
//
// 适用场景：
// - 验证熔断器 noop provider 的注册与空操作行为。
package noop

import (
	"context"
	"testing"
	"time"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/stretchr/testify/assert"
)

func TestNoopCircuitBreaker(t *testing.T) {
	cb := &noopCircuitBreaker{}

	// 测试 Allow
	err := cb.Allow(context.Background(), "test-resource")
	assert.NoError(t, err)

	// 测试 RecordSuccess
	cb.RecordSuccess(context.Background(), "test-resource")

	// 测试 RecordFailure
	cb.RecordFailure(context.Background(), "test-resource", nil)

	// 测试 Do
	executed := false
	err = cb.Do(context.Background(), "test-resource", func() error {
		executed = true
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, executed)

	// 测试 State
	state := cb.State(context.Background(), "test-resource")
	assert.Equal(t, resiliencecontract.CircuitBreakerStateClosed, state)
}

func TestNoopRateLimiter(t *testing.T) {
	rl := &noopRateLimiter{}

	// 测试 Allow
	err := rl.Allow(context.Background(), "test-resource")
	assert.NoError(t, err)

	// 测试 AllowN
	err = rl.AllowN(context.Background(), "test-resource", 10)
	assert.NoError(t, err)

	// 测试 Reserve
	reservation := rl.Reserve(context.Background(), "test-resource")
	assert.True(t, reservation.OK())

	// 测试 Wait
	err = rl.Wait(context.Background(), "test-resource")
	assert.NoError(t, err)

	// 测试 WaitTimeout
	err = rl.WaitTimeout(context.Background(), "test-resource", time.Second)
	assert.NoError(t, err)
}

func TestProvider(t *testing.T) {
	p := NewProvider()

	assert.Equal(t, "circuitbreaker.noop", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{
		resiliencecontract.CircuitBreakerKey,
		resiliencecontract.RateLimiterKey,
	}, p.Provides())
}
