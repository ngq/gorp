// Package rabbitmq provides RabbitMQ Subscriber implementation.
// This file implements the MessageSubscriber contract using amqp091-go SDK.
//
// 本包提供 RabbitMQ Subscriber 实现。
// 本文件使用 amqp091-go SDK 实现 MessageSubscriber 契约。
package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// rabbitSubscriber implements MessageSubscriber using amqp091-go.
//
// rabbitSubscriber 使用 amqp091-go 实现 MessageSubscriber。
type rabbitSubscriber struct {
	queue     *Queue
	cancelMap map[string]context.CancelFunc
	mu        sync.Mutex
}

// channelOnce wraps a channel with its close-once guard to prevent
// the double-close race between the consume goroutine's defer and
// the unsubscribe function.
//
// channelOnce 包装 channel 及其一次性关闭守卫，
// 防止消费 goroutine 的 defer 和取消订阅函数之间的双重关闭竞态。
type channelOnce struct {
	ch    *amqp.Channel
	close sync.Once
}

func (co *channelOnce) Close() error {
	var err error
	co.close.Do(func() {
		err = co.ch.Close()
	})
	return err
}

// Subscribe subscribes to a topic using queue binding.
// Implements integrationcontract.MessageSubscriber.Subscribe.
//
// Subscribe 使用队列绑定订阅 topic。
// 实现 integrationcontract.MessageSubscriber.Subscribe。
func (s *rabbitSubscriber) Subscribe(ctx context.Context, topic string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	return s.SubscribeWithGroup(ctx, topic, "", handler)
}

// SubscribeWithGroup subscribes to a topic with a specific consumer group (queue name suffix).
// In RabbitMQ, consumer group is implemented via unique queue name per consumer.
//
// SubscribeWithGroup 使用特定消费者组（队列名后缀）订阅主题。
// 在 RabbitMQ 中，消费者组通过每个消费者唯一队列名实现。
func (s *rabbitSubscriber) SubscribeWithGroup(ctx context.Context, topic string, group string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	if _, err := s.queue.getConn(); err != nil {
		return nil, err
	}

	// Create queue name with group suffix
	queueName := s.queue.cfg.RabbitMQQueuePrefix + topic
	if group != "" {
		queueName = queueName + "-" + group
	} else {
		queueName = queueName + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
	}

	subCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	if s.cancelMap == nil {
		s.cancelMap = make(map[string]context.CancelFunc)
	}
	s.cancelMap[queueName] = cancel
	s.mu.Unlock()

	// 订阅存活循环：channel/连接断开时退避重建，消费不再静默死亡。
	go func() {
		backoff := time.Second
		const maxBackoff = 30 * time.Second
		for {
			select {
			case <-subCtx.Done():
				return
			default:
			}
			err := s.runSubscription(subCtx, topic, queueName, group == "", handler)
			if errors.Is(err, context.Canceled) {
				return
			}
			slog.Warn("messagequeue.rabbitmq: subscription interrupted, re-establishing",
				"topic", topic, "queue", queueName, "error", err)
			select {
			case <-subCtx.Done():
				return
			case <-time.After(backoff):
			}
			if backoff < maxBackoff {
				backoff *= 2
			}
		}
	}()

	return func() error {
		cancel()
		return nil
	}, nil
}

