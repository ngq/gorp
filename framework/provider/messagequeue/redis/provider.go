package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/redis/go-redis/v9"
)

// Provider 提供 Redis PubSub 消息队列实现。
//
// 中文说明：
// - 使用 Redis Pub/Sub 实现发布/订阅模式；
// - 使用 Redis List 实现点对点队列；
// - 轻量级，适合简单场景；
// - 需要项目引入 github.com/redis/go-redis/v9 依赖。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "messagequeue.redis" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.MessageQueueKey, contract.MessagePublisherKey, contract.MessageSubscriberKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.MessageQueueKey, func(c contract.Container) (any, error) {
		cfg, err := getMQConfig(c)
		if err != nil {
			return nil, err
		}
		return NewQueue(cfg)
	}, true)

	c.Bind(contract.MessagePublisherKey, func(c contract.Container) (any, error) {
		cfg, err := getMQConfig(c)
		if err != nil {
			return nil, err
		}
		queue, err := NewQueue(cfg)
		if err != nil {
			return nil, err
		}
		return queue.Publisher(), nil
	}, true)

	c.Bind(contract.MessageSubscriberKey, func(c contract.Container) (any, error) {
		cfg, err := getMQConfig(c)
		if err != nil {
			return nil, err
		}
		queue, err := NewQueue(cfg)
		if err != nil {
			return nil, err
		}
		return queue.Subscriber(), nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// getMQConfig 从容器获取消息队列配置。
func getMQConfig(c contract.Container) (*contract.MessageQueueConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("messagequeue: invalid config service")
	}

	mqCfg := &contract.MessageQueueConfig{
		Type:           "redis",
		MaxRetry:       3,
		RetryDelay:     time.Second,
		Timeout:        5 * time.Second,
		ConsumerBuffer: 100,
	}

	// Redis 配置
	if addr := cfg.GetString("message_queue.redis.addr"); addr != "" {
		mqCfg.RedisAddr = addr
	} else {
		mqCfg.RedisAddr = "localhost:6379"
	}
	if password := cfg.GetString("message_queue.redis.password"); password != "" {
		mqCfg.RedisPassword = password
	}
	if db := cfg.GetInt("message_queue.redis.db"); db > 0 {
		mqCfg.RedisDB = db
	}

	// 通用配置
	if maxRetry := cfg.GetInt("message_queue.max_retry"); maxRetry > 0 {
		mqCfg.MaxRetry = maxRetry
	}

	return mqCfg, nil
}

// Queue 是 Redis 消息队列实现。
type Queue struct {
	cfg    *contract.MessageQueueConfig
	client *redis.Client
	pubsub *redis.PubSub

	mu       sync.Mutex
	subs     map[string]context.CancelFunc
	closed   bool
}

// NewQueue 创建 Redis 消息队列。
func NewQueue(cfg *contract.MessageQueueConfig) (*Queue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("messagequeue.redis: connect failed: %w", err)
	}

	return &Queue{
		cfg:  cfg,
		client: client,
		subs: make(map[string]context.CancelFunc),
	}, nil
}

// Publisher 返回消息发布者。
func (q *Queue) Publisher() contract.MessagePublisher {
	return &redisPublisher{queue: q}
}

// Subscriber 返回消息订阅者。
func (q *Queue) Subscriber() contract.MessageSubscriber {
	return &redisSubscriber{queue: q}
}

// Close 关闭消息队列连接。
func (q *Queue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil
	}
	q.closed = true

	// 取消所有订阅
	for _, cancel := range q.subs {
		cancel()
	}

	if q.pubsub != nil {
		q.pubsub.Close()
	}

	return q.client.Close()
}

// redisPublisher 是 Redis 消息发布者实现。
type redisPublisher struct {
	queue *Queue
}

// Publish 发布消息到主题（Pub/Sub 模式）。
func (p *redisPublisher) Publish(ctx context.Context, topic string, message []byte, options ...contract.PublishOption) error {
	cfg := &contract.PublishConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	// Redis Pub/Sub 不支持延迟和优先级，直接发布
	return p.queue.client.Publish(ctx, topic, message).Err()
}

