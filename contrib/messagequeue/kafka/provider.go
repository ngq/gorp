// Package kafka provides Kafka-based message queue provider for the gorp framework.
// This provider implements MessageQueue contract with IBM/sarama SDK integration.
//
// 本包提供 gorp 框架基于 Kafka 的消息队列 provider。
// 本 provider 实现 MessageQueue 契约，集成 IBM/sarama SDK。
package kafka

import (
	"errors"
	"strings"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider implements runtimecontract.ServiceProvider for Kafka message queue.
//
// Provider 实现 runtimecontract.ServiceProvider 用于 Kafka 消息队列。
type Provider struct{}

// NewProvider creates a new Kafka message queue provider.
//
// NewProvider 创建新的 Kafka 消息队列 provider。
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name.
//
// Name 返回 provider 名称。
func (p *Provider) Name() string {
	return "messagequeue.kafka"
}

// IsDefer returns true for lazy initialization.
//
// IsDefer 返回 true 表示延迟初始化。
func (p *Provider) IsDefer() bool {
	return true
}

// Provides returns the contract keys this provider satisfies.
//
// Provides 返回此 provider 满足的契约键。
func (p *Provider) Provides() []string {
	return []string{
		integrationcontract.MessageQueueKey,
		integrationcontract.MessagePublisherKey,
		integrationcontract.MessageSubscriberKey,
	}
}

// Register binds the message queue services to the container.
// Binds MessageQueue, MessagePublisher, and MessageSubscriber separately.
//
// Register 将消息队列服务绑定到容器。
// 分别绑定 MessageQueue、MessagePublisher 和 MessageSubscriber。
func (p *Provider) Register(c runtimecontract.Container) error {
	// Bind MessageQueue service
	// 绑定 MessageQueue 服务
	c.Bind(integrationcontract.MessageQueueKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getKafkaConfig(c)
		if err != nil {
			return nil, err
		}
		return NewQueue(cfg)
	}, true)

	// Bind MessagePublisher service
	// 绑定 MessagePublisher 服务
	c.Bind(integrationcontract.MessagePublisherKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getKafkaConfig(c)
		if err != nil {
			return nil, err
		}
		queue, err := NewQueue(cfg)
		if err != nil {
			return nil, err
		}
		return queue.Publisher(), nil
	}, true)

	// Bind MessageSubscriber service
	// 绑定 MessageSubscriber 服务
	c.Bind(integrationcontract.MessageSubscriberKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getKafkaConfig(c)
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
//
// Boot 对延迟 provider 无操作。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// getKafkaConfig extracts Kafka configuration from the container.
// Reads from datacontract.Config service and builds MessageQueueConfig.
//
// getKafkaConfig 从容器提取 Kafka 配置。
// 从 datacontract.Config 服务读取并构建 MessageQueueConfig。
func getKafkaConfig(c runtimecontract.Container) (*integrationcontract.MessageQueueConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("messagequeue.kafka: invalid config service")
	}

	// Default configuration
	// 默认配置
	mqCfg := &integrationcontract.MessageQueueConfig{
		Type:               "kafka",
		MaxRetry:           3,
		RetryDelay:         time.Second,
		Timeout:            5 * time.Second,
		ConsumerBuffer:     100,
		KafkaVersion:       "2.8.0",
		KafkaRequiredACKs:  -1, // WaitForAll
		KafkaPartitioner:   "hash",
		KafkaMaxMessageBytes: 1000000, // 1MB
		KafkaFlushFrequency: 0,
	}

	// Brokers configuration (required)
	// Brokers 配置（必需）
	brokers := cfg.GetString("message_queue.kafka.brokers")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	// Split by comma
	// 按逗号分割
	mqCfg.KafkaBrokers = strings.Split(brokers, ",")

	// Optional configuration
	// 可选配置
	if groupID := cfg.GetString("message_queue.kafka.group_id"); groupID != "" {
		mqCfg.KafkaGroupID = groupID
	}
	if clientID := cfg.GetString("message_queue.kafka.client_id"); clientID != "" {
		mqCfg.KafkaClientID = clientID
	}
	if version := cfg.GetString("message_queue.kafka.version"); version != "" {
		mqCfg.KafkaVersion = version
	}
	if compression := cfg.GetString("message_queue.kafka.compression"); compression != "" {
		mqCfg.KafkaCompression = compression
	}
	if partitioner := cfg.GetString("message_queue.kafka.partitioner"); partitioner != "" {
		mqCfg.KafkaPartitioner = partitioner
	}
	if acks := cfg.GetInt("message_queue.kafka.required_acks"); acks != 0 {
		mqCfg.KafkaRequiredACKs = acks
	}
	if maxBytes := cfg.GetInt("message_queue.kafka.max_message_bytes"); maxBytes > 0 {
		mqCfg.KafkaMaxMessageBytes = maxBytes
	}
	if flushMs := cfg.GetInt("message_queue.kafka.flush_frequency_ms"); flushMs > 0 {
		mqCfg.KafkaFlushFrequency = time.Duration(flushMs) * time.Millisecond
	}
	if enableTLS := cfg.GetBool("message_queue.kafka.enable_tls"); enableTLS {
		mqCfg.KafkaEnableTLS = enableTLS
		mqCfg.KafkaTLSCertFile = cfg.GetString("message_queue.kafka.tls_cert_file")
		mqCfg.KafkaTLSKeyFile = cfg.GetString("message_queue.kafka.tls_key_file")
		mqCfg.KafkaTLSCACertFile = cfg.GetString("message_queue.kafka.tls_ca_cert_file")
	}

	// Common configuration
	// 通用配置
	if maxRetry := cfg.GetInt("message_queue.max_retry"); maxRetry > 0 {
		mqCfg.MaxRetry = maxRetry
	}
	if timeoutMs := cfg.GetInt("message_queue.timeout_ms"); timeoutMs > 0 {
		mqCfg.Timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	return mqCfg, nil
}