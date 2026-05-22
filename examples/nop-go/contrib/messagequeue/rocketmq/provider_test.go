package rocketmq

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()

	require.Equal(t, "messagequeue.rocketmq", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{
		integrationcontract.MessageQueueKey,
		integrationcontract.MessagePublisherKey,
		integrationcontract.MessageSubscriberKey,
	}, p.Provides())
}

func TestProviderNewQueue(t *testing.T) {
	// Skip if no RocketMQ available - this is a unit test for contract only
	// Integration tests require real RocketMQ instance
	t.Skip("requires RocketMQ instance for integration test")
}

func TestParseDelayLevel(t *testing.T) {
	require.Equal(t, 1, parseDelayLevel(time.Millisecond))
	require.Equal(t, 1, parseDelayLevel(time.Second))
	require.Equal(t, 2, parseDelayLevel(3*time.Second))
	require.Equal(t, 3, parseDelayLevel(8*time.Second))
	require.Equal(t, 4, parseDelayLevel(20*time.Second))
	require.Equal(t, 5, parseDelayLevel(45*time.Second))
	require.Equal(t, 6, parseDelayLevel(90*time.Second))
	require.Equal(t, 16, parseDelayLevel(30*time.Minute))
	require.Equal(t, 18, parseDelayLevel(2*time.Hour))
}
