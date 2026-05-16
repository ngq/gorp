// Package noop_test provides unit tests for the message queue noop provider.
//
// 适用场景：
// - 验证消息队列 noop provider 的注册与空操作行为。
// - 覆盖 Publisher / Subscriber 全部方法与 Container 绑定验证。
package noop

import (
	"context"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/container"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNoopQueue verifies the noop message queue implementation.
//
// TestNoopQueue 验证消息队列的空操作实现。
func TestNoopQueue(t *testing.T) {
	queue := &noopQueue{}

	publisher := queue.Publisher()
	assert.NotNil(t, publisher)

	subscriber := queue.Subscriber()
	assert.NotNil(t, subscriber)

	err := queue.Close()
	assert.NoError(t, err)
}

// TestNoopPublisher verifies the noop message publisher implementation, including all publish variants.
//
// TestNoopPublisher 验证消息发布者的空操作实现，包含所有发布变体方法。
func TestNoopPublisher(t *testing.T) {
	publisher := &noopPublisher{}

	err := publisher.Publish(context.Background(), "test-topic", []byte("message"))
	assert.NoError(t, err)

	err = publisher.PublishWithDelay(context.Background(), "test-topic", []byte("message"), 5*time.Second)
	assert.NoError(t, err)

	err = publisher.PublishWithPriority(context.Background(), "test-topic", []byte("message"), 1)
	assert.NoError(t, err)

	err = publisher.Send(context.Background(), "test-queue", []byte("message"))
	assert.NoError(t, err)
}

// TestNoopSubscriber verifies the noop message subscriber implementation, including group subscription and consume.
//
// TestNoopSubscriber 验证消息订阅者的空操作实现，包含分组订阅与消费方法。
func TestNoopSubscriber(t *testing.T) {
	subscriber := &noopSubscriber{}

	unsub, err := subscriber.Subscribe(context.Background(), "test-topic", func(ctx context.Context, msg *integrationcontract.Message) error {
		return nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, unsub)

	unsub, err = subscriber.SubscribeWithGroup(context.Background(), "test-topic", "test-group", func(ctx context.Context, msg *integrationcontract.Message) error {
		return nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, unsub)

	err = subscriber.UnsubscribeAll()
	assert.NoError(t, err)
}

// TestNoopSubscriber_ConsumeRespectsContext verifies that Consume exits when context is cancelled.
//
// TestNoopSubscriber_ConsumeRespectsContext 验证 Consume 在 context 取消时退出。
func TestNoopSubscriber_ConsumeRespectsContext(t *testing.T) {
	subscriber := &noopSubscriber{}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := subscriber.Consume(ctx, "test-queue", func(ctx context.Context, msg *integrationcontract.Message) error {
		return nil
	})
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

// TestProviderContract verifies the message queue provider metadata and Provides keys.
//
// TestProviderContract 验证消息队列 provider 元信息与 Provides 契约。
func TestProviderContract(t *testing.T) {
	p := NewProvider()

	assert.Equal(t, "messagequeue.noop", p.Name())
	assert.True(t, p.IsDefer())
	assert.ElementsMatch(t, []string{
		integrationcontract.MessageQueueKey,
		integrationcontract.MessagePublisherKey,
		integrationcontract.MessageSubscriberKey,
	}, p.Provides())
}

// TestProvider_RegisterAndResolve verifies the noop provider can be registered into a container and all three keys resolve.
//
// TestProvider_RegisterAndResolve 验证 noop provider 可注册到容器并正确解析三个契约 key。
func TestProvider_RegisterAndResolve(t *testing.T) {
	c := container.New()
	require.NoError(t, c.RegisterProvider(NewProvider()))

	queue, err := c.Make(integrationcontract.MessageQueueKey)
	require.NoError(t, err)
	_, ok := queue.(integrationcontract.MessageQueue)
	assert.True(t, ok, "expected MessageQueue interface from container")

	pub, err := c.Make(integrationcontract.MessagePublisherKey)
	require.NoError(t, err)
	_, ok = pub.(integrationcontract.MessagePublisher)
	assert.True(t, ok, "expected MessagePublisher interface from container")

	sub, err := c.Make(integrationcontract.MessageSubscriberKey)
	require.NoError(t, err)
	_, ok = sub.(integrationcontract.MessageSubscriber)
	assert.True(t, ok, "expected MessageSubscriber interface from container")
}
