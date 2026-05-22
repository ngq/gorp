package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// redisPublisher implements MessagePublisher using Redis.
type redisPublisher struct {
	queue *Queue
}

// Publish sends a message to a topic using Redis Pub/Sub.
func (p *redisPublisher) Publish(ctx context.Context, topic string, message []byte, options ...integrationcontract.PublishOption) error {
	startTime := time.Now()
	cfg := &integrationcontract.PublishConfig{}
	for _, opt := range options {
		opt(cfg)
	}
	if p == nil || p.queue == nil || p.queue.client == nil {
		return errors.New("messagequeue.redis: client not initialized")
	}
	err := p.queue.client.Publish(ctx, topic, message).Err()

	// 记录发布指标
	latency := time.Since(startTime).Seconds()
	status := "success"
	if err != nil {
		status = "error"
	}
	if p.queue.metrics != nil {
		p.queue.metrics.OnPublish(topic, status, len(message), latency)
	}

	return err
}

// PublishWithDelay sends a delayed message using Redis sorted set.
// The message is added to a sorted set with score = delivery time.
// Requires a separate scheduler to process delayed messages.
//
// PublishWithDelay 发送延迟消息，使用 Redis 有序集合。
// 消息添加到有序集合，score 为投递时间。
// 需要单独的调度器处理延迟消息。
func (p *redisPublisher) PublishWithDelay(ctx context.Context, topic string, message []byte, delay time.Duration) error {
	if p == nil || p.queue == nil || p.queue.client == nil {
		return errors.New("messagequeue.redis: client not initialized")
	}
	score := float64(time.Now().Add(delay).Unix())
	key := fmt.Sprintf("delay:%s", topic)
	return p.queue.client.ZAdd(ctx, key, redis.Z{Score: score, Member: message}).Err()
}

// PublishWithPriority sends a message with priority using Redis list.
// Higher priority messages are pushed to the front of the list.
//
// PublishWithPriority 发送带优先级的消息，使用 Redis 列表。
// 高优先级消息推送到列表头部。
func (p *redisPublisher) PublishWithPriority(ctx context.Context, topic string, message []byte, priority int) error {
	if p == nil || p.queue == nil || p.queue.client == nil {
		return errors.New("messagequeue.redis: client not initialized")
	}
	queueName := fmt.Sprintf("priority:%s:%d", topic, priority)
	return p.queue.client.LPush(ctx, queueName, message).Err()
}

// Send sends a message to a queue using Redis list (RPUSH).
func (p *redisPublisher) Send(ctx context.Context, queue string, message []byte, options ...integrationcontract.PublishOption) error {
	cfg := &integrationcontract.PublishConfig{}
	for _, opt := range options {
		opt(cfg)
	}
	if p == nil || p.queue == nil || p.queue.client == nil {
		return errors.New("messagequeue.redis: client not initialized")
	}
	return p.queue.client.RPush(ctx, queue, message).Err()
}

// Underlying returns the underlying Redis client for advanced usage.
//
// Underlying 返回底层 Redis 客户端供高级使用。
func (p *redisPublisher) Underlying() any {
	if p == nil || p.queue == nil {
		return nil
	}
	return p.queue.client
}

// As attempts to cast the underlying Redis client to the target type.
//
// As 尝试将底层 Redis 客户端转换为目标类型。
func (p *redisPublisher) As(target any) bool {
	if p == nil || p.queue == nil || p.queue.client == nil {
		return false
	}
	return As(p.queue.client, target)
}

// NativePublisher implements NativePublisherProvider interface.
// Returns the underlying *redis.Client for publishing operations.
//
// NativePublisher 实现 NativePublisherProvider 接口。
// 返回底层 *redis.Client 用于发布操作。
func (p *redisPublisher) NativePublisher() any {
	return p.Underlying()
}
