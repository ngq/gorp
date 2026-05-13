// Package bootstrap_test provides integration tests for governance mode and capability provider selection.
//
// 适用场景：
// - 验证 governance mode 的检测、标准化与模式感知选择逻辑。
// - 验证各 provider backend 的 Select 优先級（backend key > config > code disable > default）。
// - 验证 RegisterSelectedMicroserviceProviders 的重载、传播与降级行为。
// - 验证 governance override 链路的优先级顺序。
package bootstrap

import (
	"context"
	"errors"
	"testing"

	"github.com/ngq/gorp/framework"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type reloadingConfigStub struct {
	selectorConfigStub
	reloads           int
	valuesAfterReload map[string]any
	reloadErr         error
}

func (s *reloadingConfigStub) Reload(ctx context.Context) error {
	s.reloads++
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
	if cfg.reloads != 1 {
		t.Fatalf("expected reload once, got %d", cfg.reloads)
	}

	assertBoundKey(t, c, transportcontract.RPCRegistryKey)
	assertBoundKey(t, c, observabilitycontract.TracerKey)
	assertBoundKey(t, c, securitycontract.ServiceAuthKey)
	assertKeyRegistered(t, c, integrationcontract.MessagePublisherKey)
	assertKeyRegistered(t, c, datacontract.DistributedLockKey)
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
		if cfg.reloads != 0 {
			t.Fatalf("backend %s expected no reload, got %d", backend, cfg.reloads)
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
	if err == nil || err.Error() != "reload failed" {
		t.Fatalf("expected reload failed error, got %v", err)
	}
}

func assertKeyRegistered(t *testing.T, c runtimecontract.Container, key string) {
	t.Helper()
	if !c.IsBind(key) {
		t.Fatalf("expected key %s to be registered", key)
	}
}
