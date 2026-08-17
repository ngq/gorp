// Package bootstrap_test provides integration tests for governance mode and capability provider selection.
//
// 适用场景：
// - 验证 governance mode 的检测、标准化与模式感知选择逻辑。
// - 验证各 provider backend 的 Select 优先級（backend key > config > code disable > default）。
// - 验证 RegisterSelectedMicroserviceProviders 的重载、传播与降级行为。
// - 验证 governance override 链路的优先级顺序。
//
// 注意：contrib 组件现在是独立模块。按 fail-fast 决策：显式配置（含
// enabled 推断出的默认后端）但 provider 未注册时，选择器返回 failingProvider
// 使启动失败；只有未配置时才回退 noop/local。
package bootstrap

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// assertFailingProvider 断言选择器对未知/未注册后端返回 fail-fast provider。
func assertFailingProvider(t *testing.T, p runtimecontract.ServiceProvider) {
	t.Helper()
	if p == nil {
		t.Fatal("expected failing provider, got nil")
	}
	if p.Name() != "capability.invalid" {
		t.Fatalf("expected fail-fast provider, got %s", p.Name())
	}
	if err := p.Register(nil); err == nil {
		t.Fatal("expected Register to return error for unknown backend")
	}
}

type selectorConfigStub struct {
	values map[string]any
}

func (s *selectorConfigStub) Env() string        { return "test" }
func (s *selectorConfigStub) Get(key string) any { return s.values[key] }
func (s *selectorConfigStub) GetString(key string) string {
	if v, ok := s.values[key].(string); ok {
		return v
	}
	return ""
}
func (s *selectorConfigStub) GetInt(key string) int {
	if v, ok := s.values[key].(int); ok {
		return v
	}
	return 0
}
func (s *selectorConfigStub) GetBool(key string) bool {
	if v, ok := s.values[key].(bool); ok {
		return v
	}
	return false
}
func (s *selectorConfigStub) GetFloat(key string) float64 {
	if v, ok := s.values[key].(float64); ok {
		return v
	}
	return 0
}
func (s *selectorConfigStub) Unmarshal(key string, out any) error { return nil }
func (s *selectorConfigStub) Watch(ctx context.Context, key string) (data.ConfigWatcher, error) {
	return nil, nil
}
func (s *selectorConfigStub) Reload(ctx context.Context) error { return nil }

// =============================================================================
// Provider 选择测试 - Select 和 Fallback 行为
// 注意：contrib 组件是独立模块，未注册时会回退到 noop/local
// =============================================================================

func TestSelectConfigSourceProvider_PrefersBackendKey(t *testing.T) {
	// nacos 是 contrib 组件，显式配置但未注册 → fail-fast
	cfg := &selectorConfigStub{values: map[string]any{"configsource.backend": "nacos"}}
	assertFailingProvider(t, SelectConfigSourceProvider(cfg))
}

func TestSelectDiscoveryProvider_PrefersBackendKey(t *testing.T) {
	// eureka 是 contrib 组件，显式配置但未注册 → fail-fast
	cfg := &selectorConfigStub{values: map[string]any{"discovery.backend": "eureka"}}
	assertFailingProvider(t, SelectDiscoveryProvider(cfg))
}

func TestSelectRPCProvider_DefaultsToNoop(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{}}
	if got := SelectRPCProvider(cfg).Name(); got != "rpc.noop" {
		t.Fatalf("expected rpc.noop, got %s", got)
	}
}

func TestSelectConfigSourceProvider_UnsetDefaultsToLocal(t *testing.T) {
	// 未配置 → 默认 local（不 fail）
	cfg := &selectorConfigStub{values: map[string]any{}}
	if got := SelectConfigSourceProvider(cfg).Name(); got != "configsource.local" {
		t.Fatalf("expected configsource.local for unset backend, got %s", got)
	}

	// 显式未知后端 → fail-fast
	unknownCfg := &selectorConfigStub{values: map[string]any{"configsource.backend": "unknown"}}
	assertFailingProvider(t, SelectConfigSourceProvider(unknownCfg))
}

func TestSelectDiscoveryProvider_UnknownBackendFailsFast(t *testing.T) {
	// 显式配置未知后端 → fail-fast（不再静默回退 noop）
	cfg := &selectorConfigStub{values: map[string]any{"discovery.backend": "unknown"}}
	assertFailingProvider(t, SelectDiscoveryProvider(cfg))
}

