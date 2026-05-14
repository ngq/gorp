package basemq_test

import (
	"context"
	"testing"
	"time"

	"github.com/ngq/gorp/contrib/internal/basemq"
	"github.com/ngq/gorp/framework/container"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

type mockQueue struct {
	closed     bool
	publisher  integrationcontract.MessagePublisher
	subscriber integrationcontract.MessageSubscriber
}

func (q *mockQueue) Publisher() integrationcontract.MessagePublisher   { return q.publisher }
func (q *mockQueue) Subscriber() integrationcontract.MessageSubscriber { return q.subscriber }
func (q *mockQueue) Close() error {
	q.closed = true
	return nil
}

type mockPublisher struct{}

func (mockPublisher) Publish(_ context.Context, _ string, _ []byte, _ ...integrationcontract.PublishOption) error {
	return nil
}
func (mockPublisher) PublishWithDelay(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return nil
}
func (mockPublisher) PublishWithPriority(_ context.Context, _ string, _ []byte, _ int) error {
	return nil
}
func (mockPublisher) Send(_ context.Context, _ string, _ []byte, _ ...integrationcontract.PublishOption) error {
	return nil
}

type mockSubscriber struct{}

func (mockSubscriber) Subscribe(_ context.Context, _ string, _ integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	return nil, nil
}
func (mockSubscriber) SubscribeWithGroup(_ context.Context, _, _ string, _ integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	return nil, nil
}
func (mockSubscriber) Consume(_ context.Context, _ string, _ integrationcontract.MessageHandler) error {
	return nil
}
func (mockSubscriber) Unsubscribe() error { return nil }

func TestBaseMQProvider_RegisterDerivesPublisherFromQueue(t *testing.T) {
	var queue *mockQueue
	cfg := &integrationcontract.MessageQueueConfig{Type: "mock"}

	base := &basemq.BaseMQProvider{
		NameStr: "messagequeue.mock",
		GetConfig: func(c runtimecontract.Container) (*integrationcontract.MessageQueueConfig, error) {
			return cfg, nil
		},
		NewQueue: func(c *integrationcontract.MessageQueueConfig) (integrationcontract.MessageQueue, error) {
			queue = &mockQueue{
				publisher:  &mockPublisher{},
				subscriber: &mockSubscriber{},
			}
			return queue, nil
		},
	}

	c := container.New()
	err := base.Register(c)
	require.NoError(t, err)

	mqAny, err := c.Make(integrationcontract.MessageQueueKey)
	require.NoError(t, err)
	mq, ok := mqAny.(integrationcontract.MessageQueue)
	require.True(t, ok)
	require.NotNil(t, mq)

	pubAny, err := c.Make(integrationcontract.MessagePublisherKey)
	require.NoError(t, err)
	require.NotNil(t, pubAny)

	subAny, err := c.Make(integrationcontract.MessageSubscriberKey)
	require.NoError(t, err)
	require.NotNil(t, subAny)
}

func TestBaseMQProvider_DestroyClosesQueue(t *testing.T) {
	var queue *mockQueue

	base := &basemq.BaseMQProvider{
		NameStr: "messagequeue.mock",
		GetConfig: func(c runtimecontract.Container) (*integrationcontract.MessageQueueConfig, error) {
			return &integrationcontract.MessageQueueConfig{Type: "mock"}, nil
		},
		NewQueue: func(c *integrationcontract.MessageQueueConfig) (integrationcontract.MessageQueue, error) {
			queue = &mockQueue{
				publisher:  &mockPublisher{},
				subscriber: &mockSubscriber{},
			}
			return queue, nil
		},
	}

	c := container.New()
	err := base.Register(c)
	require.NoError(t, err)

	_, err = c.Make(integrationcontract.MessageQueueKey)
	require.NoError(t, err)

	err = c.Destroy()
	require.NoError(t, err)
	require.True(t, queue.closed)
}

func TestBaseMQProvider_ProvidesCorrectKeys(t *testing.T) {
	base := &basemq.BaseMQProvider{NameStr: "messagequeue.mock"}
	provides := base.Provides()
	require.Equal(t, []string{
		integrationcontract.MessageQueueKey,
		integrationcontract.MessagePublisherKey,
		integrationcontract.MessageSubscriberKey,
	}, provides)
}

func TestBaseMQProvider_Metadata(t *testing.T) {
	base := &basemq.BaseMQProvider{NameStr: "messagequeue.mock"}
	require.Equal(t, "messagequeue.mock", base.Name())
	require.True(t, base.IsDefer())
}
