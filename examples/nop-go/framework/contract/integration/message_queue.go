// Application scenarios:
// - Define the message queue contract shared by asynchronous integration features.
// - Standardize publish, subscribe, delayed delivery, priority, and consumer semantics.
// - Provide one reusable config and option model across MQ backends.
//
// 适用场景：
// - 定义异步集成功能共享的消息队列契约。
// - 统一发布、订阅、延迟投递、优先级和消费语义。
// - 为不同 MQ 后端提供统一的配置和选项模型。
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

// MessageQueue combines publisher and subscriber capabilities.
//
// MessageQueue 组合消息发布与订阅能力。
type MessageQueue interface {
	Publisher() MessagePublisher
	Subscriber() MessageSubscriber
	Close() error
}

// MessagePublisher defines the outbound message publishing contract.
//
// MessagePublisher 定义出站消息发布契约。
type MessagePublisher interface {
	Publish(ctx context.Context, topic string, message []byte, options ...PublishOption) error
	PublishWithDelay(ctx context.Context, topic string, message []byte, delay time.Duration) error
	PublishWithPriority(ctx context.Context, topic string, message []byte, priority int) error
	Send(ctx context.Context, queue string, message []byte, options ...PublishOption) error
}

// MessageSubscriber defines the inbound message consumption contract.
//
// MessageSubscriber 定义入站消息消费契约。
type MessageSubscriber interface {
	Subscribe(ctx context.Context, topic string, handler MessageHandler) (UnsubscribeFunc, error)
	SubscribeWithGroup(ctx context.Context, topic string, group string, handler MessageHandler) (UnsubscribeFunc, error)
	Consume(ctx context.Context, queue string, handler MessageHandler) error
	// UnsubscribeAll closes all active subscriptions and releases resources.
	// For canceling a single subscription, use the UnsubscribeFunc returned by Subscribe/SubscribeWithGroup.
	//
	// UnsubscribeAll 关闭所有活跃订阅并释放资源。
	// 若需取消单个订阅，请使用 Subscribe/SubscribeWithGroup 返回的 UnsubscribeFunc。
	UnsubscribeAll() error
}

// MessageHandler handles one inbound message.
//
// MessageHandler 定义单条消息处理器。
type MessageHandler func(ctx context.Context, message *Message) error

// UnsubscribeFunc cancels one subscription.
//
// UnsubscribeFunc 用于取消一条订阅。
type UnsubscribeFunc func() error

// Message describes one queue message.
//
// Message 描述一条队列消息。
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

// PublishOption mutates publish config.
//
// PublishOption 用于修改发布配置。
type PublishOption func(*PublishConfig)

// PublishConfig describes message publishing options.
//
// PublishConfig 描述消息发布选项。
type PublishConfig struct {
	Delay    time.Duration
	Priority int
	Headers  map[string]string
	MaxRetry int
	TTL      time.Duration
}

// WithDelay sets message publish delay.
//
// WithDelay 设置消息发布延迟。
func WithDelay(delay time.Duration) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.Delay = delay
	}
}

// WithPriority sets message priority.
//
// WithPriority 设置消息优先级。
func WithPriority(priority int) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.Priority = priority
	}
}

// WithHeaders sets custom message headers.
//
// WithHeaders 设置自定义消息头。
func WithHeaders(headers map[string]string) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.Headers = headers
	}
}

// WithMaxRetry sets the message retry limit.
//
// WithMaxRetry 设置消息最大重试次数。
func WithMaxRetry(maxRetry int) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.MaxRetry = maxRetry
	}
}

// WithTTL sets the message time-to-live.
//
// WithTTL 设置消息生存时间。
func WithTTL(ttl time.Duration) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.TTL = ttl
	}
}

// MessageQueueConfig describes message queue runtime configuration.
//
// MessageQueueConfig 描述消息队列运行时配置。
type MessageQueueConfig struct {
	Type string

	// Redis configuration
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Kafka configuration (using IBM/sarama SDK)
	KafkaBrokers         []string
	KafkaGroupID         string
	KafkaClientID        string
	KafkaVersion         string // e.g., "2.8.0"
	KafkaCompression     string // "none", "gzip", "snappy", "lz4", "zstd"
	KafkaPartitioner     string // "hash", "random", "round-robin"
	KafkaRequiredACKs    int    // 0=NoResponse, 1=Leader, -1=All
	KafkaMaxMessageBytes int
	KafkaFlushFrequency  time.Duration
	KafkaEnableTLS       bool
	KafkaTLSCertFile     string
	KafkaTLSKeyFile      string
	KafkaTLSCACertFile   string

	// RabbitMQ configuration (using amqp091-go SDK)
	RabbitMQURL          string // "amqp://guest:guest@localhost:5672/"
	RabbitMQVHost        string
	RabbitMQExchange     string
	RabbitMQExchangeType string // "direct", "fanout", "topic", "headers"
	RabbitMQQueuePrefix  string
	RabbitMQPrefetch     int
	RabbitMQEnableTLS    bool

	// RocketMQ configuration (using apache/rocketmq-client-go SDK)
	RocketMQNamesrvAddr  string // "localhost:9876"
	RocketMQGroupName    string
	RocketMQInstanceName string
	RocketMQRetryTimes   int
	RocketMQEnableTLS    bool

	// Common configuration
	MaxRetry       int
	RetryDelay     time.Duration
	Timeout        time.Duration
	ConsumerBuffer int
}

// NativeMQClientProvider is an optional interface that MessageQueue implementations
// can satisfy to expose the underlying native MQ client for advanced usage.
// This allows "MQ-first" users to access native SDK capabilities while staying
// within the framework's governance boundary.
//
// NativeMQClientProvider 是 MessageQueue 实现可满足的可选接口，
// 用于暴露底层原生 MQ 客户端供高级使用。
// 这允许"MQ-first"用户访问原生 SDK 能力，同时保持在框架的治理边界内。
type NativeMQClientProvider interface {
	// NativeMQClient returns the underlying MQ client instance.
	// The concrete type depends on the implementation:
	//   - redis: *redis.Client
	//   - kafka: sarama.Client
	//   - rabbitmq: *amqp.Connection
	//   - rocketmq: rocketmq.Producer or rocketmq.Consumer
	//
	// NativeMQClient 返回底层 MQ 客户端实例。
	// 具体类型取决于实现。
	NativeMQClient() any
}

// NativePublisherProvider is an optional interface for accessing
// the underlying native publisher client.
//
// NativePublisherProvider 是访问底层原生发布者客户端的可选接口。
type NativePublisherProvider interface {
	// NativePublisher returns the underlying native publisher.
	// The concrete type depends on the implementation:
	//   - redis: *redis.Client
	//   - kafka: sarama.SyncProducer or sarama.AsyncProducer
	//   - rabbitmq: *amqp.Channel
	//   - rocketmq: rocketmq.Producer
	NativePublisher() any
}

// NativeSubscriberProvider is an optional interface for accessing
// the underlying native subscriber client.
//
// NativeSubscriberProvider 是访问底层原生订阅者客户端的可选接口。
type NativeSubscriberProvider interface {
	// NativeSubscriber returns the underlying native subscriber.
	// The concrete type depends on the implementation:
	//   - redis: *redis.PubSub
	//   - kafka: sarama.ConsumerGroup
	//   - rabbitmq: *amqp.Channel
	//   - rocketmq: rocketmq.PushConsumer or rocketmq.PullConsumer
	NativeSubscriber() any
}
