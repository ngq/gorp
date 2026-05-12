// Package bootstrap_test provides unit tests for the bootstrap capability matrix.
//
// 适用场景：
// - 验证引导阶段 capability matrix 的构建与 Feature 检测行为。
package bootstrap

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

type matrixReloadingConfigStub struct {
	selectorConfigStub
	reloads           int
	valuesAfterReload map[string]any
}

func (s *matrixReloadingConfigStub) Reload(ctx context.Context) error {
	s.reloads++
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
	require.Equal(t, 1, cfg.reloads)
	require.True(t, c.IsBind(transportcontract.RPCRegistryKey))
	require.True(t, c.IsBind(observabilitycontract.TracerKey))
	require.True(t, c.IsBind(securitycontract.ServiceAuthKey))
	require.True(t, c.IsBind(integrationcontract.MessagePublisherKey))
	require.True(t, c.IsBind(datacontract.DistributedLockKey))
	require.True(t, c.IsBind(resiliencecontract.CircuitBreakerKey))
}
