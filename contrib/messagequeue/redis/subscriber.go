package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

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

// deadLetterMaxRetries 是消息进入死信队列前的最大重试次数，
// 与退避一起防止毒丸消息在队列里零间隔热循环（单核 100% 空转）。
const deadLetterMaxRetries = 5

// retryBackoff 第 n 次失败后的回推延迟（1s, 2s, 4s, 8s, 16s 封顶）。
func retryBackoff(retries int) time.Duration {
	if retries < 0 {
		retries = 0
	}
	d := time.Second << uint(retries)
	if d > 16*time.Second {
		d = 16 * time.Second
	}
	return d
}

// consumeEnvelope 消息信封：JSON 编码头携带重试计数。
type consumeEnvelope struct {
	Retries int    `json:"gorp_retries"`
	Body    string `json:"body"`
}

// Consume consumes messages from a queue using Redis list (BLPOP).
func (s *redisSubscriber) Consume(ctx context.Context, queue string, handler integrationcontract.MessageHandler) error {
	deadLetterQueue := queue + ":dead"
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
		// 兼容无信封的裸消息（旧生产者或直接 RPUSH）。
		var env consumeEnvelope
		raw := result[1]
		if e := json.Unmarshal([]byte(raw), &env); e != nil || env.Body == "" {
			env = consumeEnvelope{Retries: 0, Body: raw}
		}

		message := &integrationcontract.Message{ID: "", Queue: queue, Body: []byte(env.Body), RetryCount: env.Retries, Timestamp: time.Now()}
		if err := handler(ctx, message); err != nil {
			if env.Retries+1 >= deadLetterMaxRetries {
				// 超过重试上限：进入死信队列，不再回推。
				if rErr := s.queue.client.RPush(ctx, deadLetterQueue, raw).Err(); rErr != nil {
					return fmt.Errorf("messagequeue.redis: push to dead letter queue: %w", rErr)
				}
				continue
			}
			env.Retries++
			payload, mErr := json.Marshal(env)
			if mErr != nil {
				payload = []byte(env.Body)
			}
			// 退避后回推队尾，避免零间隔热循环。
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryBackoff(env.Retries)):
			}
			if err := s.queue.client.RPush(ctx, queue, string(payload)).Err(); err != nil {
				return fmt.Errorf("messagequeue.redis: requeue message: %w", err)
			}
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
		if As(s.queue.pubsub, target) {
			return true
		}
	}
	// 否则尝试转换 client
	if s.queue.client != nil {
		return As(s.queue.client, target)
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
