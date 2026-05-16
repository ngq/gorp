// Package rocketmq provides RocketMQ Subscriber implementation.
// This file implements the MessageSubscriber contract using rocketmq-client-go SDK.
//
// 本包提供 RocketMQ Subscriber 实现。
// 本文件使用 rocketmq-client-go SDK 实现 MessageSubscriber 契约。
package rocketmq

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// rocketmqSubscriber implements MessageSubscriber using rocketmq-client-go.
// Wraps the Queue to access configuration and create consumers.
//
// rocketmqSubscriber 使用 rocketmq-client-go 实现 MessageSubscriber。
// 包装 Queue 以访问配置并创建 consumers。
type rocketmqSubscriber struct {
	queue *Queue
}

// Subscribe creates a PushConsumer for a topic.
// Implements integrationcontract.MessageSubscriber.Subscribe.
// Uses the default group name from configuration.
//
// Subscribe 创建 topic 的 PushConsumer。
// 实现 integrationcontract.MessageSubscriber.Subscribe。
// 使用配置中的默认 group name。
func (s *rocketmqSubscriber) Subscribe(ctx context.Context, topic string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	return s.SubscribeWithGroup(ctx, topic, s.queue.cfg.RocketMQGroupName, handler)
}

// SubscribeWithGroup creates a PushConsumer with a specific consumer group.
// This is the recommended way to consume from RocketMQ.
// Creates a dedicated consumer instance for this subscription.
//
// SubscribeWithGroup 创建特定消费者组的 PushConsumer。
// 这是推荐的 RocketMQ 消费方式。
// 为此订阅创建专用的 consumer 实例。
func (s *rocketmqSubscriber) SubscribeWithGroup(ctx context.Context, topic string, group string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	s.queue.mu.Lock()
	defer s.queue.mu.Unlock()

	if s.queue.closed {
		return nil, errors.New("messagequeue.rocketmq: queue closed")
	}

	// Create new consumer for this subscription
	c, err := s.queue.createConsumer(group)
	if err != nil {
		return nil, err
	}

	// Subscribe to topic with message handler
	err = c.Subscribe(topic, consumer.MessageSelector{},
		func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for _, msg := range msgs {
				// Convert RocketMQ message to contract Message
				m := &integrationcontract.Message{
					ID:         msg.MsgId,
					Topic:      msg.Topic,
					Body:       msg.Body,
					Headers:    msg.GetProperties(),
					Timestamp:  time.UnixMilli(msg.StoreTimestamp),
					RetryCount: int(msg.ReconsumeTimes),
				}
				// Invoke user handler
				if err := handler(ctx, m); err != nil {
					// Return retry later on handler error
					return consumer.ConsumeRetryLater, err
				}
			}
			return consumer.ConsumeSuccess, nil
		},
	)
	if err != nil {
		c.Shutdown()
		return nil, fmt.Errorf("messagequeue.rocketmq: subscribe failed: %w", err)
	}

	// Start consumer
	err = c.Start()
	if err != nil {
		c.Shutdown()
		return nil, fmt.Errorf("messagequeue.rocketmq: start consumer failed: %w", err)
	}

	// Store consumer reference in the consumers map (keyed by group)
	s.queue.mu.Lock()
	s.queue.consumers[group] = c
	s.queue.mu.Unlock()

	// Return unsubscribe function
	return func() error {
		s.queue.mu.Lock()
		delete(s.queue.consumers, group)
		s.queue.mu.Unlock()
		return c.Shutdown()
	}, nil
}

// Consume consumes messages from a topic.
// RocketMQ does not support this pattern directly; use SubscribeWithGroup instead.
// This method returns an error explaining the correct usage.
//
// ErrConsumeNotSupported is returned when Consume is called on RocketMQ subscriber.
// RocketMQ does not support direct queue consumption, use SubscribeWithGroup instead.
var ErrConsumeNotSupported = errors.New("messagequeue.rocketmq: Consume not supported, use SubscribeWithGroup instead")

// Consume 从 topic 消费消息。
// RocketMQ 不直接支持此模式；应使用 SubscribeWithGroup。
// 该方法返回错误，说明正确用法。
func (s *rocketmqSubscriber) Consume(ctx context.Context, queue string, handler integrationcontract.MessageHandler) error {
	return ErrConsumeNotSupported
}

// Unsubscribe shuts down all consumers.
// Implements integrationcontract.MessageSubscriber.Unsubscribe.
//
// UnsubscribeAll 关闭所有 consumer。
// 实现 integrationcontract.MessageSubscriber.UnsubscribeAll。
func (s *rocketmqSubscriber) UnsubscribeAll() error {
	s.queue.mu.Lock()
	defer s.queue.mu.Unlock()
	var firstErr error
	for group, c := range s.queue.consumers {
		if err := c.Shutdown(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(s.queue.consumers, group)
	}
	return firstErr
}

// Underlying returns the first underlying rocketmq.PushConsumer.
// This allows users to access native SDK capabilities directly.
// Note: Returns the first consumer found; for specific consumer access,
// use NativeSubscriber().
//
// Underlying 返回第一个底层 rocketmq.PushConsumer。
// 这允许用户直接访问原生 SDK 能力。
// 注意：返回找到的第一个 consumer；如需访问特定 consumer 请用 NativeSubscriber()。
func (s *rocketmqSubscriber) Underlying() any {
	if s == nil || s.queue == nil {
		return nil
	}
	for _, c := range s.queue.consumers {
		return c
	}
	return nil
}

// As attempts to cast the underlying consumer to the target type.
// Uses the internal native.As helper for type casting.
//
// As 尝试将底层 consumer 转换为目标类型。
// 使用内部 native.As 辅助函数进行类型转换。
func (s *rocketmqSubscriber) As(target any) bool {
	if s == nil || s.queue == nil {
		return false
	}
	for _, c := range s.queue.consumers {
		return internalnative.As(c, target)
	}
	return false
}

// NativeSubscriber implements NativeSubscriberProvider interface.
// Returns the underlying rocketmq.PushConsumer for native SDK access.
//
// NativeSubscriber 实现 NativeSubscriberProvider 接口。
// 返回底层 rocketmq.PushConsumer 用于原生 SDK 访问。
func (s *rocketmqSubscriber) NativeSubscriber() any {
	return s.Underlying()
}