func TestSelectMessageQueueProvider_UnknownBackendFailsFast(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"message_queue.backend": "unknown"}}
	assertFailingProvider(t, SelectMessageQueueProvider(cfg))
}

func TestSelectDistributedLockProvider_UnknownBackendFailsFast(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"distributed_lock.backend": "unknown"}}
	assertFailingProvider(t, SelectDistributedLockProvider(cfg))
}

func TestSelectCircuitBreakerProvider_AcceptsBackendAndEnabled(t *testing.T) {
	// sentinel 是 contrib 组件，显式配置但未注册 → fail-fast
	backendCfg := &selectorConfigStub{values: map[string]any{"circuit_breaker.backend": "sentinel"}}
	assertFailingProvider(t, SelectCircuitBreakerProvider(backendCfg))

	// enabled=true 推断默认 sentinel，未注册 → fail-fast
	enabledCfg := &selectorConfigStub{values: map[string]any{"circuit_breaker.enabled": true}}
	assertFailingProvider(t, SelectCircuitBreakerProvider(enabledCfg))

	noopCfg := &selectorConfigStub{values: map[string]any{"circuit_breaker.backend": "noop"}}
	if got := SelectCircuitBreakerProvider(noopCfg).Name(); got != "circuitbreaker.noop" {
		t.Fatalf("expected circuitbreaker.noop, got %s", got)
	}
}

func TestSelectLoadSheddingProvider_AcceptsBackendAndEnabled(t *testing.T) {
	// semaphore 是 framework 内建 provider
	backendCfg := &selectorConfigStub{values: map[string]any{"load_shedding.backend": "semaphore"}}
	if got := SelectLoadSheddingProvider(backendCfg).Name(); got != "loadshedding.semaphore" {
		t.Fatalf("expected loadshedding.semaphore, got %s", got)
	}

	enabledCfg := &selectorConfigStub{values: map[string]any{"load_shedding.enabled": true}}
	if got := SelectLoadSheddingProvider(enabledCfg).Name(); got != "loadshedding.semaphore" {
		t.Fatalf("expected enabled config to select semaphore, got %s", got)
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"load_shedding.backend": "noop"}}
	if got := SelectLoadSheddingProvider(noopCfg).Name(); got != "loadshedding.noop" {
		t.Fatalf("expected loadshedding.noop, got %s", got)
	}
}

func TestSelectDTMProvider_AcceptsBackendDriverAndEnabled(t *testing.T) {
	// dtmsdk 是 contrib 组件，显式配置但未注册 → fail-fast
	backendCfg := &selectorConfigStub{values: map[string]any{"dtm.backend": "dtmsdk"}}
	assertFailingProvider(t, SelectDTMProvider(backendCfg))

	driverCfg := &selectorConfigStub{values: map[string]any{"dtm.driver": "sdk"}}
	assertFailingProvider(t, SelectDTMProvider(driverCfg))

	enabledCfg := &selectorConfigStub{values: map[string]any{"dtm.enabled": true}}
	assertFailingProvider(t, SelectDTMProvider(enabledCfg))

	noopCfg := &selectorConfigStub{values: map[string]any{"dtm.backend": "noop"}}
	if got := SelectDTMProvider(noopCfg).Name(); got != "dtm.noop" {
		t.Fatalf("expected dtm.noop, got %s", got)
	}
}

func TestSelectDTMProvider_UnknownBackendFailsFast(t *testing.T) {
	// 未知 backend → fail-fast
	unknownCfg := &selectorConfigStub{values: map[string]any{"dtm.backend": "unknown"}}
	assertFailingProvider(t, SelectDTMProvider(unknownCfg))

	// 空/零值配置默认 noop（不 fail）
	emptyCfg := &selectorConfigStub{values: map[string]any{}}
	if got := SelectDTMProvider(emptyCfg).Name(); got != "dtm.noop" {
		t.Fatalf("expected dtm.noop for empty config, got %s", got)
	}
}

