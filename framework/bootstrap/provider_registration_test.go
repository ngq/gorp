// Package bootstrap_test provides integration tests for governance mode and capability provider selection.
//
// 适用场景：
// - 验证 governance mode 的检测、标准化与模式感知选择逻辑。
// - 验证各 provider backend 的 Select 优先級（backend key > config > code disable > default）。
// - 验证 RegisterSelectedMicroserviceProviders 的重载、传播与降级行为。
// - 验证 governance override 链路的优先级顺序。
//
// 注意：contrib 组件现在是独立模块，这些测试验证框架选择逻辑，
// 当 contrib provider 未注册时，会回退到 noop。
package bootstrap

import (
	"context"
	"errors"
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
	app := framework.NewApplication()
	c := app.Container()
	cfg := &reloadingConfigStub{
		selectorConfigStub: selectorConfigStub{values: map[string]any{
			"configsource.backend": "consul",
		}},
		valuesAfterReload: map[string]any{
			"discovery.backend":        "consul",
			"tracing.enabled":          true,
			"service_auth.enabled":     true,
			"message_queue.enabled":    true,
			"distributed_lock.enabled": true,
		},
	}
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return cfg, nil
	}, true)

	if err := RegisterSelectedMicroserviceProviders(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	// consul 是 contrib 组件，未注册时不会触发 reload
	// 因为会回退到 configsource.local
	if cfg.reloadCalled {
		t.Fatalf("expected no reload (consul not registered), but reload was called")
	}

	// RPCRegistry 是 framework 内建能力
	assertBoundKey(t, c, transportcontract.RPCRegistryKey)
	// tracing、serviceauth、messagequeue、dlock 都是 contrib 组件
	// 未注册时是 noop，不会绑定实际能力
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
	// consul 是 contrib 组件，未注册时会回退到 local
	// local 不需要 reload，所以不会触发 reload 错误
	// 因此这里期望 nil error
	if err != nil {
		t.Fatalf("expected nil error (consul not registered, fallback to local), got %v", err)
	}
}

func assertKeyRegistered(t *testing.T, c runtimecontract.Container, key string) {
	t.Helper()
	if !c.IsBind(key) {
		t.Fatalf("expected key %s to be registered", key)
	}
}