package rabbitmq

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

func TestQueueCloseIsIdempotent(t *testing.T) {
	q := &Queue{
		cfg:    &integrationcontract.MessageQueueConfig{Type: "rabbitmq"},
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

func TestPublisherReturnsErrorWhenQueueClosed(t *testing.T) {
	q := &Queue{
		cfg:    &integrationcontract.MessageQueueConfig{Type: "rabbitmq"},
		closed: true,
	}
	p := &rabbitPublisher{queue: q}

	err := p.Publish(context.Background(), "topic", []byte("msg"))
	require.Error(t, err)
}

func TestPublisherUnderlyingReturnsNilWhenQueueIsNil(t *testing.T) {
	p := &rabbitPublisher{queue: nil}
	// NativePublisher 调用 queue.getChannel()，nil queue 会 panic
	// 只测试 Underlying 和 As
	require.Nil(t, p.Underlying())
	require.False(t, p.As(nil))
}

func TestSubscriberReturnsErrorWhenQueueClosed(t *testing.T) {
	q := &Queue{
		cfg:    &integrationcontract.MessageQueueConfig{Type: "rabbitmq"},
		closed: true,
	}
	s := &rabbitSubscriber{queue: q}

	_, err := s.Subscribe(context.Background(), "topic", func(ctx context.Context, msg *integrationcontract.Message) error { return nil })
	require.Error(t, err)
	require.Contains(t, err.Error(), "queue closed")
}

func TestSubscriberConsumeReturnsErrorWhenQueueClosed(t *testing.T) {
	q := &Queue{
		cfg:    &integrationcontract.MessageQueueConfig{Type: "rabbitmq"},
		closed: true,
	}
	s := &rabbitSubscriber{queue: q}

	err := s.Consume(context.Background(), "queue", func(ctx context.Context, msg *integrationcontract.Message) error { return nil })
	require.Error(t, err)
	require.Contains(t, err.Error(), "queue closed")
}

func TestSubscriberUnderlyingReturnsNilWhenQueueIsNil(t *testing.T) {
	s := &rabbitSubscriber{queue: nil}
	// NativeSubscriber 调用 queue.getChannel()，nil queue 会 panic
	// 只测试 Underlying 和 As
	require.Nil(t, s.Underlying())
	require.False(t, s.As(nil))
}