package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestQueueCloseIsIdempotent(t *testing.T) {
	q := &Queue{subs: map[string]context.CancelFunc{}}
	require.NoError(t, q.Close())
	require.NoError(t, q.Close())
}

func TestSubscribeRejectsClosedQueue(t *testing.T) {
	sub := &redisSubscriber{queue: &Queue{closed: true, subs: map[string]context.CancelFunc{}}}
	_, err := sub.Subscribe(context.Background(), "topic", func(context.Context, *contract.Message) error { return nil })
	require.Error(t, err)
	require.Contains(t, err.Error(), "queue closed")
}

func TestSubscribeWithGroupDelegatesToSubscribe(t *testing.T) {
	sub := &redisSubscriber{queue: &Queue{closed: true, subs: map[string]context.CancelFunc{}}}
	_, err := sub.SubscribeWithGroup(context.Background(), "topic", "group", func(context.Context, *contract.Message) error { return nil })
	require.Error(t, err)
	require.Contains(t, err.Error(), "queue closed")
}

func TestUnsubscribeClearsSubscriptions(t *testing.T) {
	called := 0
	sub := &redisSubscriber{queue: &Queue{subs: map[string]context.CancelFunc{
		"a": func() { called++ },
		"b": func() { called++ },
	}}}
	require.NoError(t, sub.Unsubscribe())
	require.Equal(t, 2, called)
	require.Empty(t, sub.queue.subs)
}

func TestPublisherHandlesPublishOptionsWithoutPanic(t *testing.T) {
	pub := &redisPublisher{queue: &Queue{}}
	require.NotNil(t, pub)
	require.NotPanics(t, func() {
		_ = pub.Publish(context.Background(), "topic", []byte("msg"), func(cfg *contract.PublishConfig) {
			cfg.Priority = 1
		})
	})
}

func TestConsumeReturnsContextErrorWhenCancelled(t *testing.T) {
	sub := &redisSubscriber{queue: &Queue{}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := sub.Consume(ctx, "queue", func(context.Context, *contract.Message) error { return errors.New("boom") })
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}
