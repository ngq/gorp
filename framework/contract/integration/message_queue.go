package integration

import (
	"context"
	"time"
)

const (
	MessageQueueKey      = "framework.message_queue"
	MessagePublisherKey  = "framework.message_publisher"
	MessageSubscriberKey = "framework.message_subscriber"
)

type MessageQueue interface {
	Publisher() MessagePublisher
	Subscriber() MessageSubscriber
	Close() error
}

type MessagePublisher interface {
	Publish(ctx context.Context, topic string, message []byte, options ...PublishOption) error
	PublishWithDelay(ctx context.Context, topic string, message []byte, delay time.Duration) error
	PublishWithPriority(ctx context.Context, topic string, message []byte, priority int) error
	Send(ctx context.Context, queue string, message []byte, options ...PublishOption) error
}

type MessageSubscriber interface {
	Subscribe(ctx context.Context, topic string, handler MessageHandler) (UnsubscribeFunc, error)
	SubscribeWithGroup(ctx context.Context, topic string, group string, handler MessageHandler) (UnsubscribeFunc, error)
	Consume(ctx context.Context, queue string, handler MessageHandler) error
	Unsubscribe() error
}

type MessageHandler func(ctx context.Context, message *Message) error

type UnsubscribeFunc func() error

type Message struct {
	ID         string
	Topic      string
	Queue      string
	Body       []byte
	Headers    map[string]string
	Timestamp  time.Time
	Delay      time.Duration
	Priority   int
	RetryCount int
	MaxRetry   int
}

type PublishOption func(*PublishConfig)

type PublishConfig struct {
	Delay    time.Duration
	Priority int
	Headers  map[string]string
	MaxRetry int
	TTL      time.Duration
}

func WithDelay(delay time.Duration) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.Delay = delay
	}
}

func WithPriority(priority int) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.Priority = priority
	}
}

func WithHeaders(headers map[string]string) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.Headers = headers
	}
}

func WithMaxRetry(maxRetry int) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.MaxRetry = maxRetry
	}
}

func WithTTL(ttl time.Duration) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.TTL = ttl
	}
}

type MessageQueueConfig struct {
	Type string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	RabbitMQURL   string
	RabbitMQVHost string

	KafkaBrokers []string
	KafkaGroupID string

	MaxRetry       int
	RetryDelay     time.Duration
	Timeout        time.Duration
	ConsumerBuffer int
}
