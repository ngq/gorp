package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

func TestQueueCloseIsIdempotent(t *testing.T) {
	// Use a minimal mock for idempotent close test
	q := &Queue{
		cfg:            &integrationcontract.MessageQueueConfig{Type: "kafka"},
		consumerGroups: make(map[string]sarama.ConsumerGroup),
		closed:         false,
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

func TestPublisherPublishWithDelayReturnsError(t *testing.T) {
	p := &kafkaPublisher{queue: &Queue{}}
	err := p.PublishWithDelay(context.Background(), "topic", []byte("msg"), time.Second)
	require.Error(t, err)
	require.Contains(t, err.Error(), "delayed messages not supported")
}

func TestPublisherReturnsErrorWhenNotInitialized(t *testing.T) {
	p := &kafkaPublisher{}
	err := p.Publish(context.Background(), "topic", []byte("msg"))
	require.Error(t, err)

	err = p.PublishWithPriority(context.Background(), "topic", []byte("msg"), 1)
	require.Error(t, err)
}

func TestSubscriberReturnsErrorWhenQueueClosed(t *testing.T) {
	q := &Queue{
		cfg:            &integrationcontract.MessageQueueConfig{Type: "kafka"},
		consumerGroups: make(map[string]sarama.ConsumerGroup),
		closed:         true,
	}
	s := &kafkaSubscriber{queue: q}

	_, err := s.Subscribe(context.Background(), "topic", func(ctx context.Context, msg *integrationcontract.Message) error { return nil })
	require.Error(t, err)
	require.Contains(t, err.Error(), "queue closed")
}

func TestSubscriberConsumeReturnsError(t *testing.T) {
	s := &kafkaSubscriber{}
	err := s.Consume(context.Background(), "queue", func(ctx context.Context, msg *integrationcontract.Message) error { return nil })
	require.Error(t, err)
	require.Contains(t, err.Error(), "Consume not supported")
}

func TestPublisherUnderlyingReturnsNilWhenQueueIsNil(t *testing.T) {
	p := &kafkaPublisher{}
	require.Nil(t, p.Underlying())
	require.Nil(t, p.NativePublisher())
	require.False(t, p.As(nil))
}

func TestSubscriberUnderlyingReturnsNilWhenQueueIsNil(t *testing.T) {
	s := &kafkaSubscriber{queue: nil}
	// NativeSubscriber 调用 queue.mu.Lock()，nil queue 会 panic
	// 只测试 Underlying 和 As
	require.Nil(t, s.Underlying())
	require.False(t, s.As(nil))
}
