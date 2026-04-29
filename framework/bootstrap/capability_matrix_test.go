package bootstrap

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework"
	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type matrixReloadingConfigStub struct {
	selectorConfigStub
	reloads int
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
	c.Bind(contract.ConfigKey, func(contract.Container) (any, error) {
		return cfg, nil
	}, true)

	require.NoError(t, RegisterSelectedMicroserviceProviders(c))
	require.Equal(t, 1, cfg.reloads)
	require.True(t, c.IsBind(contract.RPCRegistryKey))
	require.True(t, c.IsBind(contract.TracerKey))
	require.True(t, c.IsBind(contract.ServiceAuthKey))
	require.True(t, c.IsBind(contract.MessagePublisherKey))
	require.True(t, c.IsBind(contract.DistributedLockKey))
	require.True(t, c.IsBind(contract.CircuitBreakerKey))
}
