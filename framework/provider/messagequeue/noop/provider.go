package noop

import (
	"context"
	"errors"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop 消息队列实现。
//
// 中文说明：
// - 单体项目默认使用此 provider；
// - 不引入任何外部依赖；
// - 所有发布/订阅操作为空实现。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "messagequeue.noop" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.MessageQueueKey, contract.MessagePublisherKey, contract.MessageSubscriberKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.MessageQueueKey, func(c contract.Container) (any, error) {
		return &noopQueue{}, nil
	}, true)

	c.Bind(contract.MessagePublisherKey, func(c contract.Container) (any, error) {
		return &noopPublisher{}, nil
	}, true)

	c.Bind(contract.MessageSubscriberKey, func(c contract.Container) (any, error) {
		return &noopSubscriber{}, nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// ErrNoopQueue 表示 noop 消息队列不支持实际操作。
var ErrNoopQueue = errors.New("messagequeue: noop mode, message queue not available in monolith")

// noopQueue 是 MessageQueue 的空实现。
type noopQueue struct{}

func (q *noopQueue) Publisher() contract.MessagePublisher {
	return &noopPublisher{}
}

func (q *noopQueue) Subscriber() contract.MessageSubscriber {
	return &noopSubscriber{}
}

func (q *noopQueue) Close() error { return nil }

// noopPublisher 是 MessagePublisher 的空实现。
type noopPublisher struct{}

func (p *noopPublisher) Publish(ctx context.Context, topic string, message []byte, options ...contract.PublishOption) error {
	// 空操作，忽略消息
	return nil
}

func (p *noopPublisher) PublishWithDelay(ctx context.Context, topic string, message []byte, delay time.Duration) error {
	return nil
}

func (p *noopPublisher) PublishWithPriority(ctx context.Context, topic string, message []byte, priority int) error {
	return nil
}

func (p *noopPublisher) Send(ctx context.Context, queue string, message []byte, options ...contract.PublishOption) error {
	return nil
}

// noopSubscriber 是 MessageSubscriber 的空实现。
type noopSubscriber struct{}

func (s *noopSubscriber) Subscribe(ctx context.Context, topic string, handler contract.MessageHandler) (contract.UnsubscribeFunc, error) {
	// 返回空取消函数
	return func() error { return nil }, nil
}

func (s *noopSubscriber) SubscribeWithGroup(ctx context.Context, topic string, group string, handler contract.MessageHandler) (contract.UnsubscribeFunc, error) {
	return func() error { return nil }, nil
}

func (s *noopSubscriber) Consume(ctx context.Context, queue string, handler contract.MessageHandler) error {
	// 阻塞直到上下文取消
	<-ctx.Done()
	return ctx.Err()
}

func (s *noopSubscriber) Unsubscribe() error {
	return nil
}