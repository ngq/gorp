// Package bootstrap_test provides unit tests for the bootstrap capability matrix.
//
// 适用场景：
// - 验证引导阶段 capability matrix 的构建与 Feature 检测行为。
//
// 注意：contrib 组件现在是独立模块，这些测试验证框架选择逻辑，
// 当 contrib provider 未注册时，会回退到 noop。
package bootstrap

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

type matrixReloadingConfigStub struct {
	selectorConfigStub
	reloadCalled      bool
	valuesAfterReload map[string]any
}

func (s *matrixReloadingConfigStub) Reload(ctx context.Context) error {
	s.reloadCalled = true
	for key, value := range s.valuesAfterReload {
		s.values[key] = value
	}
	return nil
}

// TestRegisterSelectedMicroserviceProviders_ProductionMainlineMatrix verifies that a
// production-mainline config referencing contrib backends without registering them
// fail-fasts (per capability decision: unknown backend → run failure, not silent noop).
//
// TestRegisterSelectedMicroserviceProviders_ProductionMainlineMatrix 验证生产主线配置
// 引用了 contrib 后端但未注册时按 fail-fast 决策失败（未知后端→启动失败，而非静默 noop）。
func TestRegisterSelectedMicroserviceProviders_ProductionMainlineMatrix(t *testing.T) {
	app := framework.NewApplication()
	c := app.Container()
	cfg := &matrixReloadingConfigStub{
		selectorConfigStub: selectorConfigStub{values: map[string]any{
			"configsource.backend": "etcd",
		}},
		valuesAfterReload: map[string]any{
			"discovery.backend":        "etcd",
			"tracing.backend":          "stdout",
			"service_auth.mode":        "token",
			"message_queue.backend":    "redis",
			"distributed_lock.backend": "redis",
			"circuit_breaker.backend":  "sentinel",
		},
	}
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
		return cfg, nil
	}, true)

	// 生产主线引用了 etcd/redis/sentinel/token 等 contrib 后端，但测试二进制未
	// 注册这些 contrib —— 按 fail-fast 决策应返回错误，而不是静默降级 noop。
	err := RegisterSelectedMicroserviceProviders(c)
	require.Error(t, err, "expected fail-fast for unregistered contrib backends in production mainline")
	require.Contains(t, err.Error(), "not registered")

	// 由于配置源 etcd 未注册，reload 不应被触发（reload 依赖 configsource 注册成功）。
	require.False(t, cfg.reloadCalled)
}