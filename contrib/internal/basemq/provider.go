// Package basemq provides a base message queue provider template.
// Concrete MQ providers embed BaseMQProvider and only supply
// provider-specific config extraction and queue construction logic.
// This eliminates the structural duplication across all MQ providers
// and fixes the P0 resource leak where Publisher/Subscriber each
// created independent connections that could never be closed.
//
// basemq 包提供消息队列 provider 基础模板。
// 具体 MQ provider 内嵌 BaseMQProvider，只需提供差异化的配置提取和队列构造逻辑。
// 这消除了所有 MQ provider 的结构性重复，
// 并修复了 Publisher/Subscriber 各自创建无法 Close 的独立连接的 P0 资源泄漏。
package basemq

import (
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// BaseMQProvider eliminates structural duplication across MQ providers.
// Concrete providers embed this struct and only supply Name, GetConfig, and NewQueue.
//
// BaseMQProvider 消除 MQ provider 的结构性重复。
// 具体 provider 内嵌此结构体，只需提供 Name、GetConfig、NewQueue 三项差异化逻辑。
type BaseMQProvider struct {
	// NameStr is the provider identifier, e.g. "messagequeue.kafka".
	NameStr string

	// GetConfig extracts provider-specific configuration from the container.
	GetConfig func(c runtimecontract.Container) (*integrationcontract.MessageQueueConfig, error)

	// NewQueue creates a Queue instance from the given config.
	NewQueue func(cfg *integrationcontract.MessageQueueConfig) (integrationcontract.MessageQueue, error)
}

// Name returns the provider identifier.
func (p *BaseMQProvider) Name() string { return p.NameStr }

// IsDefer returns true for lazy initialization.
func (p *BaseMQProvider) IsDefer() bool { return true }

// Provides returns the contract keys this provider satisfies.
func (p *BaseMQProvider) Provides() []string {
	return []string{
		integrationcontract.MessageQueueKey,
		integrationcontract.MessagePublisherKey,
		integrationcontract.MessageSubscriberKey,
	}
}

// Register binds MessageQueue, MessagePublisher, and MessageSubscriber to the container.
// MessageQueue is the single singleton that holds the underlying connection.
// Publisher and Subscriber are derived from it via c.Make, eliminating the
// resource leak where each used to create its own independent connection.
// The Queue's Close method is registered with the container's Destroy lifecycle.
//
// Register 将 MessageQueue、MessagePublisher、MessageSubscriber 绑定到容器。
// MessageQueue 是持有底层连接的唯一单例。
// Publisher 和 Subscriber 通过 c.Make 从 MessageQueue 派生，
// 消除了各自创建独立连接的资源泄漏。
// Queue 的 Close 方法注册到容器的 Destroy 生命周期。
func (p *BaseMQProvider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.MessageQueueKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := p.GetConfig(c)
		if err != nil {
			return nil, err
		}
		q, err := p.NewQueue(cfg)
		if err != nil {
			return nil, err
		}
		c.RegisterCloser(integrationcontract.MessageQueueKey, q)
		return q, nil
	}, true)

	c.Bind(integrationcontract.MessagePublisherKey, func(c runtimecontract.Container) (any, error) {
		mq, err := c.Make(integrationcontract.MessageQueueKey)
		if err != nil {
			return nil, err
		}
		return mq.(integrationcontract.MessageQueue).Publisher(), nil
	}, true)

	c.Bind(integrationcontract.MessageSubscriberKey, func(c runtimecontract.Container) (any, error) {
		mq, err := c.Make(integrationcontract.MessageQueueKey)
		if err != nil {
			return nil, err
		}
		return mq.(integrationcontract.MessageQueue).Subscriber(), nil
	}, true)

	return nil
}

// Boot does nothing for lazy providers.
func (p *BaseMQProvider) Boot(c runtimecontract.Container) error { return nil }