func TestSelectTracingProvider_AcceptsEnabledAndBackends(t *testing.T) {
	// otel 等后端是 contrib 组件，显式配置但未注册 → fail-fast
	backendCases := []string{"otel", "otlp", "grpc", "http", "stdout"}
	for _, backend := range backendCases {
		cfg := &selectorConfigStub{values: map[string]any{"tracing.backend": backend}}
		assertFailingProvider(t, SelectTracingProvider(cfg))
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"tracing.backend": "noop"}}
	if got := SelectTracingProvider(noopCfg).Name(); got != "tracing.noop" {
		t.Fatalf("expected tracing.noop, got %s", got)
	}

	// enabled=true 推断默认 otel，未注册 → fail-fast
	enabledCfg := &selectorConfigStub{values: map[string]any{"tracing.enabled": true}}
	assertFailingProvider(t, SelectTracingProvider(enabledCfg))
}

func TestSelectMetadataProvider_AcceptsEnabledAndPrefix(t *testing.T) {
	// metadata.default 是 framework 内建 provider
	enabledCfg := &selectorConfigStub{values: map[string]any{"metadata.enabled": true}}
	if got := SelectMetadataProvider(enabledCfg).Name(); got != "metadata.default" {
		t.Fatalf("expected enabled metadata to select metadata.default, got %s", got)
	}

	prefixCfg := &selectorConfigStub{values: map[string]any{"metadata.propagate_prefix": "x-"}}
	if got := SelectMetadataProvider(prefixCfg).Name(); got != "metadata.default" {
		t.Fatalf("expected propagate_prefix to select metadata.default, got %s", got)
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"metadata.backend": "noop"}}
	if got := SelectMetadataProvider(noopCfg).Name(); got != "metadata.noop" {
		t.Fatalf("expected metadata.noop, got %s", got)
	}
}

func TestSelectServiceAuthProvider_AcceptsEnabledAndMode(t *testing.T) {
	// enabled=true 推断默认 token（contrib），未注册 → fail-fast
	enabledCfg := &selectorConfigStub{values: map[string]any{"service_auth.enabled": true}}
	assertFailingProvider(t, SelectServiceAuthProvider(enabledCfg))

	// mtls 是 contrib 组件，显式配置但未注册 → fail-fast
	mtlsCfg := &selectorConfigStub{values: map[string]any{"service_auth.mode": "mtls"}}
	assertFailingProvider(t, SelectServiceAuthProvider(mtlsCfg))

	noopCfg := &selectorConfigStub{values: map[string]any{"service_auth.backend": "noop"}}
	if got := SelectServiceAuthProvider(noopCfg).Name(); got != "serviceauth.noop" {
		t.Fatalf("expected serviceauth.noop, got %s", got)
	}
}

func TestSelectMessageQueueProvider_AcceptsEnabledAndBackend(t *testing.T) {
	// enabled=true 推断默认 redis（contrib），未注册 → fail-fast
	enabledCfg := &selectorConfigStub{values: map[string]any{"message_queue.enabled": true}}
	assertFailingProvider(t, SelectMessageQueueProvider(enabledCfg))

	noopCfg := &selectorConfigStub{values: map[string]any{"message_queue.backend": "noop"}}
	if got := SelectMessageQueueProvider(noopCfg).Name(); got != "messagequeue.noop" {
		t.Fatalf("expected messagequeue.noop, got %s", got)
	}
}

// TestSelectMessageQueueProvider_ContribBackendsFailFast verifies that contrib
// backends specified but not registered produce a fail-fast provider.
//
// TestSelectMessageQueueProvider_ContribBackendsFailFast 验证指定 contrib 后端
// 但未注册时，选择器返回 fail-fast provider（不再静默回退 noop）。
func TestSelectMessageQueueProvider_ContribBackendsFailFast(t *testing.T) {
	cases := []string{"kafka", "rabbitmq", "rocketmq", "redis"}
	for _, backend := range cases {
		cfg := &selectorConfigStub{values: map[string]any{"message_queue.backend": backend}}
		assertFailingProvider(t, SelectMessageQueueProvider(cfg))
	}
}

func TestSelectDistributedLockProvider_AcceptsEnabledAndBackend(t *testing.T) {
	// enabled=true 推断默认 redis（contrib），未注册 → fail-fast
	enabledCfg := &selectorConfigStub{values: map[string]any{"distributed_lock.enabled": true}}
	assertFailingProvider(t, SelectDistributedLockProvider(enabledCfg))

	noopCfg := &selectorConfigStub{values: map[string]any{"distributed_lock.backend": "noop"}}
	if got := SelectDistributedLockProvider(noopCfg).Name(); got != "dlock.noop" {
		t.Fatalf("expected dlock.noop, got %s", got)
	}
}
