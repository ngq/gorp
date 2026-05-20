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
type Provider struct {
	BaseMQProvider
}

// NewProvider creates a new Kafka message queue provider.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "messagequeue.kafka"
	p.GetConfig = getKafkaConfig
	p.NewQueue = func(cfg *integrationcontract.MessageQueueConfig) (integrationcontract.MessageQueue, error) {
		return NewQueue(cfg)
	}
	return p
}

// getKafkaConfig extracts Kafka configuration from the container.
func getKafkaConfig(c runtimecontract.Container) (*integrationcontract.MessageQueueConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("messagequeue.kafka: invalid config service")
	}

	mqCfg := &integrationcontract.MessageQueueConfig{
		Type:                 "kafka",
		MaxRetry:             3,
		RetryDelay:           time.Second,
		Timeout:              5 * time.Second,
		ConsumerBuffer:       100,
		KafkaVersion:         "2.8.0",
		KafkaRequiredACKs:    -1,
		KafkaPartitioner:     "hash",
		KafkaMaxMessageBytes: 1000000,
		KafkaFlushFrequency:  0,
	}

	brokers := cfg.GetString("message_queue.kafka.brokers")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	mqCfg.KafkaBrokers = strings.Split(brokers, ",")

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

	if maxRetry := cfg.GetInt("message_queue.max_retry"); maxRetry > 0 {
		mqCfg.MaxRetry = maxRetry
	}
	if timeoutMs := cfg.GetInt("message_queue.timeout_ms"); timeoutMs > 0 {
		mqCfg.Timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	return mqCfg, nil
}
