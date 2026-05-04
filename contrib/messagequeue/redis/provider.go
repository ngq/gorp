package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "messagequeue.redis" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{
		integrationcontract.MessageQueueKey,
		integrationcontract.MessagePublisherKey,
		integrationcontract.MessageSubscriberKey,
	}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.MessageQueueKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getMQConfig(c)
		if err != nil {
			return nil, err
		}
		return NewQueue(cfg)
	}, true)

	c.Bind(integrationcontract.MessagePublisherKey, func(c runtimecontract.Container) (any, error) {
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

	c.Bind(integrationcontract.MessageSubscriberKey, func(c runtimecontract.Container) (any, error) {
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

func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

func getMQConfig(c runtimecontract.Container) (*integrationcontract.MessageQueueConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("messagequeue: invalid config service")
	}

	mqCfg := &integrationcontract.MessageQueueConfig{
		Type:           "redis",
		MaxRetry:       3,
		RetryDelay:     time.Second,
		Timeout:        5 * time.Second,
		ConsumerBuffer: 100,
	}
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
	if maxRetry := cfg.GetInt("message_queue.max_retry"); maxRetry > 0 {
		mqCfg.MaxRetry = maxRetry
	}
	return mqCfg, nil
}

type Queue struct {
	cfg    *integrationcontract.MessageQueueConfig
	client *redis.Client
	pubsub *redis.PubSub
	mu     sync.Mutex
	subs   map[string]context.CancelFunc
	closed bool
}

func NewQueue(cfg *integrationcontract.MessageQueueConfig) (*Queue, error) {
	client := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("messagequeue.redis: connect failed: %w", err)
	}
	return &Queue{cfg: cfg, client: client, subs: make(map[string]context.CancelFunc)}, nil
}

func (q *Queue) Publisher() integrationcontract.MessagePublisher {
	return &redisPublisher{queue: q}
}

func (q *Queue) Subscriber() integrationcontract.MessageSubscriber {
	return &redisSubscriber{queue: q}
}

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

type redisPublisher struct {
	queue *Queue
}

func (p *redisPublisher) Publish(ctx context.Context, topic string, message []byte, options ...integrationcontract.PublishOption) error {
	cfg := &integrationcontract.PublishConfig{}
	for _, opt := range options {
		opt(cfg)
	}
	if p == nil || p.queue == nil || p.queue.client == nil {
		return errors.New("messagequeue.redis: client not initialized")
	}
	return p.queue.client.Publish(ctx, topic, message).Err()
}

func (p *redisPublisher) PublishWithDelay(ctx context.Context, topic string, message []byte, delay time.Duration) error {
	if p == nil || p.queue == nil || p.queue.client == nil {
		return errors.New("messagequeue.redis: client not initialized")
	}
	score := float64(time.Now().Add(delay).Unix())
	key := fmt.Sprintf("delay:%s", topic)
	return p.queue.client.ZAdd(ctx, key, redis.Z{Score: score, Member: message}).Err()
}

func (p *redisPublisher) PublishWithPriority(ctx context.Context, topic string, message []byte, priority int) error {
	if p == nil || p.queue == nil || p.queue.client == nil {
		return errors.New("messagequeue.redis: client not initialized")
	}
	queueName := fmt.Sprintf("priority:%s:%d", topic, priority)
	return p.queue.client.LPush(ctx, queueName, message).Err()
}

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

type redisSubscriber struct {
	queue *Queue
}

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

func (s *redisSubscriber) SubscribeWithGroup(ctx context.Context, topic string, group string, handler integrationcontract.MessageHandler) (integrationcontract.UnsubscribeFunc, error) {
	_ = group
	return s.Subscribe(ctx, topic, handler)
}

func (s *redisSubscriber) Consume(ctx context.Context, queue string, handler integrationcontract.MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		result, err := s.queue.client.BLPop(ctx, time.Second, queue).Result()
		if err != nil {
			if err == redis.Nil {
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

func (s *redisSubscriber) Unsubscribe() error {
	s.queue.mu.Lock()
	defer s.queue.mu.Unlock()
	for _, cancel := range s.queue.subs {
		cancel()
	}
	s.queue.subs = make(map[string]context.CancelFunc)
	return nil
}
