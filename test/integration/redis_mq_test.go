// Package integration provides Redis message queue integration tests.
//
// 本包提供 Redis 消息队列集成测试。
package integration

import (
	"context"
	"testing"
	"time"

	redismq "github.com/ngq/gorp/contrib/messagequeue/redis"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// TestRedisPubSub tests real Redis Pub/Sub operations.
//
// TestRedisPubSub 测试真实 Redis Pub/Sub 操作。
func TestRedisPubSub(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Redis backend")
	}

	ctx := context.Background()

	// 1. Create Redis MQ
	// 创建 Redis MQ
	cfg := &integrationcontract.MessageQueueConfig{
		Type:           "redis",
		RedisAddr:      getEnvOrDefault("GORP_TEST_REDIS_ADDR", "localhost:6379"),
		MaxRetry:       3,
		RetryDelay:     time.Second,
		Timeout:        5 * time.Second,
		ConsumerBuffer: 10,
	}

	queue, err := redismq.NewQueue(cfg)
	if err != nil {
		t.Fatalf("failed to create Redis queue: %v", err)
	}
	defer queue.Close()

	publisher := queue.Publisher()
	subscriber := queue.Subscriber()

	testTopic := "test-topic-integration"

	// 2. Subscribe to topic
	// 订阅 topic
	received := make(chan []byte, 1)
	unsubscribe, err := subscriber.Subscribe(ctx, testTopic, func(ctx context.Context, msg *integrationcontract.Message) error {
		received <- msg.Body
		t.Logf("received message: topic=%s, id=%s, body=%s", msg.Topic, msg.ID, string(msg.Body))
		return nil
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	t.Logf("subscribed to topic: %s", testTopic)

	// 3. Publish message
	// 发布消息
	testMessage := "hello-redis-integration"
	err = publisher.Publish(ctx, testTopic, []byte(testMessage))
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	t.Logf("published message: %s", testMessage)

	// 4. Wait for message reception
	// 等待消息接收
	select {
	case body := <-received:
		if string(body) != testMessage {
			t.Fatalf("message mismatch: expected=%s, got=%s", testMessage, string(body))
		}
		t.Log("verified: received correct message")
	case <-time.After(5 * time.Second):
		t.Fatal("subscribe timeout: did not receive message within 5s")
	}

	// 5. Cleanup
	// 清理
	unsubscribe()
}

// TestRedisSendConsume tests direct queue send/consume (list-based queue).
//
// TestRedisSendConsume 测试直接队列发送/消费（基于列表的队列）。
func TestRedisSendConsume(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Redis backend")
	}

	ctx := context.Background()

	cfg := &integrationcontract.MessageQueueConfig{
		Type:           "redis",
		RedisAddr:      getEnvOrDefault("GORP_TEST_REDIS_ADDR", "localhost:6379"),
		MaxRetry:       3,
		RetryDelay:     time.Second,
		Timeout:        5 * time.Second,
		ConsumerBuffer: 10,
	}

	queue, err := redismq.NewQueue(cfg)
	if err != nil {
		t.Fatalf("failed to create Redis queue: %v", err)
	}
	defer queue.Close()

	publisher := queue.Publisher()
	subscriber := queue.Subscriber()

	testQueue := "test-queue-direct"

	// 1. Send to queue
	// 发送到队列
	testPayload := "queue-message-integration"
	err = publisher.Send(ctx, testQueue, []byte(testPayload))
	if err != nil {
		t.Fatalf("failed to send: %v", err)
	}

	t.Logf("sent to queue: %s", testQueue)

	// 2. Consume from queue
	// 从队列消费
	received := make(chan []byte, 1)

	go func() {
		err := subscriber.Consume(ctx, testQueue, func(ctx context.Context, msg *integrationcontract.Message) error {
			received <- msg.Body
			t.Logf("consumed message: queue=%s, body=%s", msg.Queue, string(msg.Body))
			return nil
		})
		if err != nil {
			t.Logf("consume ended: %v", err)
		}
	}()

	// 3. Wait for consumption
	// 等待消费
	select {
	case body := <-received:
		if string(body) != testPayload {
			t.Fatalf("message mismatch: expected=%s, got=%s", testPayload, string(body))
		}
		t.Log("verified: consumed correct message")
	case <-time.After(5 * time.Second):
		t.Fatal("consume timeout: did not receive message within 5s")
	}
}

// TestRedisNativeClient tests native client exposure for advanced usage.
//
// TestRedisNativeClient 测试原生客户端暴露供高级使用。
func TestRedisNativeClient(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Redis backend")
	}

	cfg := &integrationcontract.MessageQueueConfig{
		Type:      "redis",
		RedisAddr: getEnvOrDefault("GORP_TEST_REDIS_ADDR", "localhost:6379"),
	}

	queue, err := redismq.NewQueue(cfg)
	if err != nil {
		t.Fatalf("failed to create Redis queue: %v", err)
	}
	defer queue.Close()

	// Test native client exposure
	// 测试原生客户端暴露
	client := queue.NativeMQClient()
	if client == nil {
		t.Fatal("expected native client to be exposed")
	}

	t.Logf("native client exposed: %T", client)
}
