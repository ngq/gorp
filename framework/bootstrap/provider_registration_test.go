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
	"errors"
	"strings"
	"testing"

	"github.com/ngq/gorp/framework"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type reloadingConfigStub struct {
	selectorConfigStub
	reloadCalled      bool
	valuesAfterReload map[string]any
	reloadErr         error
}

func (s *reloadingConfigStub) Reload(ctx context.Context) error {
	s.reloadCalled = true
	if s.reloadErr != nil {
		return s.reloadErr
	}
	for key, value := range s.valuesAfterReload {
		s.values[key] = value
	}
	return nil
}

// testRemoteConfigSourceProvider 是一个测试用远程配置源 provider。
// 注册到 "consul" factory，使 reload 测试能在不引入真实 contrib 的情况下
// 走"非 local/noop → reload"路径。
type testRemoteConfigSourceProvider struct{}

func (p *testRemoteConfigSourceProvider) Name() string    { return "configsource.testremote" }
func (p *testRemoteConfigSourceProvider) IsDefer() bool   { return false }
func (p *testRemoteConfigSourceProvider) Provides() []string {
	return []string{datacontract.ConfigSourceKey}
}
func (p *testRemoteConfigSourceProvider) DependsOn() []string { return nil }
func (p *testRemoteConfigSourceProvider) Boot(runtimecontract.Container) error {
	return nil
}
func (p *testRemoteConfigSourceProvider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ConfigSourceKey, func(runtimecontract.Container) (any, error) {
		return &testConfigSource{}, nil
	}, true)
	return nil
}

type testConfigSource struct{}

func (s *testConfigSource) Load(ctx context.Context) (map[string]any, error) {
	return map[string]any{}, nil
}
func (s *testConfigSource) Get(ctx context.Context, key string) (any, error) {
	return nil, nil
}
func (s *testConfigSource) Set(ctx context.Context, key string, value any) error {
	return nil
}
func (s *testConfigSource) Watch(ctx context.Context, key string) (datacontract.ConfigWatcher, error) {
	return nil, nil
}
func (s *testConfigSource) Close() error { return nil }

// registerTestConsulConfigSource 把测试 config source 注册到 "consul" 后端。
func registerTestConsulConfigSource() {
	RegisterConfigSourceProviderFactory("consul", func() runtimecontract.ServiceProvider {
		return &testRemoteConfigSourceProvider{}
	})
}

// =============================================================================
// RegisterSelectedMicroserviceProviders 注册与重载行为
// =============================================================================

func TestRegisterSelectedMicroserviceProviders_SkipsWithoutConfigBinding(t *testing.T) {
	app := framework.NewApplication()
	c := app.Container()
	if err := RegisterSelectedMicroserviceProviders(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRegisterSelectedMicroserviceProviders_ReloadsRemoteConfigSourceBeforeSelectingOthers(t *testing.T) {
	// 注册测试用 "consul" 配置源，使 reload 路径生效。
	registerTestConsulConfigSource()

	app := framework.NewApplication()
	c := app.Container()
	cfg := &reloadingConfigStub{
		selectorConfigStub: selectorConfigStub{values: map[string]any{
			"configsource.backend": "consul",
		}},
		valuesAfterReload: map[string]any{
			// 其余能力显式 noop，聚焦验证"远程配置源触发 reload"
			"discovery.backend":        "noop",
			"tracing.backend":         "noop",
			"service_auth.backend":    "noop",
			"message_queue.backend":   "noop",
			"distributed_lock.backend": "noop",
			"circuit_breaker.backend": "noop",
		},
	}
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return cfg, nil
	}, true)

	if err := RegisterSelectedMicroserviceProviders(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	// consul 配置源已注册，应触发 reload
	if !cfg.reloadCalled {
		t.Fatalf("expected reload for remote config source, but reload was NOT called")
	}

	// RPCRegistry 是 framework 内建能力
	assertBoundKey(t, c, transportcontract.RPCRegistryKey)
}

func TestRegisterSelectedMicroserviceProviders_DoesNotReloadLocalOrNoopConfigSource(t *testing.T) {
	for _, backend := range []string{"local", "noop"} {
		app := framework.NewApplication()
		c := app.Container()
		cfg := &reloadingConfigStub{selectorConfigStub: selectorConfigStub{values: map[string]any{
			"configsource.backend": backend,
		}}}
		c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
			return cfg, nil
		}, true)
		if err := RegisterSelectedMicroserviceProviders(c); err != nil {
			t.Fatalf("backend %s expected nil error, got %v", backend, err)
		}
		if cfg.reloadCalled {
			t.Fatalf("backend %s expected no reload, got reload called", backend)
		}
	}
}

func TestRegisterSelectedMicroserviceProviders_PropagatesReloadError(t *testing.T) {
	// 注册测试用 "consul" 配置源，使 reload 错误能够被触发并传播。
	registerTestConsulConfigSource()

	app := framework.NewApplication()
	c := app.Container()
	cfg := &reloadingConfigStub{
		selectorConfigStub: selectorConfigStub{values: map[string]any{"configsource.backend": "consul"}},
		reloadErr:          errors.New("reload failed"),
	}
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return cfg, nil
	}, true)

	err := RegisterSelectedMicroserviceProviders(c)
	// 远程配置源 reload 失败应传播错误
	if err == nil {
		t.Fatal("expected reload error to propagate, got nil")
	}
	if !strings.Contains(err.Error(), "reload failed") {
		t.Fatalf("expected reload failed in error, got %v", err)
	}
}

func assertKeyRegistered(t *testing.T, c runtimecontract.Container, key string) {
	t.Helper()
	if !c.IsBind(key) {
		t.Fatalf("expected key %s to be registered", key)
	}
}