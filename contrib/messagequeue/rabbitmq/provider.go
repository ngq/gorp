// Package rabbitmq provides RabbitMQ-based message queue provider for the gorp framework.
// This provider implements MessageQueue contract with amqp091-go SDK integration.
//
// 本包提供 gorp 框架基于 RabbitMQ 的消息队列 provider。
// 本 provider 实现 MessageQueue 契约，集成 amqp091-go SDK。
package rabbitmq

import (
	"errors"
	"time"

	"github.com/ngq/gorp/contrib/internal/basemq"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider implements runtimecontract.ServiceProvider for RabbitMQ message queue.
type Provider struct {
	basemq.BaseMQProvider
}

// NewProvider creates a new RabbitMQ message queue provider.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "messagequeue.rabbitmq"
	p.GetConfig = getRabbitMQConfig
	p.NewQueue = func(cfg *integrationcontract.MessageQueueConfig) (integrationcontract.MessageQueue, error) {
		return NewQueue(cfg)
	}
	return p
}

// getRabbitMQConfig extracts RabbitMQ configuration from the container.
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
		Type:                 "rabbitmq",
		MaxRetry:             3,
		RetryDelay:           time.Second,
		Timeout:              5 * time.Second,
		ConsumerBuffer:       10,
		RabbitMQURL:          "amqp://guest:guest@localhost:5672/",
		RabbitMQVHost:        "/",
		RabbitMQExchange:     "",
		RabbitMQExchangeType: "topic",
		RabbitMQPrefetch:     10,
	}

	if url := cfg.GetString("message_queue.rabbitmq.url"); url != "" {
		mqCfg.RabbitMQURL = url
	}
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
