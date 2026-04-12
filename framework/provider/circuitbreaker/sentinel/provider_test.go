package sentinel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/assert"
)

func TestSentinelCircuitBreaker_Do_Success(t *testing.T) {
	cfg := &contract.CircuitBreakerConfig{
		Enabled:  false, // 禁用 Sentinel 初始化
		Strategy: "sentinel",
	}
	cb := NewSentinelCircuitBreaker(cfg)

	// 执行成功的函数
	err := cb.Do(context.Background(), "test-resource", func() error {
		return nil
	})
	assert.NoError(t, err)
}

func TestSentinelCircuitBreaker_Do_Error(t *testing.T) {
	cfg := &contract.CircuitBreakerConfig{
		Enabled:  false,
		Strategy: "sentinel",
	}
	cb := NewSentinelCircuitBreaker(cfg)

	// 执行失败的函数
	execErr := errors.New("execution error")
	err := cb.Do(context.Background(), "test-resource", func() error {
		return execErr
	})
	assert.Error(t, err)
	assert.Equal(t, execErr, err)
}

func TestSentinelCircuitBreaker_State(t *testing.T) {
	cfg := &contract.CircuitBreakerConfig{
		Enabled:  false,
		Strategy: "sentinel",
	}
	cb := NewSentinelCircuitBreaker(cfg)

	// 未配置规则时返回 Closed 状态
	state := cb.State(context.Background(), "test-resource")
	assert.Equal(t, contract.CircuitBreakerStateClosed, state)
}

func TestSentinelRateLimiter_Allow(t *testing.T) {
	cfg := &contract.CircuitBreakerConfig{
		Enabled:  false,
		Strategy: "sentinel",
	}
	rl := NewSentinelRateLimiter(cfg)

	// 未配置规则时允许请求
	err := rl.Allow(context.Background(), "test-resource")
	assert.NoError(t, err)
}

func TestSentinelRateLimiter_AllowN(t *testing.T) {
	cfg := &contract.CircuitBreakerConfig{
		Enabled:  false,
		Strategy: "sentinel",
	}
	rl := NewSentinelRateLimiter(cfg)

	// 未配置规则时允许批量请求
	err := rl.AllowN(context.Background(), "test-resource", 5)
	assert.NoError(t, err)
}

func TestSentinelRateLimiter_Wait(t *testing.T) {
	cfg := &contract.CircuitBreakerConfig{
		Enabled:  false,
		Strategy: "sentinel",
	}
	rl := NewSentinelRateLimiter(cfg)

	err := rl.Wait(context.Background(), "test-resource")
	assert.NoError(t, err)
}

func TestSentinelRateLimiter_WaitTimeout(t *testing.T) {
	cfg := &contract.CircuitBreakerConfig{
		Enabled:  false,
		Strategy: "sentinel",
	}
	rl := NewSentinelRateLimiter(cfg)

	// 短超时应该允许
	err := rl.WaitTimeout(context.Background(), "test-resource", 100*time.Millisecond)
	assert.NoError(t, err)
}

func TestSentinelReservation(t *testing.T) {
	r := &sentinelReservation{}

	assert.True(t, r.OK())
	assert.Equal(t, time.Duration(0), r.Delay())
	r.Cancel()
	r.CancelAt(time.Now())
}

func TestProvider_Name(t *testing.T) {
	p := NewProvider()
	assert.Equal(t, "circuitbreaker.sentinel", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{contract.CircuitBreakerKey, contract.RateLimiterKey}, p.Provides())
}