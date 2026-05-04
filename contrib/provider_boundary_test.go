package contrib

import (
	"testing"

	circuitbreakersentinel "github.com/ngq/gorp/contrib/circuitbreaker/sentinel"
	configsourceconsul "github.com/ngq/gorp/contrib/configsource/consul"
	dlockredis "github.com/ngq/gorp/contrib/dlock/redis"
	logzap "github.com/ngq/gorp/contrib/log/zap"
	mqredis "github.com/ngq/gorp/contrib/messagequeue/redis"
	registryconsul "github.com/ngq/gorp/contrib/registry/consul"
	serviceauthmtls "github.com/ngq/gorp/contrib/serviceauth/mtls"
	serviceauthtoken "github.com/ngq/gorp/contrib/serviceauth/token"
	tracingotel "github.com/ngq/gorp/contrib/tracing/otel"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

func TestContribProvidersFollowProviderBoundary(t *testing.T) {
	tests := []struct {
		name        string
		provider    runtimecontract.ServiceProvider
		deferred    bool
		providesKey []string
	}{
		{name: "registry.consul", provider: registryconsul.NewProvider(), deferred: true, providesKey: []string{transportcontract.RPCRegistryKey}},
		{name: "configsource.consul", provider: configsourceconsul.NewProvider(), deferred: true, providesKey: []string{datacontract.ConfigSourceKey}},
		{name: "tracing.otel", provider: tracingotel.NewProvider(), deferred: true, providesKey: []string{observabilitycontract.TracerKey, observabilitycontract.TracerProviderKey}},
		{name: "messagequeue.redis", provider: mqredis.NewProvider(), deferred: true, providesKey: []string{integrationcontract.MessageQueueKey, integrationcontract.MessagePublisherKey, integrationcontract.MessageSubscriberKey}},
		{name: "dlock.redis", provider: dlockredis.NewProvider(), deferred: true, providesKey: []string{datacontract.DistributedLockKey}},
		{name: "serviceauth.token", provider: serviceauthtoken.NewProvider(), deferred: true, providesKey: []string{securitycontract.ServiceAuthKey, securitycontract.ServiceIdentityKey}},
		{name: "serviceauth.mtls", provider: serviceauthmtls.NewProvider(), deferred: true, providesKey: []string{securitycontract.ServiceAuthKey, securitycontract.ServiceIdentityKey}},
		{name: "circuitbreaker.sentinel", provider: circuitbreakersentinel.NewProvider(), deferred: true, providesKey: []string{resiliencecontract.CircuitBreakerKey, resiliencecontract.RateLimiterKey}},
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
