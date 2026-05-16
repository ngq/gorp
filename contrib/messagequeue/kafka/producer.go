// Package kafka provides Kafka Publisher implementation.
// This file implements the MessagePublisher contract using IBM/sarama SDK.
//
// 本包提供 Kafka Publisher 实现。
// 本文件使用 IBM/sarama SDK 实现 MessagePublisher 契约。
package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/IBM/sarama"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// kafkaPublisher implements MessagePublisher using IBM/sarama.
// Wraps Queue to access sync producer for message publishing.
//
// kafkaPublisher 使用 IBM/sarama 实现 MessagePublisher。
// 包装 Queue 以访问 sync producer 进行消息发布。
type kafkaPublisher struct {
	queue *Queue
}

// Publish sends a message to a Kafka topic.
// Implements integrationcontract.MessagePublisher.Publish.
// Applies headers from PublishConfig if provided.
//
// Publish 将消息发送到 Kafka topic。
// 实现 integrationcontract.MessagePublisher.Publish。
// 如果提供了 Headers，从 PublishConfig 应用。
func (p *kafkaPublisher) Publish(ctx context.Context, topic string, message []byte, options ...integrationcontract.PublishOption) error {
	if p == nil || p.queue == nil || p.queue.syncProducer == nil {
		return errors.New("messagequeue.kafka: producer not initialized")
	}

	cfg := &integrationcontract.PublishConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	// Apply headers if provided
	// 如果提供了 headers，应用到消息
	if len(cfg.Headers) > 0 {
		for k, v := range cfg.Headers {
			msg.Headers = append(msg.Headers, sarama.RecordHeader{
				Key:   sarama.ByteEncoder(k),
				Value: sarama.ByteEncoder(v),
			})
		}
	}

	_, _, err := p.queue.syncProducer.SendMessage(msg)
	return err
}

// PublishWithDelay returns error as Kafka does not natively support delayed messages.
// Users should implement delayed delivery using external scheduling or Redis-based delay queue.
//
// PublishWithDelay 返回错误，因为 Kafka 不原生支持延迟消息。
// 用户应使用外置调度或 Redis 延迟队列实现延迟投递。
func (p *kafkaPublisher) PublishWithDelay(ctx context.Context, topic string, message []byte, delay time.Duration) error {
	return errors.New("messagequeue.kafka: delayed messages not supported, use external scheduler or Redis delay queue")
}

// PublishWithPriority sends a message with partition key for priority-like behavior.
// Kafka does not have native priority, but partition key can route messages consistently.
// Uses priority number as partition key and optionally directs to specific partition.
//
// PublishWithPriority 发送带分区键的消息实现类似优先级的路由。
// Kafka 无原生优先级，但分区键可一致路由消息。
// 使用 priority 数作为分区键，可选直接发送到特定分区。
func (p *kafkaPublisher) PublishWithPriority(ctx context.Context, topic string, message []byte, priority int) error {
	if p == nil || p.queue == nil || p.queue.syncProducer == nil {
		return errors.New("messagequeue.kafka: producer not initialized")
	}

	// Use priority as a message key for consistent routing within the topic.
	// Do NOT use priority as a partition number — partition count varies by cluster
	// and priority > partition count would cause errors.
	// 将 priority 作为消息键实现 topic 内的一致路由。
	// 不将 priority 作为分区号——分区数因集群而异，priority > 分区数会报错。
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(fmt.Sprintf("priority-%d", priority)),
		Value: sarama.ByteEncoder(message),
	}

	_, _, err := p.queue.syncProducer.SendMessage(msg)
	return err
}

// Send sends a message treating the queue name as topic.
// Implements integrationcontract.MessagePublisher.Send.
// In Kafka, queue concept maps to topic directly.
//
// Send 将消息发送到队列（作为 topic）。
// 实现 integrationcontract.MessagePublisher.Send。
// 在 Kafka 中，queue 概念直接映射到 topic。
func (p *kafkaPublisher) Send(ctx context.Context, queue string, message []byte, options ...integrationcontract.PublishOption) error {
	return p.Publish(ctx, queue, message, options...)
}

// Underlying returns the underlying sarama.SyncProducer.
// Allows direct access to native producer for advanced operations.
//
// Underlying 返回底层 sarama.SyncProducer。
// 允许直接访问原生 producer 进行高级操作。
func (p *kafkaPublisher) Underlying() any {
	if p == nil || p.queue == nil {
		return nil
	}
	return p.queue.syncProducer
}

// As attempts to cast the underlying producer to the target type.
// Uses internalnative.As for type assertion.
//
// As 尝试将底层 producer 转换为目标类型。
// 使用 internalnative.As 进行类型断言。
func (p *kafkaPublisher) As(target any) bool {
	if p == nil || p.queue == nil || p.queue.syncProducer == nil {
		return false
	}
	return internalnative.As(p.queue.syncProducer, target)
}

// NativePublisher implements NativePublisherProvider interface.
// Returns the underlying sarama.SyncProducer for publishing.
//
// NativePublisher 实现 NativePublisherProvider 接口。
// 返回底层 sarama.SyncProducer 用于发布。
func (p *kafkaPublisher) NativePublisher() any {
	return p.Underlying()
}
