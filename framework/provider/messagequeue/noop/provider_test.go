package noop

import (
	"context"
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/stretchr/testify/assert"
)

func TestNoopQueue(t *testing.T) {
	queue := &noopQueue{}

	// 测试 Publisher
	publisher := queue.Publisher()
	assert.NotNil(t, publisher)

	// 测试 Subscriber
	subscriber := queue.Subscriber()
	assert.NotNil(t, subscriber)

	// 测试 Close
	err := queue.Close()
	assert.NoError(t, err)
}

func TestNoopPublisher(t *testing.T) {
	publisher := &noopPublisher{}

	// 测试 Publish
	err := publisher.Publish(context.Background(), "test-topic", []byte("message"))
	assert.NoError(t, err)

	// 测试 Send
	err = publisher.Send(context.Background(), "test-queue", []byte("message"))
	assert.NoError(t, err)
}

func TestNoopSubscriber(t *testing.T) {
	subscriber := &noopSubscriber{}

	// 测试 Subscribe
	unsub, err := subscriber.Subscribe(context.Background(), "test-topic", func(ctx context.Context, msg *integrationcontract.Message) error {
		return nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, unsub)

	// 测试 Unsubscribe
	err = subscriber.Unsubscribe()
	assert.NoError(t, err)
}

func TestProvider(t *testing.T) {
	p := NewProvider()

	assert.Equal(t, "messagequeue.noop", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{
		integrationcontract.MessageQueueKey,
		integrationcontract.MessagePublisherKey,
		integrationcontract.MessageSubscriberKey,
	}, p.Provides())
}
