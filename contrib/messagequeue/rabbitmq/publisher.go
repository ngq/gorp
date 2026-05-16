// Package rabbitmq provides RabbitMQ Publisher implementation.
// This file implements the MessagePublisher contract using amqp091-go SDK.
//
// 本包提供 RabbitMQ Publisher 实现。
// 本文件使用 amqp091-go SDK 实现 MessagePublisher 契约。
package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// rabbitPublisher implements MessagePublisher using amqp091-go.
//
// rabbitPublisher 使用 amqp091-go 实现 MessagePublisher。
type rabbitPublisher struct {
	queue *Queue
}

// Publish sends a message to a topic (routing key).
// Implements integrationcontract.MessagePublisher.Publish.
//
// Publish 将消息发送到 topic（routing key）。
// 实现 integrationcontract.MessagePublisher.Publish。
func (p *rabbitPublisher) Publish(ctx context.Context, topic string, message []byte, options ...integrationcontract.PublishOption) error {
	ch, err := p.queue.getChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	cfg := &integrationcontract.PublishConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	// Publish to exchange (or direct to queue if no exchange)
	exchange := p.queue.cfg.RabbitMQExchange

	return ch.PublishWithContext(
		ctx,
		exchange,
		topic, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/octet-stream",
			Body:         message,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent, // durable message
		},
	)
}

// PublishWithDelay sends a delayed message using x-delayed-message header.
// Requires RabbitMQ delayed message plugin or dead-letter queue setup.
//
// PublishWithDelay 发送延迟消息，使用 x-delayed-message header。
// 需要 RabbitMQ delayed message 插件或死信队列设置。
func (p *rabbitPublisher) PublishWithDelay(ctx context.Context, topic string, message []byte, delay time.Duration) error {
	ch, err := p.queue.getChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	exchange := p.queue.cfg.RabbitMQExchange

	return ch.PublishWithContext(
		ctx,
		exchange,
		topic,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/octet-stream",
			Body:         message,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
			Headers: amqp.Table{
				"x-delay": delay.Milliseconds(),
			},
		},
	)
}

// PublishWithPriority sends a message with priority.
// Requires queue to be declared with x-max-priority argument.
//
// PublishWithPriority 发送带优先级的消息。
// 需要队列声明时设置 x-max-priority 参数。
func (p *rabbitPublisher) PublishWithPriority(ctx context.Context, topic string, message []byte, priority int) error {
	ch, err := p.queue.getChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	exchange := p.queue.cfg.RabbitMQExchange

	return ch.PublishWithContext(
		ctx,
		exchange,
		topic,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/octet-stream",
			Body:         message,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
			Priority:     uint8(priority),
		},
	)
}

// Send sends a message directly to a queue.
// Implements integrationcontract.MessagePublisher.Send.
//
// Send 将消息直接发送到队列。
// 实现 integrationcontract.MessagePublisher.Send。
func (p *rabbitPublisher) Send(ctx context.Context, queue string, message []byte, options ...integrationcontract.PublishOption) error {
	ch, err := p.queue.getChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	cfg := &integrationcontract.PublishConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	// Declare queue first (durable)
	queueName := p.queue.cfg.RabbitMQQueuePrefix + queue
	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("messagequeue.rabbitmq: declare queue failed: %w", err)
	}

	return ch.PublishWithContext(
		ctx,
		"", // direct to queue (empty exchange)
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/octet-stream",
			Body:         message,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent,
		},
	)
}

// Underlying returns the underlying *amqp.Connection.
//
// Underlying 返回底层 *amqp.Connection。
func (p *rabbitPublisher) Underlying() any {
	if p == nil || p.queue == nil {
		return nil
	}
	return p.queue.conn
}

// As attempts to cast the underlying connection to the target type.
//
// As 尝试将底层连接转换为目标类型。
func (p *rabbitPublisher) As(target any) bool {
	if p == nil || p.queue == nil || p.queue.conn == nil {
		return false
	}
	return internalnative.As(p.queue.conn, target)
}

// NativePublisher implements NativePublisherProvider interface.
// Returns the underlying *amqp.Connection for advanced publishing.
// Callers should create and close their own channels from this connection.
//
// NativePublisher 实现 NativePublisherProvider 接口。
// 返回底层 *amqp.Connection 用于高级发布操作。
// 调用方应从该连接创建和关闭自己的 channel。
func (p *rabbitPublisher) NativePublisher() any {
	if p == nil || p.queue == nil || p.queue.conn == nil {
		return nil
	}
	return p.queue.conn
}
