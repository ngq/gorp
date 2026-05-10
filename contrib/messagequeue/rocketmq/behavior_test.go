package rocketmq

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

func TestQueueCloseIsIdempotent(t *testing.T) {
	q := &Queue{
		cfg:    &integrationcontract.MessageQueueConfig{Type: "rocketmq"},
		closed: false,
	}

	// First close should succeed
	err := q.Close()
	require.NoError(t, err)
	require.True(t, q.closed)

	// Second close should also succeed (idempotent)
	err = q.Close()
	require.NoError(t, err)
}

func TestQueueUnderlyingReturnsNilWhenQueueIsNil(t *testing.T) {
	var q *Queue
	require.Nil(t, q.Underlying())
	require.Nil(t, q.NativeMQClient())
	require.False(t, q.As(nil))
}

func TestPublisherReturnsErrorWhenNotInitialized(t *testing.T) {
	p := &rocketmqPublisher{}
	err := p.Publish(context.Background(), "topic", []byte("msg"))
	require.Error(t, err)

	err = p.PublishWithDelay(context.Background(), "topic", []byte("msg"), 0)
	require.Error(t, err)

	err = p.PublishWithPriority(context.Background(), "topic", []byte("msg"), 1)
	require.Error(t, err)
}

func TestPublisherUnderlyingReturnsNilWhenQueueIsNil(t *testing.T) {
	p := &rocketmqPublisher{}
	require.Nil(t, p.Underlying())
	require.Nil(t, p.NativePublisher())
	require.False(t, p.As(nil))
}

func TestSubscriberReturnsErrorWhenQueueClosed(t *testing.T) {
	q := &Queue{
		cfg:    &integrationcontract.MessageQueueConfig{Type: "rocketmq", RocketMQGroupName: "test"},
		closed: true,
	}
	s := &rocketmqSubscriber{queue: q}

	_, err := s.Subscribe(context.Background(), "topic", func(ctx context.Context, msg *integrationcontract.Message) error { return nil })
	require.Error(t, err)
	require.Contains(t, err.Error(), "queue closed")
}

func TestSubscriberConsumeReturnsError(t *testing.T) {
	s := &rocketmqSubscriber{}
	err := s.Consume(context.Background(), "queue", func(ctx context.Context, msg *integrationcontract.Message) error { return nil })
	require.Error(t, err)
	require.Contains(t, err.Error(), "Consume not supported")
}

func TestSubscriberUnderlyingReturnsNilWhenQueueIsNil(t *testing.T) {
	s := &rocketmqSubscriber{}
	require.Nil(t, s.Underlying())
	require.Nil(t, s.NativeSubscriber())
	require.False(t, s.As(nil))
}