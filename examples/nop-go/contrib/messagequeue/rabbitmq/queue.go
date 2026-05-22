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

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// Queue implements integrationcontract.MessageQueue using amqp091-go SDK.
// Manages connection, channels, and provides publisher/subscriber factories.
// Publisher operations create short-lived channels that are closed after use.
// Subscriber operations own their channels and close them on unsubscribe.
//
// Queue 使用 amqp091-go SDK 实现 integrationcontract.MessageQueue。
// 管理连接、channels，并提供 publisher/subscriber 工厂。
// Publisher 操作创建短期 channel，使用后关闭。
// Subscriber 操作持有自己的 channel，在取消订阅时关闭。
type Queue struct {
	cfg    *integrationcontract.MessageQueueConfig
	conn   *amqp.Connection
	mu     sync.Mutex
	closed bool
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

	// Create initial channel for exchange declaration
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

	// Close setup channel — publishers/subscribers will create their own.
	// 关闭初始化 channel — publisher/subscriber 会创建自己的 channel。
	ch.Close()

	return &Queue{
		cfg:  cfg,
		conn: conn,
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

// Close closes the RabbitMQ connection.
// Implements integrationcontract.MessageQueue.Close.
// Individual channels are closed by their owners (publishers close after use,
// subscribers close on unsubscribe).
//
// Close 关闭 RabbitMQ 连接。
// 实现 integrationcontract.MessageQueue.Close。
// 各 channel 由其持有者关闭（publisher 使用后关闭，subscriber 取消订阅时关闭）。
func (q *Queue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return nil
	}
	q.closed = true

	// Close connection — this implicitly closes all channels on it.
	// 关闭连接——这会隐式关闭其上的所有 channel。
	if q.conn != nil {
		return q.conn.Close()
	}
	return nil
}

// getChannel creates a new AMQP channel for short-lived operations.
// The caller MUST close the channel after use (typically via defer).
// This avoids the channel leak that occurs when channels are stored
// in a list but never closed.
//
// getChannel 创建新的 AMQP channel 用于短期操作。
// 调用方必须在使用后关闭 channel（通常通过 defer）。
// 这避免了将 channel 存入列表但从不关闭导致的泄漏。
func (q *Queue) getChannel() (*amqp.Channel, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil, errors.New("messagequeue.rabbitmq: queue closed")
	}

	ch, err := q.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("messagequeue.rabbitmq: create channel failed: %w", err)
	}

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
	return As(q.conn, target)
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
// The caller MUST close the channel after use.
//
// NativeChannel 返回一个新 channel 用于高级操作。
// 调用方必须在使用后关闭此 channel。
func (q *Queue) NativeChannel() (*amqp.Channel, error) {
	return q.getChannel()
}
