package redis

import (
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "messagequeue.redis", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{contract.MessageQueueKey, contract.MessagePublisherKey, contract.MessageSubscriberKey}, p.Provides())
}
