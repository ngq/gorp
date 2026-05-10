// Package rabbitmq provides RabbitMQ Queue implementation.
// This file implements the MessageQueue core structure with connection and channel management.
//
// 本包提供 RabbitMQ Queue 实现。
// 本文件实现 MessageQueue 核心结构，包含连接和 channel 管理。
package rabbitmq

import (
	"errors"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// Queue implements integrationcontract.MessageQueue using amqp091-go SDK.
// Manages connection, channels, and provides publisher/subscriber factories.
//
// Queue 使用 amqp091-go SDK 实现 integrationcontract.MessageQueue。
// 管理连接、channels，并提供 publisher/subscriber 工厂。
type Queue struct {
	cfg      *integrationcontract.MessageQueueConfig
	conn     *amqp.Connection
	channels []*amqp.Channel
	mu       sync.Mutex
	closed   bool
}

// NewQueue creates a new RabbitMQ Queue instance.
// Establishes connection, declares exchange if configured, and sets up QoS.
//
// NewQueue 创建新的 RabbitMQ Queue 实例。
// 建立连接，如配置则声明 exchange，并设置 QoS。
func NewQueue(cfg *integrationcontract.MessageQueueConfig) (*Queue, error) {
	// Parse URL and connect
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("messagequeue.rabbitmq: connect failed: %w", err)
	}

	// Create initial channel for setup
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("messagequeue.rabbitmq: create channel failed: %w", err)
	}

	// Declare exchange if configured
	if cfg.RabbitMQExchange != "" {
		err = ch.ExchangeDeclare(
			cfg.RabbitMQExchange,
			cfg.RabbitMQExchangeType,
			true,  // durable
			false, // auto-delete
			false, // internal
			false, // no-wait
			nil,   // args
		)
		if err != nil {
			ch.Close()
			conn.Close()
			return nil, fmt.Errorf("messagequeue.rabbitmq: declare exchange failed: %w", err)
		}
	}

	// Set prefetch for this channel
	err = ch.Qos(cfg.RabbitMQPrefetch, 0, false)
	if err != nil {
		// Non-critical error, continue
	}

	return &Queue{
		cfg:      cfg,
		conn:     conn,
		channels: []*amqp.Channel{ch},
	}, nil
}

// Publisher returns a RabbitMQ-based MessagePublisher.
//
// Publisher 返回基于 RabbitMQ 的 MessagePublisher。
func (q *Queue) Publisher() integrationcontract.MessagePublisher {
	return &rabbitPublisher{queue: q}
}

// Subscriber returns a RabbitMQ-based MessageSubscriber.
//
// Subscriber 返回基于 RabbitMQ 的 MessageSubscriber。
func (q *Queue) Subscriber() integrationcontract.MessageSubscriber {
	return &rabbitSubscriber{queue: q}
}

// Close closes all RabbitMQ resources.
// Implements integrationcontract.MessageQueue.Close.
//
// Close 关闭所有 RabbitMQ 资源。
// 实现 integrationcontract.MessageQueue.Close。
func (q *Queue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return nil
	}
	q.closed = true

	// Close all channels
	for _, ch := range q.channels {
		ch.Close()
	}
	q.channels = nil

	// Close connection
	if q.conn != nil {
		return q.conn.Close()
	}
	return nil
}

// getChannel returns an available channel or creates a new one.
//
// getChannel 返回可用 channel 或创建新的。
func (q *Queue) getChannel() (*amqp.Channel, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil, errors.New("messagequeue.rabbitmq: queue closed")
	}

	// Return existing channel if available
	for _, ch := range q.channels {
		if !ch.IsClosed() {
			return ch, nil
		}
	}

	// Create new channel
	ch, err := q.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("messagequeue.rabbitmq: create channel failed: %w", err)
	}

	// Set QoS
	err = ch.Qos(q.cfg.RabbitMQPrefetch, 0, false)
	if err != nil {
		// Non-critical, continue
	}

	q.channels = append(q.channels, ch)
	return ch, nil
}

// Underlying returns the underlying *amqp.Connection for advanced usage.
// This allows users to access native RabbitMQ SDK capabilities such as
// publisher confirms, transactions, custom channels, etc.
//
// Underlying 返回底层 *amqp.Connection 供高级使用。
// 这允许用户访问原生 RabbitMQ SDK 能力，如 publisher confirms、事务、自定义 channel 等。
func (q *Queue) Underlying() any {
	if q == nil {
		return nil
	}
	return q.conn
}

// As attempts to cast the underlying *amqp.Connection to the target type.
//
// As 尝试将底层 *amqp.Connection 转换为目标类型。
func (q *Queue) As(target any) bool {
	if q == nil || q.conn == nil {
		return false
	}
	return internalnative.As(q.conn, target)
}

// NativeMQClient implements NativeMQClientProvider interface.
// Returns the underlying *amqp.Connection.
//
// NativeMQClient 实现 NativeMQClientProvider 接口。
// 返回底层 *amqp.Connection。
func (q *Queue) NativeMQClient() any {
	return q.Underlying()
}

// NativeChannel returns a fresh channel for advanced operations.
// Users should close this channel after use.
//
// NativeChannel 返回一个新 channel 用于高级操作。
// 用户应在使用后关闭此 channel。
func (q *Queue) NativeChannel() (*amqp.Channel, error) {
	return q.getChannel()
}