# contrib/messagequeue/rabbitmq

RabbitMQ 消息队列 provider，使用 amqp091-go SDK 实现。

## SDK

使用 [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go)（RabbitMQ 官方维护）：
- AMQP 0-9-1 协议完整支持
- 稳定、文档完善
- RabbitMQ 官方维护

## 配置

```yaml
message_queue:
  rabbitmq:
    url: "amqp://guest:guest@localhost:5672/"
    vhost: "/"
    exchange: "my-exchange"      # 可选，不配置则直接发布到队列
    exchange_type: "topic"       # direct, fanout, topic, headers
    queue_prefix: ""             # 队列名前缀
    prefetch: 10                 # 消费者 QoS
    enable_tls: false
```

## 使用

```go
// 标准抽象路径
mq := c.MustMake(integrationcontract.MessageQueueKey).(integrationcontract.MessageQueue)
mq.Publisher().Publish(ctx, "topic", message)

// 消费者订阅
subscriber := mq.Subscriber()
unsubscribe, err := subscriber.Subscribe(ctx, "topic", handler)
```

## 下探原生 SDK

```go
mq := c.MustMake(integrationcontract.MessageQueueKey).(integrationcontract.MessageQueue)

// 获取 Connection
if conn, ok := messagequeue.NativeRabbitMQConnection(mq); ok {
    // 创建新 Channel 使用高级特性
    ch, _ := conn.Channel()
    
    // 启用 publisher confirms
    ch.Confirm(false)
    confirms := ch.NotifyPublish(make(chan amqp.Confirmation))
    
    // 发送消息并等待确认
    ch.PublishWithContext(ctx, "exchange", "routing-key", false, false, amqp.Publishing{Body: message})
    <-confirms
}

// 获取现有 Channel
pub := mq.Publisher()
if ch, ok := messagequeue.NativeRabbitMQChannel(pub); ok {
    // 使用 Channel 高级特性
    ch.TxSelect() // 启用事务
    ch.PublishWithContext(ctx, "", "queue", false, false, amqp.Publishing{Body: message})
    ch.TxCommit()
}
```

## 特性适配

| 契约方法 | RabbitMQ 实现 | 说明 |
|---------|--------------|------|
| Publish | Exchange.Publish | routing key = topic |
| PublishWithDelay | x-delayed-message header | 需插件或死信队列 |
| PublishWithPriority | Priority 字段 | 需队列声明 x-max-priority |
| Send | Queue.Publish | 直接队列模式 |
| Subscribe | Queue.Consume | 绑定 Exchange 到 Queue |
| SubscribeWithGroup | 唯一队列名 | 每组独立队列 |
| Consume | Queue.Consume | 队列消费模式 |

## 注意事项

1. **延迟消息**：需要 RabbitMQ delayed message 插件或使用死信队列 + TTL 实现
2. **优先级消息**：队列声明时需设置 `x-max-priority` 参数
3. **消息确认**：通过下探获取 Channel 后可启用 publisher confirms 或事务
4. **Exchange 类型**：支持 direct/fanout/topic/headers，推荐 topic 用于路由