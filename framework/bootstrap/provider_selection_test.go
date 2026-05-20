// Package bootstrap_test provides integration tests for governance mode and capability provider selection.
//
// 适用场景：
// - 验证 governance mode 的检测、标准化与模式感知选择逻辑。
// - 验证各 provider backend 的 Select 优先級（backend key > config > code disable > default）。
// - 验证 RegisterSelectedMicroserviceProviders 的重载、传播与降级行为。
// - 验证 governance override 链路的优先级顺序。
//
// 注意：contrib 组件现在是独立模块，这些测试验证框架选择逻辑，
// 当 contrib provider 未注册时，会回退到 noop/local provider。
package bootstrap

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract/data"
)

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
	// nacos 是 contrib 组件，未注册时回退到 local
	cfg := &selectorConfigStub{values: map[string]any{"configsource.backend": "nacos"}}
	if got := SelectConfigSourceProvider(cfg).Name(); got != "configsource.local" {
		t.Fatalf("expected configsource.local (nacos not registered), got %s", got)
	}
}

func TestSelectDiscoveryProvider_PrefersBackendKey(t *testing.T) {
	// eureka 是 contrib 组件，未注册时回退到 noop
	cfg := &selectorConfigStub{values: map[string]any{"discovery.backend": "eureka"}}
	if got := SelectDiscoveryProvider(cfg).Name(); got != "discovery.noop" {
		t.Fatalf("expected discovery.noop (eureka not registered), got %s", got)
	}
}

func TestSelectRPCProvider_DefaultsToNoop(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{}}
	if got := SelectRPCProvider(cfg).Name(); got != "rpc.noop" {
		t.Fatalf("expected rpc.noop, got %s", got)
	}
}

func TestSelectConfigSourceProvider_FallsBackToLocal(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"configsource.backend": "unknown"}}
	if got := SelectConfigSourceProvider(cfg).Name(); got != "configsource.local" {
		t.Fatalf("expected configsource.local fallback, got %s", got)
	}
}

func TestSelectDiscoveryProvider_FallsBackToNoop(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"discovery.backend": "unknown"}}
	if got := SelectDiscoveryProvider(cfg).Name(); got != "discovery.noop" {
		t.Fatalf("expected discovery.noop fallback, got %s", got)
	}
}

func TestSelectMessageQueueProvider_FallsBackToNoop(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"message_queue.backend": "unknown"}}
	if got := SelectMessageQueueProvider(cfg).Name(); got != "messagequeue.noop" {
		t.Fatalf("expected messagequeue.noop fallback, got %s", got)
	}
}

func TestSelectDistributedLockProvider_FallsBackToNoop(t *testing.T) {
	cfg := &selectorConfigStub{values: map[string]any{"distributed_lock.backend": "unknown"}}
	if got := SelectDistributedLockProvider(cfg).Name(); got != "dlock.noop" {
		t.Fatalf("expected dlock.noop fallback, got %s", got)
	}
}

func TestSelectCircuitBreakerProvider_AcceptsBackendAndEnabled(t *testing.T) {
	// sentinel 是 contrib 组件，未注册时回退到 noop
	backendCfg := &selectorConfigStub{values: map[string]any{"circuit_breaker.backend": "sentinel"}}
	if got := SelectCircuitBreakerProvider(backendCfg).Name(); got != "circuitbreaker.noop" {
		t.Fatalf("expected circuitbreaker.noop (sentinel not registered), got %s", got)
	}

	// enabled=true 但 contrib 未注册，回退到 noop
	enabledCfg := &selectorConfigStub{values: map[string]any{"circuit_breaker.enabled": true}}
	if got := SelectCircuitBreakerProvider(enabledCfg).Name(); got != "circuitbreaker.noop" {
		t.Fatalf("expected circuitbreaker.noop (sentinel not registered), got %s", got)
	}

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
	// dtmsdk 是 contrib 组件，未注册时回退到 noop
	backendCfg := &selectorConfigStub{values: map[string]any{"dtm.backend": "dtmsdk"}}
	if got := SelectDTMProvider(backendCfg).Name(); got != "dtm.noop" {
		t.Fatalf("expected dtm.noop (dtmsdk not registered), got %s", got)
	}

	driverCfg := &selectorConfigStub{values: map[string]any{"dtm.driver": "sdk"}}
	if got := SelectDTMProvider(driverCfg).Name(); got != "dtm.noop" {
		t.Fatalf("expected dtm.noop (dtmsdk not registered), got %s", got)
	}

	enabledCfg := &selectorConfigStub{values: map[string]any{"dtm.enabled": true}}
	if got := SelectDTMProvider(enabledCfg).Name(); got != "dtm.noop" {
		t.Fatalf("expected dtm.noop (dtmsdk not registered), got %s", got)
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"dtm.backend": "noop"}}
	if got := SelectDTMProvider(noopCfg).Name(); got != "dtm.noop" {
		t.Fatalf("expected dtm.noop, got %s", got)
	}
}

