package redis

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/ngq/gorp/contrib/internal/native"
	"github.com/ngq/gorp/contrib/messagequeue"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// Queue implements integrationcontract.MessageQueue using Redis SDK.
type Queue struct {
	cfg     *integrationcontract.MessageQueueConfig
	client  *redis.Client
	pubsub  *redis.PubSub
	mu      sync.Mutex
	subs    map[string]context.CancelFunc
	closed  bool
	metrics *messagequeue.MetricsRecorder // Prometheus 指标记录器
}

// NewQueue creates a new Redis Queue instance.
func NewQueue(cfg *integrationcontract.MessageQueueConfig) (*Queue, error) {
	client := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("messagequeue.redis: connect failed: %w", err)
	}
	return &Queue{
		cfg:     cfg,
		client:  client,
		subs:    make(map[string]context.CancelFunc),
		metrics: messagequeue.NewMetricsRecorder(),
	}, nil
}

// Publisher returns a Redis-based MessagePublisher.
func (q *Queue) Publisher() integrationcontract.MessagePublisher {
	return &redisPublisher{queue: q}
}

// Subscriber returns a Redis-based MessageSubscriber.
func (q *Queue) Subscriber() integrationcontract.MessageSubscriber {
	return &redisSubscriber{queue: q}
}

// Close closes all Redis resources.
func (q *Queue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return nil
	}
	q.closed = true
	for _, cancel := range q.subs {
		cancel()
	}
	if q.pubsub != nil {
		q.pubsub.Close()
	}
	if q.client == nil {
		return nil
	}
	return q.client.Close()
}

// Underlying returns the underlying Redis client for advanced usage.
// This allows users to access native Redis SDK capabilities such as
// pipeline, Lua scripts, streams, transactions, etc.
//
// Underlying 返回底层 Redis 客户端供高级使用。
// 这允许用户访问原生 Redis SDK 能力，如 pipeline、Lua 脚本、streams、事务等。
func (q *Queue) Underlying() any {
	if q == nil {
		return nil
	}
	return q.client
}

// As attempts to cast the underlying Redis client to the target type.
// This is useful for type-safe access to the native client.
//
// As 尝试将底层 Redis 客户端转换为目标类型。
// 用于类型安全地访问原生客户端。
func (q *Queue) As(target any) bool {
	if q == nil || q.client == nil {
		return false
	}
	return native.As(q.client, target)
}

// NativeMQClient implements NativeMQClientProvider interface.
// Returns the underlying *redis.Client.
//
// NativeMQClient 实现 NativeMQClientProvider 接口。
// 返回底层 *redis.Client。
func (q *Queue) NativeMQClient() any {
	return q.Underlying()
}
