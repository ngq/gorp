// Package rocketmq provides RocketMQ-based message queue provider for the gorp framework.
// This provider implements MessageQueue contract with apache/rocketmq-client-go SDK integration.
//
// 本包提供 gorp 框架基于 RocketMQ 的消息队列 provider。
// 本 provider 实现 MessageQueue 契约，集成 apache/rocketmq-client-go SDK。
package rocketmq

import (
	"errors"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider implements runtimecontract.ServiceProvider for RocketMQ message queue.
// Registers MessageQueue, MessagePublisher, and MessageSubscriber services.
//
// Provider 实现 RocketMQ 消息队列的 runtimecontract.ServiceProvider。
// 注册 MessageQueue、MessagePublisher 和 MessageSubscriber 服务。
type Provider struct{}

// NewProvider creates a new RocketMQ message queue provider.
//
// NewProvider 创建新的 RocketMQ 消息队列 provider。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name.
//
// Name 返回 provider 名称。
func (p *Provider) Name() string { return "messagequeue.rocketmq" }

// IsDefer returns true for lazy initialization.
// RocketMQ resources are created on first access, not at boot time.
//
// IsDefer 返回 true 表示延迟初始化。
// RocketMQ 资源在首次访问时创建，而非启动时。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the contract keys this provider satisfies.
// Provides MessageQueue, MessagePublisher, and MessageSubscriber.
//
// Provides 返回此 provider 满足的契约键。
// 提供 MessageQueue、MessagePublisher 和 MessageSubscriber。
func (p *Provider) Provides() []string {
	return []string{
		integrationcontract.MessageQueueKey,
		integrationcontract.MessagePublisherKey,
		integrationcontract.MessageSubscriberKey,
	}
}

// Register binds the message queue services to the container.
// Creates lazy bindings for MessageQueue, Publisher, and Subscriber.
// Each binding extracts configuration and creates Queue on demand.
//
// Register 将消息队列服务绑定到容器。
// 为 MessageQueue、Publisher 和 Subscriber 创建延迟绑定。
// 每个绑定提取配置并按需创建 Queue。
func (p *Provider) Register(c runtimecontract.Container) error {
	// Bind MessageQueue service (lazy singleton)
	c.Bind(integrationcontract.MessageQueueKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getRocketMQConfig(c)
		if err != nil {
			return nil, err
		}
		return NewQueue(cfg)
	}, true)

	// Bind MessagePublisher service (lazy singleton)
	c.Bind(integrationcontract.MessagePublisherKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getRocketMQConfig(c)
		if err != nil {
			return nil, err
		}
		queue, err := NewQueue(cfg)
		if err != nil {
			return nil, err
		}
		return queue.Publisher(), nil
	}, true)

	// Bind MessageSubscriber service (lazy singleton)
	c.Bind(integrationcontract.MessageSubscriberKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getRocketMQConfig(c)
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
// RocketMQ initialization happens on first access.
//
// Boot 对延迟 provider 不执行任何操作。
// RocketMQ 初始化在首次访问时发生。
func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

// getRocketMQConfig extracts RocketMQ configuration from the container.
// Reads configuration from the Config service and builds MessageQueueConfig.
// Provides default values for optional settings.
//
// getRocketMQConfig 从容器提取 RocketMQ 配置。
// 从 Config 服务读取配置并构建 MessageQueueConfig。
// 为可选设置提供默认值。
func getRocketMQConfig(c runtimecontract.Container) (*integrationcontract.MessageQueueConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("messagequeue.rocketmq: invalid config service")
	}

	// Build default configuration
	mqCfg := &integrationcontract.MessageQueueConfig{
		Type:                 "rocketmq",
		MaxRetry:             3,
		RetryDelay:           time.Second,
		Timeout:              5 * time.Second,
		ConsumerBuffer:       100,
		RocketMQNamesrvAddr:  "localhost:9876",
		RocketMQGroupName:    "default-group",
		RocketMQInstanceName: "",
		RocketMQRetryTimes:   3,
	}

	// Nameserver address (required)
	// Multiple addresses can be separated by semicolon
	if namesrv := cfg.GetString("message_queue.rocketmq.namesrv_addr"); namesrv != "" {
		mqCfg.RocketMQNamesrvAddr = namesrv
	}

	// Optional configuration
	if groupName := cfg.GetString("message_queue.rocketmq.group_name"); groupName != "" {
		mqCfg.RocketMQGroupName = groupName
	}
	if instanceName := cfg.GetString("message_queue.rocketmq.instance_name"); instanceName != "" {
		mqCfg.RocketMQInstanceName = instanceName
	}
	if retryTimes := cfg.GetInt("message_queue.rocketmq.retry_times"); retryTimes > 0 {
		mqCfg.RocketMQRetryTimes = retryTimes
	}
	if enableTLS := cfg.GetBool("message_queue.rocketmq.enable_tls"); enableTLS {
		mqCfg.RocketMQEnableTLS = enableTLS
	}

	// Common configuration
	if maxRetry := cfg.GetInt("message_queue.max_retry"); maxRetry > 0 {
		mqCfg.MaxRetry = maxRetry
	}
	if timeoutMs := cfg.GetInt("message_queue.timeout_ms"); timeoutMs > 0 {
		mqCfg.Timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	return mqCfg, nil
}