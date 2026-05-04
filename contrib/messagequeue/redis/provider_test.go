package redis

import (
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "messagequeue.redis", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{integrationcontract.MessageQueueKey, integrationcontract.MessagePublisherKey, integrationcontract.MessageSubscriberKey}, p.Provides())
}
