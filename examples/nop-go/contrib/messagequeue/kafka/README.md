# contrib/messagequeue/kafka

Kafka 消息队列 provider，使用 IBM/sarama SDK 实现。

## SDK

使用 [github.com/IBM/sarama](https://github.com/IBM/sarama)（原 Shopify/sarama）：
- 纯 Go 实现，无 C 绑定
- 成熟稳定，广泛使用
- 支持 Kafka 0.10+ 特性

## 配置

```yaml
message_queue:
  kafka:
    brokers:
      - localhost:9092
    group_id: my-consumer-group
    client_id: my-service
    version: "2.8.0"
    compression: gzip      # none, gzip, snappy, lz4, zstd
    partitioner: hash      # hash, random, round-robin
    required_acks: -1      # 0=NoResponse, 1=Leader, -1=All
    max_message_bytes: 1000000
    flush_frequency_ms: 100
    enable_tls: false
```

## 使用

```go
// 标准抽象路径
mq := c.MustMake(integrationcontract.MessageQueueKey).(integrationcontract.MessageQueue)
mq.Publisher().Publish(ctx, "topic", message)

// 消费者组订阅
subscriber := mq.Subscriber()
unsubscribe, err := subscriber.SubscribeWithGroup(ctx, "topic", "my-group", handler)
```

## 下探原生 SDK

```go
mq := c.MustMake(integrationcontract.MessageQueueKey).(integrationcontract.MessageQueue)

// 获取 sarama.Client
if client, ok := messagequeue.NativeKafkaClient(mq); ok {
    // 使用 Sarama 高级特性
    topics, _ := client.Topics()
    partitions, _ := client.Partitions("my-topic")
    
    // 创建 admin client
    admin, _ := sarama.NewClusterAdminFromClient(client)
    admin.CreateTopic("new-topic", &sarama.TopicDetail{}, false)
}

// 获取 SyncProducer
pub := mq.Publisher()
if producer, ok := messagequeue.NativeKafkaProducer(pub); ok {
    // 使用 producer 高级特性
    msg := &sarama.ProducerMessage{
        Topic: "topic",
        Key:   sarama.StringEncoder("key"),
        Value: sarama.ByteEncoder(message),
    }
    partition, offset, err := producer.SendMessage(msg)
}
```

## 特性适配

| 契约方法 | Kafka 实现 | 说明 |
|---------|-----------|------|
| Publish | SendMessage | 同步发送 |
| PublishWithDelay | 返回错误 | Kafka 不支持，建议外置调度 |
| PublishWithPriority | 分区路由 | 通过 partition key 实现 |
| Send | SendMessage | 视 queue 为 topic |
| Subscribe | ConsumerGroup | 自动创建唯一组名 |
| SubscribeWithGroup | ConsumerGroup | 推荐方式 |
| Consume | 返回错误 | 不支持，使用 SubscribeWithGroup |

## 注意事项

1. **延迟消息**：Kafka 不原生支持延迟消息，建议使用 Redis 延迟队列 + 定时投递
2. **消费者组**：推荐使用 SubscribeWithGroup，保证消息可靠消费
3. **消息顺序**：同一 partition 内保证顺序，使用 partition key 控制
4. **事务**：通过下探获取 client 后可使用 Kafka 事务特性