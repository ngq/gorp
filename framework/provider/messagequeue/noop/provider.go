package noop

import (
	"context"
	"errors"
	"time"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "messagequeue.noop" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{integrationcontract.MessageQueueKey, integrationcontract.MessagePublisherKey, integrationcontract.MessageSubscriberKey}
}

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

func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

var ErrNoopQueue = errors.New("messagequeue: noop mode, message queue not available in monolith")

type noopQueue struct{}

func (q *noopQueue) Publisher() integrationcontract.MessagePublisher {
	return &noopPublisher{}
}

func (q *noopQueue) Subscriber() integrationcontract.MessageSubscriber {
	return &noopSubscriber{}
}

func (q *noopQueue) Close() error { return nil }

type noopPublisher struct{}

func (p *noopPublisher) Publish(ctx context.Context, topic string, message []byte, options ...integrationcontract.PublishOption) error {
	return nil
}

func (p *noopPublisher) PublishWithDelay(ctx context.Context, topic string, message []byte, delay time.Duration) error {
	return nil
}

func (p *noopPublisher) PublishWithPriority(ctx context.Context, topic string, message []byte, priority int) error {
	return nil
}

func (p *noopPublisher) Send(ctx context.Context, queue string, message []byte, options ...integrationcontract.PublishOption) error {
	return nil
}

type noopSubscriber struct{}

func (s *noopSubscriber) Subscribe(ctx context.Context, topic string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	return func() error { return nil }, nil
}

func (s *noopSubscriber) SubscribeWithGroup(ctx context.Context, topic string, group string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	return func() error { return nil }, nil
}

func (s *noopSubscriber) Consume(ctx context.Context, queue string, handler integrationcontract.MessageHandler) error {
	<-ctx.Done()
	return ctx.Err()
}

func (s *noopSubscriber) Unsubscribe() error {
	return nil
}