func TestSelectDTMProvider_FallsBackToNoop(t *testing.T) {
	// 未知 backend 应回退到 noop
	unknownCfg := &selectorConfigStub{values: map[string]any{"dtm.backend": "unknown"}}
	if got := SelectDTMProvider(unknownCfg).Name(); got != "dtm.noop" {
		t.Fatalf("expected dtm.noop fallback for unknown backend, got %s", got)
	}

	// 空/零值配置默认 noop
	emptyCfg := &selectorConfigStub{values: map[string]any{}}
	if got := SelectDTMProvider(emptyCfg).Name(); got != "dtm.noop" {
		t.Fatalf("expected dtm.noop for empty config, got %s", got)
	}
}

func TestSelectTracingProvider_AcceptsEnabledAndBackends(t *testing.T) {
	// otel 是 contrib 组件，未注册时回退到 noop
	backendCases := []string{"otel", "otlp", "grpc", "http", "stdout"}
	for _, backend := range backendCases {
		cfg := &selectorConfigStub{values: map[string]any{"tracing.backend": backend}}
		if got := SelectTracingProvider(cfg).Name(); got != "tracing.noop" {
			t.Fatalf("backend %s: expected tracing.noop (otel not registered), got %s", backend, got)
		}
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"tracing.backend": "noop"}}
	if got := SelectTracingProvider(noopCfg).Name(); got != "tracing.noop" {
		t.Fatalf("expected tracing.noop, got %s", got)
	}

	// enabled=true 但 contrib 未注册，回退到 noop
	enabledCfg := &selectorConfigStub{values: map[string]any{"tracing.enabled": true}}
	if got := SelectTracingProvider(enabledCfg).Name(); got != "tracing.noop" {
		t.Fatalf("expected tracing.noop (otel not registered), got %s", got)
	}
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
	// token 是 contrib 组件，未注册时回退到 noop
	enabledCfg := &selectorConfigStub{values: map[string]any{"service_auth.enabled": true}}
	if got := SelectServiceAuthProvider(enabledCfg).Name(); got != "serviceauth.noop" {
		t.Fatalf("expected serviceauth.noop (token not registered), got %s", got)
	}

	// mtls 是 contrib 组件，未注册时回退到 noop
	mtlsCfg := &selectorConfigStub{values: map[string]any{"service_auth.mode": "mtls"}}
	if got := SelectServiceAuthProvider(mtlsCfg).Name(); got != "serviceauth.noop" {
		t.Fatalf("expected serviceauth.noop (mtls not registered), got %s", got)
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"service_auth.backend": "noop"}}
	if got := SelectServiceAuthProvider(noopCfg).Name(); got != "serviceauth.noop" {
		t.Fatalf("expected serviceauth.noop, got %s", got)
	}
}

func TestSelectMessageQueueProvider_AcceptsEnabledAndBackend(t *testing.T) {
	// redis MQ 是 contrib 组件，未注册时回退到 noop
	enabledCfg := &selectorConfigStub{values: map[string]any{"message_queue.enabled": true}}
	if got := SelectMessageQueueProvider(enabledCfg).Name(); got != "messagequeue.noop" {
		t.Fatalf("expected messagequeue.noop (redis mq not registered), got %s", got)
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"message_queue.backend": "noop"}}
	if got := SelectMessageQueueProvider(noopCfg).Name(); got != "messagequeue.noop" {
		t.Fatalf("expected messagequeue.noop, got %s", got)
	}
}

// TestSelectMessageQueueProvider_AcceptsContribBackends verifies that when contrib backends
// are specified but not registered, the selector falls back to noop.
//
// TestSelectMessageQueueProvider_AcceptsContribBackends 验证当指定 contrib 后端但未注册时，选择器回退到 noop。
func TestSelectMessageQueueProvider_AcceptsContribBackends(t *testing.T) {
	cases := []string{"kafka", "rabbitmq", "rocketmq", "redis"}
	for _, backend := range cases {
		cfg := &selectorConfigStub{values: map[string]any{"message_queue.backend": backend}}
		if got := SelectMessageQueueProvider(cfg).Name(); got != "messagequeue.noop" {
			t.Errorf("backend=%s: expected messagequeue.noop (contrib not registered), got %s", backend, got)
		}
	}
}

func TestSelectDistributedLockProvider_AcceptsEnabledAndBackend(t *testing.T) {
	// redis dlock 是 contrib 组件，未注册时回退到 noop
	enabledCfg := &selectorConfigStub{values: map[string]any{"distributed_lock.enabled": true}}
	if got := SelectDistributedLockProvider(enabledCfg).Name(); got != "dlock.noop" {
		t.Fatalf("expected dlock.noop (redis not registered), got %s", got)
	}

	noopCfg := &selectorConfigStub{values: map[string]any{"distributed_lock.backend": "noop"}}
	if got := SelectDistributedLockProvider(noopCfg).Name(); got != "dlock.noop" {
		t.Fatalf("expected dlock.noop, got %s", got)
	}
}
