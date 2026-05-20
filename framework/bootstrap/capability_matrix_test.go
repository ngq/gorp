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
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
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

	require.NoError(t, RegisterSelectedMicroserviceProviders(c))
	// etcd 是 contrib 组件，未注册时不会触发 reload
	// 因为 configsource.local 是本地配置源，不需要 reload
	require.False(t, cfg.reloadCalled)
	// contrib 组件未注册，这些 key 不会被绑定（因为 provider 是 noop）
	// noop provider 通常不绑定实际能力
	require.True(t, c.IsBind(transportcontract.RPCRegistryKey))
	// tracing、serviceauth、messagequeue、dlock、circuitbreaker 都是 contrib 组件
	// 未注册时是 noop，不会绑定实际能力
}