// Package kafka provides Kafka Subscriber implementation.
// This file implements the MessageSubscriber contract using IBM/sarama SDK.
//
// 本包提供 Kafka Subscriber 实现。
// 本文件使用 IBM/sarama SDK 实现 MessageSubscriber 契约。
package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/IBM/sarama"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// kafkaSubscriber implements MessageSubscriber using IBM/sarama ConsumerGroup.
// Wraps Queue to manage consumer group lifecycle.
//
// kafkaSubscriber 使用 IBM/sarama ConsumerGroup 实现 MessageSubscriber。
// 包装 Queue 以管理 consumer group 生命周期。
type kafkaSubscriber struct {
	queue *Queue
}

// Subscribe creates a consumer for a topic without explicit group.
// This uses simple consumer mode with auto-generated unique group name.
// Kafka requires consumer group for reliable consumption.
//
// Subscribe 创建不带显式分组的消费者。
// 使用简单消费者模式，自动生成唯一组名。
// Kafka 需要 consumer group 才能可靠消费。
func (s *kafkaSubscriber) Subscribe(ctx context.Context, topic string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	// Kafka requires consumer group for reliable consumption
	// For simple subscribe, we create a unique group name
	// Kafka 需要 consumer group 才能可靠消费
	// 对于简单订阅，创建一个唯一组名
	groupID := fmt.Sprintf("single-%s-%d", topic, time.Now().UnixNano())
	return s.SubscribeWithGroup(ctx, topic, groupID, handler)
}

// SubscribeWithGroup creates a consumer group for a topic.
// This is the recommended way to consume from Kafka.
// Implements integrationcontract.MessageSubscriber.SubscribeWithGroup.
//
// SubscribeWithGroup 创建消费者组订阅主题。
// 这是推荐的 Kafka 消费方式。
// 实现 integrationcontract.MessageSubscriber.SubscribeWithGroup。
func (s *kafkaSubscriber) SubscribeWithGroup(ctx context.Context, topic string, group string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	s.queue.mu.Lock()
	defer s.queue.mu.Unlock()
	if s.queue.closed {
		return nil, errors.New("messagequeue.kafka: queue closed")
	}

	// Check if consumer group already exists
	// 检查是否已存在 consumer group
	if existingGroup, ok := s.queue.consumerGroups[group]; ok {
		// Increment reference count for shared group
		// 增加共享 group 的引用计数
		if s.queue.consumerGroupRefs != nil {
			s.queue.consumerGroupRefs[group]++
		}
		// Wrap existing group with new handler
		// 使用新 handler 包装已存在的 group
		return s.wrapConsumerGroup(ctx, group, existingGroup, topic, handler)
	}

	// Create new consumer group
	// 创建新的 consumer group
	saramaCfg := buildSaramaConfig(s.queue.cfg)
	saramaCfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(s.queue.cfg.KafkaBrokers, group, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("messagequeue.kafka: create consumer group failed: %w", err)
	}

	s.queue.consumerGroups[group] = consumerGroup
	if s.queue.consumerGroupRefs == nil {
		s.queue.consumerGroupRefs = make(map[string]int)
	}
	s.queue.consumerGroupRefs[group] = 1

	return s.wrapConsumerGroup(ctx, group, consumerGroup, topic, handler)
}

// wrapConsumerGroup wraps a consumer group with handler logic.
// Starts a background goroutine to consume messages.
// Returns unsubscribe function to cancel consumption and close group.
//
// wrapConsumerGroup 用 handler 逻辑包装 consumer group。
// 启动后台 goroutine 消费消息。
// 返回 unsubscribe 函数用于取消消费和关闭 group。
func (s *kafkaSubscriber) wrapConsumerGroup(ctx context.Context, groupID string, group sarama.ConsumerGroup, topic string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	consumer := &kafkaConsumer{handler: handler}

	// Start consuming in background
	// 在后台启动消费
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := group.Consume(ctx, []string{topic}, consumer)
				if err != nil {
					if errors.Is(err, sarama.ErrClosedConsumerGroup) {
						return
					}
					// Log error and continue
					// 记录错误并继续
					slog.Warn("messagequeue.kafka: consume error", "topic", topic, "group", groupID, "error", err)
					continue
				}
			}
		}
	}()

	return func() error {
		cancel()
		// Only close the consumer group if this is the last subscription using it.
		// Decrement the reference count; close the group only when it reaches zero.
		// 仅当这是使用该 consumer group 的最后一个订阅时才关闭它。
		// 减少引用计数；仅在归零时关闭 group。
		s.queue.mu.Lock()
		if s.queue.consumerGroupRefs != nil {
			s.queue.consumerGroupRefs[groupID]--
			if s.queue.consumerGroupRefs[groupID] <= 0 {
				delete(s.queue.consumerGroupRefs, groupID)
				delete(s.queue.consumerGroups, groupID)
				s.queue.mu.Unlock()
				return group.Close()
			}
		}
		s.queue.mu.Unlock()
		return nil
	}, nil
}

