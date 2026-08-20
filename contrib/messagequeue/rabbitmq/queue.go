// Package rabbitmq provides RabbitMQ Queue implementation.
// This file implements the MessageQueue core structure with connection and channel management.
//
// 本包提供 RabbitMQ Queue 实现。
// 本文件实现 MessageQueue 核心结构，包含连接和 channel 管理。
package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// Queue implements integrationcontract.MessageQueue using amqp091-go SDK.
// 使用 ResilientConnection 统一托管连接、Single-Flight 重连与 Exchange 自动恢复。
type Queue struct {
	cfg     *integrationcontract.MessageQueueConfig
	resConn *integrationcontract.ResilientConnection[*amqp.Connection]
}

// NewQueue 创建新的 RabbitMQ Queue 实例。
func NewQueue(cfg *integrationcontract.MessageQueueConfig) (*Queue, error) {
	opts := integrationcontract.ResilientOptions[*amqp.Connection]{
		Dialer: func(ctx context.Context) (*amqp.Connection, error) {
			conn, err := amqp.Dial(cfg.RabbitMQURL)
			if err != nil {
				return nil, fmt.Errorf("messagequeue.rabbitmq: connect failed: %w", err)
			}
			return conn, nil
		},
		Checker: func(conn *amqp.Connection) bool {
			return conn != nil && !conn.IsClosed()
		},
		OnReconnect: func(ctx context.Context, conn *amqp.Connection) error {
			if cfg.RabbitMQExchange != "" {
				ch, chErr := conn.Channel()
				if chErr != nil {
					return chErr
				}
				defer ch.Close()
				return ch.ExchangeDeclare(
					cfg.RabbitMQExchange,
					cfg.RabbitMQExchangeType,
					true, false, false, false, nil,
				)
			}
			return nil
		},
		Closer: func(conn *amqp.Connection) error {
			if conn != nil && !conn.IsClosed() {
				return conn.Close()
			}
			return nil
		},
	}

	resConn, err := integrationcontract.NewResilientConnection(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	// 初始创建时声明 exchange
	if cfg.RabbitMQExchange != "" {
		conn, err := resConn.Get(context.Background())
		if err != nil {
			_ = resConn.Close()
			return nil, fmt.Errorf("messagequeue.rabbitmq: get initial conn failed: %w", err)
		}
		ch, chErr := conn.Channel()
		if chErr != nil {
			_ = resConn.Close()
			return nil, fmt.Errorf("messagequeue.rabbitmq: create initial channel failed: %w", chErr)
		}
		if err := ch.ExchangeDeclare(
			cfg.RabbitMQExchange,
			cfg.RabbitMQExchangeType,
			true, false, false, false, nil,
		); err != nil {
			_ = ch.Close()
			_ = resConn.Close()
			return nil, fmt.Errorf("messagequeue.rabbitmq: declare exchange failed: %w", err)
		}
		_ = ch.Close()
	}

	return &Queue{
		cfg:     cfg,
		resConn: resConn,
	}, nil
}

// Publisher 返回基于 RabbitMQ 的 MessagePublisher。
func (q *Queue) Publisher() integrationcontract.MessagePublisher {
	return &rabbitPublisher{queue: q}
}

// Subscriber 返回基于 RabbitMQ 的 MessageSubscriber。
func (q *Queue) Subscriber() integrationcontract.MessageSubscriber {
	return &rabbitSubscriber{queue: q}
}

// Close 关闭 RabbitMQ 连接。
func (q *Queue) Close() error {
	return q.resConn.Close()
}

// getConn 返回当前可用连接，连接断开时自动通过 ResilientConnection 重连。
func (q *Queue) getConn() (*amqp.Connection, error) {
	return q.resConn.Get(context.Background())
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
	conn, err := q.getConn()
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
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
	if q == nil || q.resConn == nil {
		return nil
	}
	conn, _ := q.resConn.Get(context.Background())
	return conn
}

// As attempts to cast the underlying *amqp.Connection to the target type.
//
// As 尝试将底层 *amqp.Connection 转换为目标类型。
func (q *Queue) As(target any) bool {
	conn := q.Underlying()
	if conn == nil {
		return false
	}
	return As(conn, target)
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
