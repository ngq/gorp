// Package noop_test provides unit tests for the message queue noop provider.
//
// 适用场景：
// - 验证消息队列 noop provider 的注册与空操作行为。
package noop

import (
	"context"
	"testing"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	"github.com/stretchr/testify/assert"
)

// TestNoopQueue verifies the noop message queue implementation.
//
// TestNoopQueue 验证消息队列的空操作实现。
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

// TestNoopPublisher verifies the noop message publisher implementation.
//
// TestNoopPublisher 验证消息发布者的空操作实现。
func TestNoopPublisher(t *testing.T) {
	publisher := &noopPublisher{}

	// 测试 Publish
	err := publisher.Publish(context.Background(), "test-topic", []byte("message"))
	assert.NoError(t, err)

	// 测试 Send
	err = publisher.Send(context.Background(), "test-queue", []byte("message"))
	assert.NoError(t, err)
}

// TestNoopSubscriber verifies the noop message subscriber implementation.
//
// TestNoopSubscriber 验证消息订阅者的空操作实现。
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

// TestMessageQueueProvider verifies the message queue provider registration.
//
// TestMessageQueueProvider 验证消息队列服务提供者的注册。
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
