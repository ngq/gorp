package rabbitmq

import (
	"testing"

	"github.com/stretchr/testify/require"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()

	require.Equal(t, "messagequeue.rabbitmq", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{
		integrationcontract.MessageQueueKey,
		integrationcontract.MessagePublisherKey,
		integrationcontract.MessageSubscriberKey,
	}, p.Provides())
}

func TestProviderNewQueue(t *testing.T) {
	// Skip if no RabbitMQ available - this is a unit test for contract only
	// Integration tests require real RabbitMQ instance
	t.Skip("requires RabbitMQ instance for integration test")
}