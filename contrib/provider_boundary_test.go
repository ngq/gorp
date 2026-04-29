package contrib

import (
	"testing"

	circuitbreakersentinel "github.com/ngq/gorp/contrib/circuitbreaker/sentinel"
	configsourceconsul "github.com/ngq/gorp/contrib/configsource/consul"
	dlockredis "github.com/ngq/gorp/contrib/dlock/redis"
	logzap "github.com/ngq/gorp/contrib/log/zap"
	mqredis "github.com/ngq/gorp/contrib/mq/redis"
	registryconsul "github.com/ngq/gorp/contrib/registry/consul"
	serviceauthmtls "github.com/ngq/gorp/contrib/serviceauth/mtls"
	serviceauthtoken "github.com/ngq/gorp/contrib/serviceauth/token"
	tracingotel "github.com/ngq/gorp/contrib/tracing/otel"
	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestContribProvidersFollowProviderBoundary(t *testing.T) {
	tests := []struct {
		name        string
		provider    contract.ServiceProvider
		deferred    bool
		providesKey []string
	}{
		{name: "registry.consul", provider: registryconsul.NewProvider(), deferred: true, providesKey: []string{contract.RPCRegistryKey}},
		{name: "configsource.consul", provider: configsourceconsul.NewProvider(), deferred: true, providesKey: []string{contract.ConfigSourceKey}},
		{name: "tracing.otel", provider: tracingotel.NewProvider(), deferred: true, providesKey: []string{contract.TracerKey, contract.TracerProviderKey}},
		{name: "messagequeue.redis", provider: mqredis.NewProvider(), deferred: true, providesKey: []string{contract.MessageQueueKey, contract.MessagePublisherKey, contract.MessageSubscriberKey}},
		{name: "dlock.redis", provider: dlockredis.NewProvider(), deferred: true, providesKey: []string{contract.DistributedLockKey}},
		{name: "serviceauth.token", provider: serviceauthtoken.NewProvider(), deferred: true, providesKey: []string{contract.ServiceAuthKey, contract.ServiceIdentityKey}},
		{name: "serviceauth.mtls", provider: serviceauthmtls.NewProvider(), deferred: true, providesKey: []string{contract.ServiceAuthKey, contract.ServiceIdentityKey}},
		{name: "circuitbreaker.sentinel", provider: circuitbreakersentinel.NewProvider(), deferred: true, providesKey: []string{contract.CircuitBreakerKey, contract.RateLimiterKey}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEmpty(t, tt.provider.Name())
			require.Equal(t, tt.deferred, tt.provider.IsDefer())
			require.ElementsMatch(t, tt.providesKey, tt.provider.Provides())
		})
	}
}

func TestContribLogBackendAvailability(t *testing.T) {
	logger, err := logzap.New("info", "console")
	require.NoError(t, err)
	require.NotNil(t, logger)
}