// runSubscription 建立一次完整的订阅（channel + queue + bind + consume），
// 直到 context 取消或投递 channel 关闭（连接断开）。
func (s *rabbitSubscriber) runSubscription(ctx context.Context, topic string, queueName string, anonymous bool, handler integrationcontract.MessageHandler) error {
	conn, err := s.queue.getConn()
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("messagequeue.rabbitmq: create channel failed: %w", err)
	}
	wrappedCh := &channelOnce{ch: ch}
	defer wrappedCh.Close()

	if err := ch.Qos(s.queue.cfg.RabbitMQPrefetch, 0, false); err != nil {
		return err
	}

	// 匿名（无 group）订阅用 exclusive + auto-delete 队列：连接关闭即由
	// broker 自动删除。此前用 durable 队列且从不 QueueDelete，每次进程
	// 重启都在 broker 上遗留一个永久孤儿队列（消息黑洞）。
	durable, autoDelete, exclusive := true, false, false
	if anonymous {
		durable, autoDelete, exclusive = false, true, true
	}
	q, err := ch.QueueDeclare(
		queueName,
		durable,
		autoDelete,
		exclusive,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("messagequeue.rabbitmq: declare queue failed: %w", err)
	}

	if s.queue.cfg.RabbitMQExchange != "" {
		if err := ch.QueueBind(q.Name, topic, s.queue.cfg.RabbitMQExchange, false, nil); err != nil {
			return fmt.Errorf("messagequeue.rabbitmq: bind queue failed: %w", err)
		}
	}

	msgs, err := ch.Consume(
		q.Name,
		"",        // consumer tag (auto-generated)
		false,     // auto-ack (manual for reliability)
		exclusive, // exclusive consumer for anonymous queues
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("messagequeue.rabbitmq: consume failed: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				// channel 关闭（连接断开/rebalance）：返回由外层重建订阅。
				return errors.New("messagequeue.rabbitmq: delivery channel closed")
			}
			message := &integrationcontract.Message{
				ID:        msg.MessageId,
				Topic:     msg.RoutingKey,
				Body:      msg.Body,
				Headers:   extractAMQPHeaders(msg.Headers),
				Timestamp: msg.Timestamp,
			}
			if err := handler(ctx, message); err != nil {
				// Nack and requeue
				msg.Nack(false, true)
			} else {
				// Ack
				msg.Ack(false)
			}
		}
	}
}

// Consume consumes messages from a specific queue.
// Implements integrationcontract.MessageSubscriber.Consume.
//
// Consume 从特定队列消费消息。
// 实现 integrationcontract.MessageSubscriber.Consume。
func (s *rabbitSubscriber) Consume(ctx context.Context, queue string, handler integrationcontract.MessageHandler) error {
	conn, err := s.queue.getConn()
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("messagequeue.rabbitmq: create channel failed: %w", err)
	}
	defer ch.Close()

	queueName := s.queue.cfg.RabbitMQQueuePrefix + queue

	msgs, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("messagequeue.rabbitmq: consume failed: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				// channel 关闭不能伪装成功：调用方需要可区分的错误来决定重试。
				return errors.New("messagequeue.rabbitmq: delivery channel closed during consume")
			}
			message := &integrationcontract.Message{
				ID:        msg.MessageId,
				Topic:     msg.RoutingKey,
				Body:      msg.Body,
				Headers:   extractAMQPHeaders(msg.Headers),
				Timestamp: msg.Timestamp,
			}
			if err := handler(ctx, message); err != nil {
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
			}
		}
	}
}

// Unsubscribe cancels all active subscriptions.
// Implements integrationcontract.MessageSubscriber.Unsubscribe.
//
// UnsubscribeAll 取消所有活跃订阅。
// 实现 integrationcontract.MessageSubscriber.UnsubscribeAll。
func (s *rabbitSubscriber) UnsubscribeAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, cancel := range s.cancelMap {
		cancel()
	}
	s.cancelMap = make(map[string]context.CancelFunc)
	return nil
}

// Underlying returns the underlying *amqp.Connection.
//
// Underlying 返回底层 *amqp.Connection。
func (s *rabbitSubscriber) Underlying() any {
	if s == nil || s.queue == nil {
		return nil
	}
	return s.queue.Underlying()
}

// As attempts to cast the underlying connection to the target type.
//
// As 尝试将底层连接转换为目标类型。
func (s *rabbitSubscriber) As(target any) bool {
	conn := s.Underlying()
	if conn == nil {
		return false
	}
	return As(conn, target)
}

// NativeSubscriber implements NativeSubscriberProvider interface.
// Returns the underlying *amqp.Connection for advanced subscription operations.
// Callers should create and close their own channels from this connection.
//
// NativeSubscriber 实现 NativeSubscriberProvider 接口。
// 返回底层 *amqp.Connection 用于高级订阅操作。
// 调用方应从该连接创建和关闭自己的 channel。
func (s *rabbitSubscriber) NativeSubscriber() any {
	return s.Underlying()
}

// extractAMQPHeaders converts amqp.Table to map[string]string.
//
// extractAMQPHeaders 将 amqp.Table 转换为 map[string]string。
func extractAMQPHeaders(headers amqp.Table) map[string]string {
	result := make(map[string]string)
	for k, v := range headers {
		if str, ok := v.(string); ok {
			result[k] = str
		} else {
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}
