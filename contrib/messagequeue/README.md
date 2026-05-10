# contrib/messagequeue

该目录承接 gorp 的消息队列后端实现，支持多 MQ 后端统一抽象与下探原生 SDK。

## 当前状态

| Provider | 状态 | SDK |
|----------|------|-----|
| `redis` | ✅ 已完成 | github.com/redis/go-redis/v9 |
| `kafka` | ✅ 已完成 | github.com/IBM/sarama |
| `rabbitmq` | ✅ 已完成 | github.com/rabbitmq/amqp091-go |
| `rocketmq` | ✅ 已完成 | github.com/apache/rocketmq-client-go/v2 |

## 下探机制

所有 provider 都支持下探原生 SDK，允许用户使用高级特性：

```go
// Redis 下探
if client, ok := messagequeue.NativeRedisClient(mq); ok {
    client.XAdd(ctx, &redis.XAddArgs{Stream: "events", Values: data})
    client.Eval(ctx, luaScript, keys, args)
}

// Kafka 下探
if client, ok := messagequeue.NativeKafkaClient(mq); ok {
    topics, _ := client.Topics()
    admin, _ := sarama.NewClusterAdminFromClient(client)
}

// RabbitMQ 下探
if conn, ok := messagequeue.NativeRabbitMQConnection(mq); ok {
    ch, _ := conn.Channel()
    ch.Confirm(false) // 启用 publisher confirms
}

// RocketMQ 下探
if producer, ok := messagequeue.NativeRocketMQProducer(pub); ok {
    msg.WithShardingKey("order-123") // 顺序消息
}
```

## 统一契约

所有 provider 实现统一契约：

- `MessageQueue`：组合 Publisher + Subscriber
- `MessagePublisher`：Publish / PublishWithDelay / PublishWithPriority / Send
- `MessageSubscriber`：Subscribe / SubscribeWithGroup / Consume / Unsubscribe

## 各 Provider 特性

### Redis
- 轻量级，适合简单场景
- 支持 Pub/Sub、Stream、延迟队列（ZSet）
- 下探可用：pipeline、Lua 脚本、事务

### Kafka
- 高吞吐量，适合大规模场景
- 支持 Consumer Group
- 延迟消息不支持（需外置调度）
- 下探可用：事务、压缩、自定义分区器

### RabbitMQ
- 功能丰富，支持 Exchange/Queue 模型
- 支持 x-delayed-message 插件或死信队列延迟
- 支持优先级队列（需声明 x-max-priority）
- 下探可用：publisher confirms、事务、自定义 Channel

### RocketMQ
- 支持固定延迟等级（1-18）
- 支持顺序消息、事务消息
- 下探可用：分片键、批量发送、Tag 过滤

## 配置示例

```yaml
message_queue:
  # Redis
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0

  # Kafka
  kafka:
    brokers:
      - "localhost:9092"
    group_id: "my-group"
    client_id: "my-service"
    version: "2.8.0"

  # RabbitMQ
  rabbitmq:
    url: "amqp://guest:guest@localhost:5672/"
    exchange: "my-exchange"
    exchange_type: "topic"

  # RocketMQ
  rocketmq:
    namesrv_addr: "localhost:9876"
    group_name: "my-group"
```

## 文件结构

```
contrib/messagequeue/
├── native.go              # 统一下探辅助函数
├── README.md              # 本文档
│
├── redis/
│   ├── provider.go        # Redis Provider
│   ├── provider_test.go
│   ├── behavior_test.go
│   └── README.md
│
├── kafka/
│   ├── provider.go        # Kafka Provider
│   ├── config.go          # Sarama 配置构建
│   ├── provider_test.go
│   ├── behavior_test.go
│   └── README.md
│
├── rabbitmq/
│   ├── provider.go        # RabbitMQ Provider
│   ├── provider_test.go
│   ├── behavior_test.go
│   └── README.md
│
└── rocketmq/
│   ├── provider.go        # RocketMQ Provider
│   ├── provider_test.go
│   ├── behavior_test.go
│   └── README.md
```

## P0 约束

- 所有 provider 必须实现 `NativeMQClientProvider` 可选接口
- 所有 provider 必须提供 `Underlying()` + `As()` 方法
- 所有 provider 必须有 provider_test.go + behavior_test.go 测试
- 文档必须包含 SDK 选择说明、配置示例、下探示例