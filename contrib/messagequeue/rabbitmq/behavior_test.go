package rabbitmq

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

func newClosedTestQueue() *Queue {
	opts := integrationcontract.ResilientOptions[*amqp.Connection]{
		Dialer: func(ctx context.Context) (*amqp.Connection, error) {
			return nil, nil
		},
	}
	rc, _ := integrationcontract.NewResilientConnection(context.Background(), opts)
	_ = rc.Close()
	return &Queue{
		cfg:     &integrationcontract.MessageQueueConfig{Type: "rabbitmq"},
		resConn: rc,
	}
}

func TestQueueCloseIsIdempotent(t *testing.T) {
	q := newClosedTestQueue()

	// First close should succeed
	err := q.Close()
	require.NoError(t, err)

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
	q := newClosedTestQueue()
	p := &rabbitPublisher{queue: q}

	err := p.Publish(context.Background(), "topic", []byte("msg"))
	require.Error(t, err)
}

func TestPublisherUnderlyingReturnsNilWhenQueueIsNil(t *testing.T) {
	p := &rabbitPublisher{queue: nil}
	require.Nil(t, p.Underlying())
	require.False(t, p.As(nil))
}

func TestSubscriberReturnsErrorWhenQueueClosed(t *testing.T) {
	q := newClosedTestQueue()
	s := &rabbitSubscriber{queue: q}

	_, err := s.Subscribe(context.Background(), "topic", func(ctx context.Context, msg *integrationcontract.Message) error { return nil })
	require.Error(t, err)
}

func TestSubscriberConsumeReturnsErrorWhenQueueClosed(t *testing.T) {
	q := newClosedTestQueue()
	s := &rabbitSubscriber{queue: q}

	err := s.Consume(context.Background(), "queue", func(ctx context.Context, msg *integrationcontract.Message) error { return nil })
	require.Error(t, err)
}

func TestSubscriberUnderlyingReturnsNilWhenQueueIsNil(t *testing.T) {
	s := &rabbitSubscriber{queue: nil}
	require.Nil(t, s.Underlying())
	require.False(t, s.As(nil))
}
