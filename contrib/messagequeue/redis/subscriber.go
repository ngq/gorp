package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	internalnative "github.com/ngq/gorp/contrib/internal/native"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// redisSubscriber implements MessageSubscriber using Redis.
type redisSubscriber struct {
	queue *Queue
}

// Subscribe subscribes to a topic using Redis Pub/Sub.
func (s *redisSubscriber) Subscribe(ctx context.Context, topic string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	s.queue.mu.Lock()
	defer s.queue.mu.Unlock()
	if s.queue.closed {
		return nil, errors.New("messagequeue.redis: queue closed")
	}
	pubsub := s.queue.client.Subscribe(ctx, topic)
	subCtx, cancel := context.WithCancel(ctx)
	subKey := fmt.Sprintf("sub:%s", topic)
	s.queue.subs[subKey] = cancel
	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()
		for {
			select {
			case <-subCtx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				message := &integrationcontract.Message{ID: "", Topic: topic, Body: []byte(msg.Payload), Timestamp: time.Now()}
				_ = handler(subCtx, message)
			}
		}
	}()
	return func() error {
		cancel()
		s.queue.mu.Lock()
		delete(s.queue.subs, subKey)
		s.queue.mu.Unlock()
		return pubsub.Close()
	}, nil
}

// SubscribeWithGroup subscribes to a topic with group (delegates to Subscribe).
// Redis Pub/Sub does not support consumer groups natively.
func (s *redisSubscriber) SubscribeWithGroup(ctx context.Context, topic string, group string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	_ = group
	return s.Subscribe(ctx, topic, handler)
}

// Consume consumes messages from a queue using Redis list (BLPOP).
func (s *redisSubscriber) Consume(ctx context.Context, queue string, handler integrationcontract.MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		result, err := s.queue.client.BLPop(ctx, time.Second, queue).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			return err
		}
		message := &integrationcontract.Message{ID: "", Queue: queue, Body: []byte(result[1]), Timestamp: time.Now()}
		if err := handler(ctx, message); err != nil {
			s.queue.client.RPush(ctx, queue, result[1])
		}
	}
}

// UnsubscribeAll cancels all active subscriptions.
func (s *redisSubscriber) UnsubscribeAll() error {
	s.queue.mu.Lock()
	defer s.queue.mu.Unlock()
	for _, cancel := range s.queue.subs {
		cancel()
	}
	s.queue.subs = make(map[string]context.CancelFunc)
	return nil
}

// Underlying returns the underlying Redis PubSub for advanced usage.
//
// Underlying 返回底层 Redis PubSub 供高级使用。
func (s *redisSubscriber) Underlying() any {
	if s == nil || s.queue == nil {
		return nil
	}
	return s.queue.pubsub
}

// As attempts to cast the underlying Redis PubSub to the target type.
//
// As 尝试将底层 Redis PubSub 转换为目标类型。
func (s *redisSubscriber) As(target any) bool {
	if s == nil || s.queue == nil {
		return false
	}
	// 如果 pubsub 存在，尝试转换它
	if s.queue.pubsub != nil {
		if internalnative.As(s.queue.pubsub, target) {
			return true
		}
	}
	// 否则尝试转换 client
	if s.queue.client != nil {
		return internalnative.As(s.queue.client, target)
	}
	return false
}

// NativeSubscriber implements NativeSubscriberProvider interface.
// Returns the underlying *redis.PubSub for subscription operations.
//
// NativeSubscriber 实现 NativeSubscriberProvider 接口。
// 返回底层 *redis.PubSub 用于订阅操作。
func (s *redisSubscriber) NativeSubscriber() any {
	return s.queue.pubsub
}
