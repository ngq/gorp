// Package noop provides a no-op message queue for monolith scenarios.
// This message queue does nothing: Publish succeeds silently, Subscribe returns empty unsubscribe.
// Note: Message queue is not available in monolith mode.
//
// 空消息队列实现包，用于单体应用场景。
// 此消息队列不执行任何操作：Publish 静默成功，Subscribe 返回空取消订阅。
// 注意：消息队列在单体模式下不可用。
package noop

import (
	"context"
	"errors"
	"time"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers no-op message queue contracts.
//
// Provider 注册空消息队列契约。
type Provider struct{}

// NewProvider creates a new no-op message queue provider instance.
//
// NewProvider 创建新的空消息队列 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "messagequeue.noop".
//
// Name 返回 Provider 名称 "messagequeue.noop"。
func (p *Provider) Name() string { return "messagequeue.noop" }

// IsDefer returns true, message queue can be deferred until first use.
//
// IsDefer 返回 true，消息队列可延迟初始化直到首次使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the message queue contract keys.
//
// Provides 返回消息队列契约键列表。
func (p *Provider) Provides() []string {
	return []string{integrationcontract.MessageQueueKey, integrationcontract.MessagePublisherKey, integrationcontract.MessageSubscriberKey}
}

// Register binds the no-op message queue components to the container.
//
// Register 将空消息队列组件绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.MessageQueueKey, func(c runtimecontract.Container) (any, error) {
		return &noopQueue{}, nil
	}, true)

	c.Bind(integrationcontract.MessagePublisherKey, func(c runtimecontract.Container) (any, error) {
		return &noopPublisher{}, nil
	}, true)

	c.Bind(integrationcontract.MessageSubscriberKey, func(c runtimecontract.Container) (any, error) {
		return &noopSubscriber{}, nil
	}, true)

	return nil
}

// Boot is a no-op for this provider.
//
// Boot 此 Provider 无启动逻辑。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// ErrNoopQueue indicates message queue is not available in monolith mode.
//
// ErrNoopQueue 表示消息队列在单体模式下不可用。
var ErrNoopQueue = errors.New("messagequeue: noop mode, message queue not available in monolith")

// noopQueue implements MessageQueue with no-op behavior.
//
// noopQueue 使用空行为实现 MessageQueue 接口。
type noopQueue struct{}

// Publisher returns a no-op publisher.
//
// Publisher 返回空发布者。
func (q *noopQueue) Publisher() integrationcontract.MessagePublisher {
	return &noopPublisher{}
}

// Subscriber returns a no-op subscriber.
//
// Subscriber 返回空订阅者。
func (q *noopQueue) Subscriber() integrationcontract.MessageSubscriber {
	return &noopSubscriber{}
}

// Close does nothing and returns nil.
//
// Close 不执行任何操作并返回 nil。
func (q *noopQueue) Close() error { return nil }

// noopPublisher implements MessagePublisher with no-op behavior.
//
// noopPublisher 使用空行为实现 MessagePublisher 接口。
type noopPublisher struct{}

// Publish does nothing and returns nil.
//
// Publish 不执行任何操作并返回 nil。
func (p *noopPublisher) Publish(ctx context.Context, topic string, message []byte, options ...integrationcontract.PublishOption) error {
	return nil
}

// PublishWithDelay does nothing and returns nil.
//
// PublishWithDelay 不执行任何操作并返回 nil。
func (p *noopPublisher) PublishWithDelay(ctx context.Context, topic string, message []byte, delay time.Duration) error {
	return nil
}

// PublishWithPriority does nothing and returns nil.
//
// PublishWithPriority 不执行任何操作并返回 nil。
func (p *noopPublisher) PublishWithPriority(ctx context.Context, topic string, message []byte, priority int) error {
	return nil
}

// Send does nothing and returns nil.
//
// Send 不执行任何操作并返回 nil。
func (p *noopPublisher) Send(ctx context.Context, queue string, message []byte, options ...integrationcontract.PublishOption) error {
	return nil
}

// noopSubscriber implements MessageSubscriber with no-op behavior.
//
// noopSubscriber 使用空行为实现 MessageSubscriber 接口。
type noopSubscriber struct{}

// Subscribe returns an empty unsubscribe function.
//
// Subscribe 返回空取消订阅函数。
func (s *noopSubscriber) Subscribe(ctx context.Context, topic string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	return func() error { return nil }, nil
}

// SubscribeWithGroup returns an empty unsubscribe function.
//
// SubscribeWithGroup 返回空取消订阅函数。
func (s *noopSubscriber) SubscribeWithGroup(ctx context.Context, topic string, group string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	return func() error { return nil }, nil
}

// Consume waits for context cancellation.
//
// Consume 等待 context 取消。
func (s *noopSubscriber) Consume(ctx context.Context, queue string, handler integrationcontract.MessageHandler) error {
	<-ctx.Done()
	return ctx.Err()
}

// Unsubscribe does nothing and returns nil.
//
// Unsubscribe 不执行任何操作并返回 nil。
func (s *noopSubscriber) Unsubscribe() error {
	return nil
}