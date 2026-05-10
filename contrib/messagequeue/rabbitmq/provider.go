// Package rabbitmq provides RabbitMQ-based message queue provider for the gorp framework.
// This provider implements MessageQueue contract with amqp091-go SDK integration.
//
// 本包提供 gorp 框架基于 RabbitMQ 的消息队列 provider。
// 本 provider 实现 MessageQueue 契约，集成 amqp091-go SDK。
package rabbitmq

import (
	"errors"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider implements runtimecontract.ServiceProvider for RabbitMQ message queue.
type Provider struct{}

// NewProvider creates a new RabbitMQ message queue provider.
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name.
func (p *Provider) Name() string { return "messagequeue.rabbitmq" }

// IsDefer returns true for lazy initialization.
func (p *Provider) IsDefer() bool { return true }

// Provides returns the contract keys this provider satisfies.
func (p *Provider) Provides() []string {
	return []string{
		integrationcontract.MessageQueueKey,
		integrationcontract.MessagePublisherKey,
		integrationcontract.MessageSubscriberKey,
	}
}

// Register binds the message queue services to the container.
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.MessageQueueKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getRabbitMQConfig(c)
		if err != nil {
			return nil, err
		}
		return NewQueue(cfg)
	}, true)

	c.Bind(integrationcontract.MessagePublisherKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getRabbitMQConfig(c)
		if err != nil {
			return nil, err
		}
		queue, err := NewQueue(cfg)
		if err != nil {
			return nil, err
		}
		return queue.Publisher(), nil
	}, true)

	c.Bind(integrationcontract.MessageSubscriberKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getRabbitMQConfig(c)
		if err != nil {
			return nil, err
		}
		queue, err := NewQueue(cfg)
		if err != nil {
			return nil, err
		}
		return queue.Subscriber(), nil
	}, true)

	return nil
}

// Boot does nothing for lazy providers.
func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

// getRabbitMQConfig extracts RabbitMQ configuration from the container.
//
// getRabbitMQConfig 从容器提取 RabbitMQ 配置。
func getRabbitMQConfig(c runtimecontract.Container) (*integrationcontract.MessageQueueConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("messagequeue.rabbitmq: invalid config service")
	}

	mqCfg := &integrationcontract.MessageQueueConfig{
		Type:              "rabbitmq",
		MaxRetry:          3,
		RetryDelay:        time.Second,
		Timeout:           5 * time.Second,
		ConsumerBuffer:    10,
		RabbitMQURL:       "amqp://guest:guest@localhost:5672/",
		RabbitMQVHost:     "/",
		RabbitMQExchange:  "",
		RabbitMQExchangeType: "topic",
		RabbitMQPrefetch:  10,
	}

	// URL configuration (required)
	if url := cfg.GetString("message_queue.rabbitmq.url"); url != "" {
		mqCfg.RabbitMQURL = url
	}

	// Optional configuration
	if vhost := cfg.GetString("message_queue.rabbitmq.vhost"); vhost != "" {
		mqCfg.RabbitMQVHost = vhost
	}
	if exchange := cfg.GetString("message_queue.rabbitmq.exchange"); exchange != "" {
		mqCfg.RabbitMQExchange = exchange
	}
	if exchangeType := cfg.GetString("message_queue.rabbitmq.exchange_type"); exchangeType != "" {
		mqCfg.RabbitMQExchangeType = exchangeType
	}
	if queuePrefix := cfg.GetString("message_queue.rabbitmq.queue_prefix"); queuePrefix != "" {
		mqCfg.RabbitMQQueuePrefix = queuePrefix
	}
	if prefetch := cfg.GetInt("message_queue.rabbitmq.prefetch"); prefetch > 0 {
		mqCfg.RabbitMQPrefetch = prefetch
	}
	if enableTLS := cfg.GetBool("message_queue.rabbitmq.enable_tls"); enableTLS {
		mqCfg.RabbitMQEnableTLS = enableTLS
	}

	if maxRetry := cfg.GetInt("message_queue.max_retry"); maxRetry > 0 {
		mqCfg.MaxRetry = maxRetry
	}
	if timeoutMs := cfg.GetInt("message_queue.timeout_ms"); timeoutMs > 0 {
		mqCfg.Timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	return mqCfg, nil
}