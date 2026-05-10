# contrib/messagequeue/rocketmq

RocketMQ 消息队列 provider，使用 apache/rocketmq-client-go SDK 实现。

## SDK

使用 [github.com/apache/rocketmq-client-go/v2](https://github.com/apache/rocketmq-client-go)（Apache 官方维护）：
- 支持 RocketMQ 5.x
- 支持 Producer/Consumer/PushConsumer/PullConsumer
- Apache 官方维护

## 配置

```yaml
message_queue:
  rocketmq:
    namesrv_addr: "localhost:9876"
    group_name: "my-consumer-group"
    instance_name: ""           # 可选
    retry_times: 3              # 发送重试次数
    enable_tls: false
```

## 使用

```go
// 标准抽象路径
mq := c.MustMake(integrationcontract.MessageQueueKey).(integrationcontract.MessageQueue)
mq.Publisher().Publish(ctx, "topic", message)

// 消费者订阅
subscriber := mq.Subscriber()
unsubscribe, err := subscriber.SubscribeWithGroup(ctx, "topic", "my-group", handler)
```

## 下探原生 SDK

```go
mq := c.MustMake(integrationcontract.MessageQueueKey).(integrationcontract.MessageQueue)

// 获取 Producer
pub := mq.Publisher()
if producer, ok := messagequeue.NativeRocketMQProducer(pub); ok {
    // 使用 RocketMQ 高级特性
    
    // 顺序消息
    msg := &primitive.Message{Topic: "topic", Body: message}
    msg.WithShardingKey("order-123") // 分片键
    result, _ := producer.SendSync(ctx, msg)
    
    // 事务消息
    // (需要创建事务 producer)
    
    // 批量发送
    msgs := []*primitive.Message{{Topic: "topic", Body: msg1}, {Topic: "topic", Body: msg2}}
    producer.SendSync(ctx, msgs...)
}
```

## 特性适配

| 契约方法 | RocketMQ 实现 | 说明 |
|---------|--------------|------|
| Publish | SendSync | 同步发送 |
| PublishWithDelay | DelayTimeLevel | 固定 18 级延迟 |
| PublishWithPriority | Tag | 无原生优先级，用 Tag 区分 |
| Send | Topic:Tag | queue 视为 topic:tag |
| Subscribe | PushConsumer | 订阅 topic |
| SubscribeWithGroup | Consumer Group | 推荐方式 |
| Consume | 返回错误 | 不支持，使用 SubscribeWithGroup |

## 延迟消息等级

RocketMQ 支持固定延迟等级（1-18）：

| 等级 | 延迟时间 |
|-----|---------|
| 1 | 1s |
| 2 | 5s |
| 3 | 10s |
| 4 | 30s |
| 5 | 1m |
| 6 | 2m |
| 7 | 3m |
| 8 | 4m |
| 9 | 5m |
| 10 | 6m |
| 11 | 7m |
| 12 | 8m |
| 13 | 9m |
| 14 | 10m |
| 15 | 20m |
| 16 | 30m |
| 17 | 1h |
| 18 | 2h |

## 注意事项

1. **延迟消息**：使用固定等级，无法精确指定延迟时间
2. **顺序消息**：通过下探获取 Producer 后使用 `WithShardingKey`
3. **事务消息**：需要单独创建事务 Producer（通过下探）
4. **Tag 过滤**：订阅时可指定 Tag 表达式过滤消息