// PublishWithDelay 发布延迟消息（使用 List + 定时任务模拟）。
func (p *redisPublisher) PublishWithDelay(ctx context.Context, topic string, message []byte, delay time.Duration) error {
	// 使用 Sorted Set 存储延迟消息
	score := float64(time.Now().Add(delay).Unix())
	key := fmt.Sprintf("delay:%s", topic)
	return p.queue.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: message,
	}).Err()
}

// PublishWithPriority 发布优先级消息（使用多个 List 模拟）。
func (p *redisPublisher) PublishWithPriority(ctx context.Context, topic string, message []byte, priority int) error {
	// 使用不同的队列存储不同优先级
	queueName := fmt.Sprintf("priority:%s:%d", topic, priority)
	return p.queue.client.LPush(ctx, queueName, message).Err()
}

// Send 发送消息到队列（点对点模式）。
func (p *redisPublisher) Send(ctx context.Context, queue string, message []byte, options ...contract.PublishOption) error {
	cfg := &contract.PublishConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	// 使用 List 实现队列
	return p.queue.client.RPush(ctx, queue, message).Err()
}

// redisSubscriber 是 Redis 消息订阅者实现。
type redisSubscriber struct {
	queue *Queue
}

// Subscribe 订阅主题。
func (s *redisSubscriber) Subscribe(ctx context.Context, topic string, handler contract.MessageHandler) (contract.UnsubscribeFunc, error) {
	s.queue.mu.Lock()
	defer s.queue.mu.Unlock()

	if s.queue.closed {
		return nil, errors.New("messagequeue.redis: queue closed")
	}

	// 创建订阅
	pubsub := s.queue.client.Subscribe(ctx, topic)

	// 创建取消上下文
	subCtx, cancel := context.WithCancel(ctx)
	subKey := fmt.Sprintf("sub:%s", topic)
	s.queue.subs[subKey] = cancel

	// 启动接收 goroutine
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

				// 构建消息对象
				message := &contract.Message{
					ID:        "",
					Topic:     topic,
					Body:      []byte(msg.Payload),
					Timestamp: time.Now(),
				}

				// 调用处理函数
				_ = handler(subCtx, message)
			}
		}
	}()

	// 返回取消函数
	return func() error {
		cancel()
		s.queue.mu.Lock()
		delete(s.queue.subs, subKey)
		s.queue.mu.Unlock()
		return pubsub.Close()
	}, nil
}

// SubscribeWithGroup 以消费者组方式订阅（Redis Pub/Sub 不支持，退化为普通订阅）。
func (s *redisSubscriber) SubscribeWithGroup(ctx context.Context, topic string, group string, handler contract.MessageHandler) (contract.UnsubscribeFunc, error) {
	// Redis Pub/Sub 不支持消费者组，退化为普通订阅
	return s.Subscribe(ctx, topic, handler)
}

// Consume 消费队列消息（点对点模式）。
func (s *redisSubscriber) Consume(ctx context.Context, queue string, handler contract.MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 阻塞获取消息
		result, err := s.queue.client.BLPop(ctx, time.Second, queue).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return err
		}

		// 构建消息对象
		message := &contract.Message{
			ID:        "",
			Queue:     queue,
			Body:      []byte(result[1]),
			Timestamp: time.Now(),
		}

		// 调用处理函数
		if err := handler(ctx, message); err != nil {
			// 处理失败，重新放回队列
			s.queue.client.RPush(ctx, queue, result[1])
		}
	}
}

// Unsubscribe 取消所有订阅。
func (s *redisSubscriber) Unsubscribe() error {
	s.queue.mu.Lock()
	defer s.queue.mu.Unlock()

	for _, cancel := range s.queue.subs {
		cancel()
	}
	s.queue.subs = make(map[string]context.CancelFunc)

	return nil
}