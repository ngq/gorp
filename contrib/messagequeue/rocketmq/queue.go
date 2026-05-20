// Package rocketmq provides RocketMQ Queue implementation.
// This file implements the MessageQueue core structure with producer and consumer management.
//
// 本包提供 RocketMQ Queue 实现。
// 本文件实现 MessageQueue 核心结构，包含 producer 和 consumer 管理。
package rocketmq

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/producer"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// ErrQueueClosed is returned when operations are attempted on a closed queue.
var ErrQueueClosed = errors.New("messagequeue.rocketmq: queue closed")

// Queue implements integrationcontract.MessageQueue using rocketmq-client-go SDK.
// Manages producer and consumer instances, provides publisher/subscriber factories.
// Each subscription group gets its own consumer instance to avoid resource leaks.
//
// Queue 使用 rocketmq-client-go SDK 实现 integrationcontract.MessageQueue。
// 管理 producer 和 consumer 实例，提供 publisher/subscriber 工厂。
// 每个订阅组拥有独立的 consumer 实例以避免资源泄漏。
type Queue struct {
	cfg       *integrationcontract.MessageQueueConfig
	producer  rocketmq.Producer
	consumers map[string]rocketmq.PushConsumer
	mu        sync.Mutex
	closed    bool
}

// NewQueue creates a new RocketMQ Queue instance.
// Creates and starts a Producer for publishing messages.
//
// NewQueue 创建新的 RocketMQ Queue 实例。
// 创建并启动 Producer 用于发布消息。
func NewQueue(cfg *integrationcontract.MessageQueueConfig) (*Queue, error) {
	// Parse nameserver addresses (semicolon-separated)
	namesrvAddr := strings.Split(cfg.RocketMQNamesrvAddr, ";")

	// Create producer with configuration
	p, err := rocketmq.NewProducer(
		producer.WithNameServer(namesrvAddr),
		producer.WithGroupName(cfg.RocketMQGroupName),
		producer.WithInstanceName(cfg.RocketMQInstanceName),
		producer.WithRetry(cfg.RocketMQRetryTimes),
	)
	if err != nil {
		return nil, fmt.Errorf("messagequeue.rocketmq: create producer failed: %w", err)
	}

	// Start producer
	err = p.Start()
	if err != nil {
		return nil, fmt.Errorf("messagequeue.rocketmq: start producer failed: %w", err)
	}

	return &Queue{
		cfg:       cfg,
		producer:  p,
		consumers: make(map[string]rocketmq.PushConsumer),
	}, nil
}

// Publisher returns a RocketMQ-based MessagePublisher.
//
// Publisher 返回基于 RocketMQ 的 MessagePublisher。
func (q *Queue) Publisher() integrationcontract.MessagePublisher {
	return &rocketmqPublisher{queue: q}
}

// Subscriber returns a RocketMQ-based MessageSubscriber.
//
// Subscriber 返回基于 RocketMQ 的 MessageSubscriber。
func (q *Queue) Subscriber() integrationcontract.MessageSubscriber {
	return &rocketmqSubscriber{queue: q}
}

// Close closes all RocketMQ resources.
// Implements integrationcontract.MessageQueue.Close.
// Shuts down both producer and consumer if they exist.
//
// Close 关闭所有 RocketMQ 资源。
// 实现 integrationcontract.MessageQueue.Close。
// 如果存在则关闭 producer 和 consumer。
func (q *Queue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return nil
	}
	q.closed = true

	// Close producer
	if q.producer != nil {
		q.producer.Shutdown()
	}

	// Close all consumers
	for group, c := range q.consumers {
		c.Shutdown()
		delete(q.consumers, group)
	}

	return nil
}

// Underlying returns the underlying rocketmq.Producer for advanced usage.
// This allows users to access native RocketMQ SDK capabilities such as
// order messages, batch sending, transaction messages, etc.
//
// Underlying 返回底层 rocketmq.Producer 供高级使用。
// 这允许用户访问原生 RocketMQ SDK 能力，如顺序消息、批量发送、事务消息等。
func (q *Queue) Underlying() any {
	if q == nil {
		return nil
	}
	return q.producer
}

// As attempts to cast the underlying producer to the target type.
// Uses the internal native.As helper for type casting.
//
// As 尝试将底层 producer 转换为目标类型。
// 使用内部 native.As 辅助函数进行类型转换。
func (q *Queue) As(target any) bool {
	if q == nil || q.producer == nil {
		return false
	}
	return As(q.producer, target)
}

// NativeMQClient implements NativeMQClientProvider interface.
// Returns the underlying rocketmq.Producer.
// This allows "MQ-first" users to access native SDK capabilities
// while staying within the framework's governance boundary.
//
// NativeMQClient 实现 NativeMQClientProvider 接口。
// 返回底层 rocketmq.Producer。
// 这允许"MQ-first"用户访问原生 SDK 能力，同时保持在框架治理边界内。
func (q *Queue) NativeMQClient() any {
	return q.Underlying()
}

// createConsumer creates a new PushConsumer for subscribing to messages.
// This is called by Subscriber when SubscribeWithGroup is invoked.
// Each subscription creates its own consumer instance.
//
// createConsumer 创建新的 PushConsumer 用于订阅消息。
// 当 Subscriber 调用 SubscribeWithGroup 时触发。
// 每个订阅创建自己的 consumer 实例。
func (q *Queue) createConsumer(group string) (rocketmq.PushConsumer, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil, ErrQueueClosed
	}

	// Reuse existing consumer for the same group
	if c, ok := q.consumers[group]; ok {
		return c, nil
	}

	namesrvAddr := strings.Split(q.cfg.RocketMQNamesrvAddr, ";")

	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer(namesrvAddr),
		consumer.WithGroupName(group),
		consumer.WithInstance(q.cfg.RocketMQInstanceName),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromFirstOffset),
	)
	if err != nil {
		return nil, fmt.Errorf("messagequeue.rocketmq: create consumer failed: %w", err)
	}

	q.consumers[group] = c
	return c, nil
}