// Consume consumes messages from a topic (treated as queue).
// Kafka does not support direct queue consumption, use SubscribeWithGroup instead.
// Returns error indicating unsupported operation.
//
// ErrConsumeNotSupported is returned when Consume is called on Kafka subscriber.
// Kafka does not support direct queue consumption, use SubscribeWithGroup instead.
var ErrConsumeNotSupported = errors.New("messagequeue.kafka: Consume not supported, use SubscribeWithGroup instead")

// Consume 从 topic（作为 queue）消费消息。
// Kafka 不支持直接 queue 消费，应使用 SubscribeWithGroup。
// 返回错误表示不支持此操作。
func (s *kafkaSubscriber) Consume(ctx context.Context, queue string, handler integrationcontract.MessageHandler) error {
	return ErrConsumeNotSupported
}

// UnsubscribeAll closes all consumer groups.
// Implements integrationcontract.MessageSubscriber.UnsubscribeAll.
//
// UnsubscribeAll 关闭所有 consumer groups。
// 实现 integrationcontract.MessageSubscriber.UnsubscribeAll。
func (s *kafkaSubscriber) UnsubscribeAll() error {
	s.queue.mu.Lock()
	defer s.queue.mu.Unlock()
	for _, group := range s.queue.consumerGroups {
		group.Close()
	}
	s.queue.consumerGroups = make(map[string]sarama.ConsumerGroup)
	return nil
}

// Underlying returns the underlying sarama.Client.
// Allows direct access to native client for advanced operations.
//
// Underlying 返回底层 sarama.Client。
// 允许直接访问原生 client 进行高级操作。
func (s *kafkaSubscriber) Underlying() any {
	if s == nil || s.queue == nil {
		return nil
	}
	return s.queue.client
}

// As attempts to cast the underlying client to the target type.
// Uses internalnative.As for type assertion.
//
// As 尝试将底层 client 转换为目标类型。
// 使用 internalnative.As 进行类型断言。
func (s *kafkaSubscriber) As(target any) bool {
	if s == nil || s.queue == nil || s.queue.client == nil {
		return false
	}
	return As(s.queue.client, target)
}

// NativeSubscriber implements NativeSubscriberProvider interface.
// Returns the first consumer group by sorted key name for deterministic behavior.
//
// NativeSubscriber 实现 NativeSubscriberProvider 接口。
// 按排序后的 key 名返回第一个 consumer group，确保行为确定性。
func (s *kafkaSubscriber) NativeSubscriber() any {
	s.queue.mu.Lock()
	defer s.queue.mu.Unlock()
	keys := make([]string, 0, len(s.queue.consumerGroups))
	for k := range s.queue.consumerGroups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		return s.queue.consumerGroups[k]
	}
	return nil
}

// kafkaConsumer implements sarama.ConsumerGroupHandler.
// Handles Setup, Cleanup, and ConsumeClaim lifecycle callbacks.
//
// kafkaConsumer 实现 sarama.ConsumerGroupHandler。
// 处理 Setup、Cleanup 和 ConsumeClaim 生命周期回调。
type kafkaConsumer struct {
	handler integrationcontract.MessageHandler
}

// Setup is called at the beginning of a new session, before ConsumeClaim.
// No setup needed for this simple handler.
//
// Setup 在新 session 开始时调用，在 ConsumeClaim 之前。
// 此简单 handler 无需 setup。
func (c *kafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is called at the end of a session, once all ConsumeClaim goroutines have exited.
// No cleanup needed for this simple handler.
//
// Cleanup 在 session 结束时调用，所有 ConsumeClaim goroutine 退出后。
// 此简单 handler 无需 cleanup。
func (c *kafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a goroutine to consume messages from the given claim.
// Handles each message, calls user handler, and marks message as consumed on success.
//
// ConsumeClaim 必须启动 goroutine 从给定 claim 消费消息。
// 处理每条消息，调用用户 handler，成功后标记消息已消费。
func (c *kafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &integrationcontract.Message{
			ID:        fmt.Sprintf("%d-%d", msg.Partition, msg.Offset),
			Topic:     msg.Topic,
			Body:      msg.Value,
			Headers:   extractHeaders(msg.Headers),
			Timestamp: msg.Timestamp,
		}
		if err := c.handler(session.Context(), message); err != nil {
			// Handler error, do not mark message as consumed
			// Handler 错误，不标记消息已消费
			continue
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

// extractHeaders converts sarama headers to map[string]string.
//
// extractHeaders 将 sarama headers 转换为 map[string]string。
func extractHeaders(headers []*sarama.RecordHeader) map[string]string {
	result := make(map[string]string)
	for _, h := range headers {
		result[string(h.Key)] = string(h.Value)
	}
	return result
}
