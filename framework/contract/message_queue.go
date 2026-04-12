package contract

import (
	"context"
	"time"
)

const (
	// MessageQueueKey 是消息队列在容器中的绑定 key。
	//
	// 中文说明：
	// - 用于异步消息发布和订阅；
	// - 支持 Redis PubSub、RabbitMQ、Kafka 等多种实现；
	// - noop 实现空操作，单体项目零依赖。
	MessageQueueKey = "framework.message_queue"

	// MessagePublisherKey 是消息发布者在容器中的绑定 key。
	MessagePublisherKey = "framework.message_publisher"

	// MessageSubscriberKey 是消息订阅者在容器中的绑定 key。
	MessageSubscriberKey = "framework.message_subscriber"
)

// MessageQueue 消息队列接口。
//
// 中文说明：
	// - 统一的消息队列抽象；
// - 支持发布/订阅模式；
// - 支持点对点模式（Queue）。
type MessageQueue interface {
	// Publisher 返回消息发布者。
	Publisher() MessagePublisher

	// Subscriber 返回消息订阅者。
	Subscriber() MessageSubscriber

	// Close 关闭消息队列连接。
	Close() error
}

// MessagePublisher 消息发布者接口。
//
// 中文说明：
// - 发布消息到指定主题或队列；
// - 支持同步和异步发布；
// - 支持消息属性（如延迟、优先级）。
type MessagePublisher interface {
	// Publish 发布消息到主题。
	//
	// 中文说明：
	// - topic: 主题名称；
	// - message: 消息内容；
	// - options: 发布选项（如延迟时间）。
	Publish(ctx context.Context, topic string, message []byte, options ...PublishOption) error

	// PublishWithDelay 发布延迟消息。
	//
	// 中文说明：
	// - delay: 延迟时间。
	PublishWithDelay(ctx context.Context, topic string, message []byte, delay time.Duration) error

	// PublishWithPriority 发布优先级消息。
	//
	// 中文说明：
	// - priority: 优先级（1-9，数字越大优先级越高）。
	PublishWithPriority(ctx context.Context, topic string, message []byte, priority int) error

	// Send 发送消息到队列（点对点模式）。
	//
	// 中文说明：
	// - queue: 队列名称；
	// - 每条消息只会被一个消费者处理。
	Send(ctx context.Context, queue string, message []byte, options ...PublishOption) error
}

// MessageSubscriber 消息订阅者接口。
//
// 中文说明：
// - 订阅主题或队列；
// - 接收并处理消息；
// - 支持消费确认（ACK）。
type MessageSubscriber interface {
	// Subscribe 订阅主题。
	//
	// 中文说明：
	// - topic: 主题名称；
	// - handler: 消息处理函数；
	// - 返回取消订阅函数。
	Subscribe(ctx context.Context, topic string, handler MessageHandler) (UnsubscribeFunc, error)

	// SubscribeWithGroup 以消费者组方式订阅。
	//
	// 中文说明：
	// - group: 消费者组名称；
	// - 同一组内的消费者共同处理消息（负载均衡）。
	SubscribeWithGroup(ctx context.Context, topic string, group string, handler MessageHandler) (UnsubscribeFunc, error)

	// Consume 消费队列消息（点对点模式）。
	//
	// 中文说明：
	// - queue: 队列名称；
	// - handler: 消息处理函数。
	Consume(ctx context.Context, queue string, handler MessageHandler) error

	// Unsubscribe 取消所有订阅。
	Unsubscribe() error
}

// MessageHandler 消息处理函数。
//
// 中文说明：
// - ctx: 上下文；
// - message: 消息对象；
// - 返回错误表示处理失败，消息将被重试或进入死信队列。
type MessageHandler func(ctx context.Context, message *Message) error

// UnsubscribeFunc 取消订阅函数。
type UnsubscribeFunc func() error

// Message 消息对象。
type Message struct {
	// ID 消息唯一标识
	ID string

	// Topic 主题名称
	Topic string

	// Queue 队列名称（点对点模式）
	Queue string

	// Body 消息体
	Body []byte

	// Headers 消息头
	Headers map[string]string

	// Timestamp 消息时间戳
	Timestamp time.Time

	// Delay 延迟时间
	Delay time.Duration

	// Priority 优先级
	Priority int

	// RetryCount 重试次数
	RetryCount int

	// MaxRetry 最大重试次数
	MaxRetry int
}

// PublishOption 发布选项。
type PublishOption func(*PublishConfig)

// PublishConfig 发布配置。
type PublishConfig struct {
	// Delay 延迟时间
	Delay time.Duration

	// Priority 优先级
	Priority int

	// Headers 消息头
	Headers map[string]string

	// MaxRetry 最大重试次数
	MaxRetry int

	// TTL 消息存活时间
	TTL time.Duration
}

// WithDelay 设置延迟时间。
func WithDelay(delay time.Duration) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.Delay = delay
	}
}

// WithPriority 设置优先级。
func WithPriority(priority int) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.Priority = priority
	}
}

// WithHeaders 设置消息头。
func WithHeaders(headers map[string]string) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.Headers = headers
	}
}

// WithMaxRetry 设置最大重试次数。
func WithMaxRetry(maxRetry int) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.MaxRetry = maxRetry
	}
}

// WithTTL 设置消息存活时间。
func WithTTL(ttl time.Duration) PublishOption {
	return func(cfg *PublishConfig) {
		cfg.TTL = ttl
	}
}

// MessageQueueConfig 消息队列配置。
type MessageQueueConfig struct {
	// Type 队列类型：noop/redis/rabbitmq/kafka
	Type string

	// Redis 配置
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// RabbitMQ 配置
	RabbitMQURL   string
	RabbitMQVHost string

	// Kafka 配置
	KafkaBrokers []string
	KafkaGroupID string

	// 通用配置
	MaxRetry       int           // 最大重试次数
	RetryDelay     time.Duration // 重试延迟
	Timeout        time.Duration // 操作超时
	ConsumerBuffer int           // 消费者缓冲区大小
}