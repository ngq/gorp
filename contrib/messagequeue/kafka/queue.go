// Package kafka provides Kafka Queue implementation.
// This file implements the MessageQueue core structure with client and producer management.
//
// 本包提供 Kafka Queue 实现。
// 本文件实现 MessageQueue 核心结构，包含 client 和 producer 管理。
package kafka

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/IBM/sarama"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// Queue implements integrationcontract.MessageQueue using IBM/sarama SDK.
// Manages sarama client, sync producer, and consumer groups.
//
// Queue 使用 IBM/sarama SDK 实现 integrationcontract.MessageQueue。
// 管理 sarama client、sync producer 和 consumer groups。
type Queue struct {
	cfg               *integrationcontract.MessageQueueConfig
	client            sarama.Client
	syncProducer      sarama.SyncProducer
	asyncProducer     sarama.AsyncProducer // optional, for high-throughput scenarios
	consumerGroups    map[string]sarama.ConsumerGroup
	consumerGroupRefs map[string]int
	mu                sync.Mutex
	closed            bool
}

// NewQueue creates a new Kafka Queue instance.
// Establishes sarama client and creates sync producer from client.
//
// NewQueue 创建新的 Kafka Queue 实例。
// 建立 sarama client 并从 client 创建 sync producer。
func NewQueue(cfg *integrationcontract.MessageQueueConfig) (*Queue, error) {
	saramaCfg := buildSaramaConfig(cfg)

	// Create Sarama client
	// 创建 Sarama client
	client, err := sarama.NewClient(cfg.KafkaBrokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("messagequeue.kafka: create client failed: %w", err)
	}

	// Create sync producer from client
	// Required for SyncProducer: Producer.Return.Successes must be true
	// 从 client 创建 sync producer
	// SyncProducer 需要 Producer.Return.Successes = true
	syncProducer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("messagequeue.kafka: create sync producer failed: %w", err)
	}

	return &Queue{
		cfg:            cfg,
		client:         client,
		syncProducer:   syncProducer,
		consumerGroups: make(map[string]sarama.ConsumerGroup),
	}, nil
}

// Publisher returns a Kafka-based MessagePublisher.
//
// Publisher 返回基于 Kafka 的 MessagePublisher。
func (q *Queue) Publisher() integrationcontract.MessagePublisher {
	return &kafkaPublisher{queue: q}
}

// Subscriber returns a Kafka-based MessageSubscriber.
//
// Subscriber 返回基于 Kafka 的 MessageSubscriber。
func (q *Queue) Subscriber() integrationcontract.MessageSubscriber {
	return &kafkaSubscriber{queue: q}
}

// Close closes all Kafka resources.
// Implements integrationcontract.MessageQueue.Close.
// Closes consumer groups, producers, and client in order.
// Logs warnings for non-fatal close errors and returns the first critical error.
//
// Close 关闭所有 Kafka 资源。
// 实现 integrationcontract.MessageQueue.Close。
// 按顺序关闭 consumer groups、producers 和 client。
// 对非致命关闭错误记录警告，返回第一个关键错误。
func (q *Queue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return nil
	}
	q.closed = true

	var errs []error

	// Close consumer groups first
	// 先关闭 consumer groups
	for _, group := range q.consumerGroups {
		if err := group.Close(); err != nil {
			slog.Warn("messagequeue.kafka: close consumer group failed", "error", err)
			errs = append(errs, err)
		}
	}
	q.consumerGroups = nil

	// Close producers
	// 关闭 producers
	if q.syncProducer != nil {
		if err := q.syncProducer.Close(); err != nil {
			slog.Warn("messagequeue.kafka: close sync producer failed", "error", err)
			errs = append(errs, err)
		}
	}
	if q.asyncProducer != nil {
		if err := q.asyncProducer.Close(); err != nil {
			slog.Warn("messagequeue.kafka: close async producer failed", "error", err)
			errs = append(errs, err)
		}
	}

	// Close client last
	// 最后关闭 client
	if q.client != nil {
		if err := q.client.Close(); err != nil {
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	}
	return errors.Join(errs...)
}

// Underlying returns the underlying sarama.Client for advanced usage.
// This allows users to access native Kafka SDK capabilities such as
// custom partitioners, transactions, admin operations, etc.
//
// Underlying 返回底层 sarama.Client 供高级使用。
// 这允许用户访问原生 Kafka SDK 能力，如自定义分区器、事务、管理操作等。
func (q *Queue) Underlying() any {
	if q == nil {
		return nil
	}
	return q.client
}

// As attempts to cast the underlying sarama.Client to the target type.
// Uses internalnative.As for type assertion.
//
// As 尝试将底层 sarama.Client 转换为目标类型。
// 使用 internalnative.As 进行类型断言。
func (q *Queue) As(target any) bool {
	if q == nil || q.client == nil {
		return false
	}
	return internalnative.As(q.client, target)
}

// NativeMQClient implements NativeMQClientProvider interface.
// Returns the underlying sarama.Client.
//
// NativeMQClient 实现 NativeMQClientProvider 接口。
// 返回底层 sarama.Client。
func (q *Queue) NativeMQClient() any {
	return q.Underlying()
}

// NativeSyncProducer returns the underlying sarama.SyncProducer.
// Useful for users who want to use producer directly for advanced operations.
//
// NativeSyncProducer 返回底层 sarama.SyncProducer。
// 用于用户直接使用 producer 进行高级操作。
func (q *Queue) NativeSyncProducer() sarama.SyncProducer {
	return q.syncProducer
}
