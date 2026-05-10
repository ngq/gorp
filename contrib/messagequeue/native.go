// Package messagequeue provides helper functions for accessing native MQ clients.
// These functions allow "MQ-first" users to use native SDK capabilities
// while remaining within the framework's governance boundary.
//
// 本包提供访问原生 MQ 客户端的辅助函数。
// 这些函数允许"MQ-first"用户使用原生 SDK 能力，
// 同时保持在框架的治理边界内。
package messagequeue

import (
	"github.com/IBM/sarama"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"github.com/apache/rocketmq-client-go/v2"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// NativeRedisClient extracts the underlying *redis.Client from a MessageQueue.
// Returns the client and true if the MQ implementation is Redis-backed.
// Returns nil and false otherwise.
//
// NativeRedisClient 从 MessageQueue 中提取底层 *redis.Client。
// 如果 MQ 实现基于 Redis，返回 client 和 true。
// 否则返回 nil 和 false。
//
// Example:
//
//	mq := c.MustMake(integrationcontract.MessageQueueKey).(integrationcontract.MessageQueue)
//	if client, ok := messagequeue.NativeRedisClient(mq); ok {
//	    // 使用 Redis SDK 高级特性
//	    client.XAdd(ctx, &redis.XAddArgs{Stream: "events", Values: data})
//	    client.Eval(ctx, luaScript, keys, args)
//	}
func NativeRedisClient(mq integrationcontract.MessageQueue) (*redis.Client, bool) {
	provider, ok := mq.(integrationcontract.NativeMQClientProvider)
	if !ok {
		return nil, false
	}
	clientAny := provider.NativeMQClient()
	if clientAny == nil {
		return nil, false
	}
	client, ok := clientAny.(*redis.Client)
	return client, ok
}

// NativeRedisPublisher extracts *redis.Client from MessagePublisher.
// Returns the client and true if the publisher implementation is Redis-backed.
//
// NativeRedisPublisher 从 MessagePublisher 中提取 *redis.Client。
// 如果发布者实现基于 Redis，返回 client 和 true。
func NativeRedisPublisher(pub integrationcontract.MessagePublisher) (*redis.Client, bool) {
	provider, ok := pub.(integrationcontract.NativePublisherProvider)
	if !ok {
		return nil, false
	}
	clientAny := provider.NativePublisher()
	if clientAny == nil {
		return nil, false
	}
	client, ok := clientAny.(*redis.Client)
	return client, ok
}

// NativeRedisPubSub extracts *redis.PubSub from MessageSubscriber.
// Returns the pubsub and true if the subscriber implementation is Redis-backed.
//
// NativeRedisPubSub 从 MessageSubscriber 中提取 *redis.PubSub。
// 如果订阅者实现基于 Redis，返回 pubsub 和 true。
func NativeRedisPubSub(sub integrationcontract.MessageSubscriber) (*redis.PubSub, bool) {
	provider, ok := sub.(integrationcontract.NativeSubscriberProvider)
	if !ok {
		return nil, false
	}
	pubsubAny := provider.NativeSubscriber()
	if pubsubAny == nil {
		return nil, false
	}
	pubsub, ok := pubsubAny.(*redis.PubSub)
	return pubsub, ok
}

// NativeKafkaClient extracts the underlying sarama.Client from a MessageQueue.
// Returns the client and true if the MQ implementation is Kafka-backed.
// Returns nil and false otherwise.
//
// NativeKafkaClient 从 MessageQueue 中提取底层 sarama.Client。
// 如果 MQ 实现基于 Kafka，返回 client 和 true。
// 否则返回 nil 和 false。
//
// Example:
//
//	mq := c.MustMake(integrationcontract.MessageQueueKey).(integrationcontract.MessageQueue)
//	if client, ok := messagequeue.NativeKafkaClient(mq); ok {
//	    // 使用 Sarama SDK 高级特性
//	    topics, _ := client.Topics()
//	    partitions, _ := client.Partitions("my-topic")
//	}
func NativeKafkaClient(mq integrationcontract.MessageQueue) (sarama.Client, bool) {
	provider, ok := mq.(integrationcontract.NativeMQClientProvider)
	if !ok {
		return nil, false
	}
	clientAny := provider.NativeMQClient()
	if clientAny == nil {
		return nil, false
	}
	client, ok := clientAny.(sarama.Client)
	return client, ok
}

// NativeKafkaProducer extracts sarama.SyncProducer from MessagePublisher.
// Returns the producer and true if the publisher implementation is Kafka-backed.
//
// NativeKafkaProducer 从 MessagePublisher 中提取 sarama.SyncProducer。
// 如果发布者实现基于 Kafka，返回 producer 和 true。
func NativeKafkaProducer(pub integrationcontract.MessagePublisher) (sarama.SyncProducer, bool) {
	provider, ok := pub.(integrationcontract.NativePublisherProvider)
	if !ok {
		return nil, false
	}
	producerAny := provider.NativePublisher()
	if producerAny == nil {
		return nil, false
	}
	// 支持 SyncProducer 和 AsyncProducer
	if syncProducer, ok := producerAny.(sarama.SyncProducer); ok {
		return syncProducer, true
	}
	return nil, false
}

// NativeKafkaConsumerGroup extracts sarama.ConsumerGroup from MessageSubscriber.
// Returns the consumer group and true if the subscriber implementation is Kafka-backed.
//
// NativeKafkaConsumerGroup 从 MessageSubscriber 中提取 sarama.ConsumerGroup。
// 如果订阅者实现基于 Kafka，返回 consumerGroup 和 true。
func NativeKafkaConsumerGroup(sub integrationcontract.MessageSubscriber) (sarama.ConsumerGroup, bool) {
	provider, ok := sub.(integrationcontract.NativeSubscriberProvider)
	if !ok {
		return nil, false
	}
	consumerAny := provider.NativeSubscriber()
	if consumerAny == nil {
		return nil, false
	}
	consumerGroup, ok := consumerAny.(sarama.ConsumerGroup)
	return consumerGroup, ok
}

// NativeRabbitMQConnection extracts the underlying *amqp.Connection from a MessageQueue.
// Returns the connection and true if the MQ implementation is RabbitMQ-backed.
// Returns nil and false otherwise.
//
// NativeRabbitMQConnection 从 MessageQueue 中提取底层 *amqp.Connection。
// 如果 MQ 实现基于 RabbitMQ，返回 connection 和 true。
// 否则返回 nil 和 false。
//
// Example:
//
//	mq := c.MustMake(integrationcontract.MessageQueueKey).(integrationcontract.MessageQueue)
//	if conn, ok := messagequeue.NativeRabbitMQConnection(mq); ok {
//	    // 创建新 Channel 使用高级特性
//	    ch, _ := conn.Channel()
//	    ch.Confirm(false) // 启用 publisher confirms
//	}
func NativeRabbitMQConnection(mq integrationcontract.MessageQueue) (*amqp.Connection, bool) {
	provider, ok := mq.(integrationcontract.NativeMQClientProvider)
	if !ok {
		return nil, false
	}
	connAny := provider.NativeMQClient()
	if connAny == nil {
		return nil, false
	}
	conn, ok := connAny.(*amqp.Connection)
	return conn, ok
}

// NativeRabbitMQChannel extracts *amqp.Channel from MessagePublisher or MessageSubscriber.
// Returns the channel and true if the implementation is RabbitMQ-backed.
//
// NativeRabbitMQChannel 从 MessagePublisher 或 MessageSubscriber 中提取 *amqp.Channel。
// 如果实现基于 RabbitMQ，返回 channel 和 true。
func NativeRabbitMQChannel(pubOrSub any) (*amqp.Channel, bool) {
	switch v := pubOrSub.(type) {
	case integrationcontract.NativePublisherProvider:
		channelAny := v.NativePublisher()
		if channelAny == nil {
			return nil, false
		}
		channel, ok := channelAny.(*amqp.Channel)
		return channel, ok
	case integrationcontract.NativeSubscriberProvider:
		channelAny := v.NativeSubscriber()
		if channelAny == nil {
			return nil, false
		}
		channel, ok := channelAny.(*amqp.Channel)
		return channel, ok
	default:
		return nil, false
	}
}

// NativeRocketMQProducer extracts rocketmq.Producer from MessageQueue or MessagePublisher.
// Returns the producer and true if the implementation is RocketMQ-backed.
//
// NativeRocketMQProducer 从 MessageQueue 或 MessagePublisher 中提取 rocketmq.Producer。
// 如果实现基于 RocketMQ，返回 producer 和 true。
//
// Note: The returned type is rocketmq.Producer interface from rocketmq-client-go.
func NativeRocketMQProducer(mqOrPub any) (rocketmq.Producer, bool) {
	switch v := mqOrPub.(type) {
	case integrationcontract.NativeMQClientProvider:
		producerAny := v.NativeMQClient()
		if producerAny == nil {
			return nil, false
		}
		p, ok := producerAny.(rocketmq.Producer)
		return p, ok
	case integrationcontract.NativePublisherProvider:
		producerAny := v.NativePublisher()
		if producerAny == nil {
			return nil, false
		}
		p, ok := producerAny.(rocketmq.Producer)
		return p, ok
	default:
		return nil, false
	}
}