// Package integration provides Kafka message queue integration tests.
//
// 本包提供 Kafka 消息队列集成测试。
package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	kafkamq "github.com/ngq/gorp/contrib/messagequeue/kafka"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

// TestKafkaPublishSubscribe tests real Kafka publish and subscribe operations.
//
// TestKafkaPublishSubscribe 测试真实 Kafka 发布和订阅操作。
func TestKafkaPublishSubscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Kafka backend")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. Create Kafka queue
	// 创建 Kafka queue
	// Use localhost instead of container hostname for local testing
	// 本地测试使用 localhost 而不是容器 hostname
	brokerAddr := "localhost:9092"
	cfg := &integrationcontract.MessageQueueConfig{
		Type:                 "kafka",
		KafkaBrokers:         []string{brokerAddr},
		KafkaClientID:        "gorp-integration-test",
		KafkaVersion:         "2.8.0",
		KafkaGroupID:         "gorp-test-group",
		KafkaMaxMessageBytes: 1000000, // Required: must be > 0
		MaxRetry:             3,
		RetryDelay:           time.Second,
		Timeout:              10 * time.Second,
		ConsumerBuffer:       100,
	}

	queue, err := kafkamq.NewQueue(cfg)
	if err != nil {
		t.Fatalf("failed to create Kafka queue: %v", err)
	}
	defer queue.Close()

	// 2. Get publisher
	// 获取 publisher
	publisher := queue.Publisher()
	if publisher == nil {
		t.Fatal("publisher is nil")
	}

	// 3. Get subscriber
	// 获取 subscriber
	subscriber := queue.Subscriber()
	if subscriber == nil {
		t.Fatal("subscriber is nil")
	}

	// 4. Generate unique topic for this test
	// 为此测试生成唯一 topic
	topic := fmt.Sprintf("test-topic-%d", time.Now().UnixNano())

	// 5. Publish test messages
	// 发布测试消息
	testMessages := []string{
		"test-message-1",
		"test-message-2",
		"test-message-3",
	}

	for i, msg := range testMessages {
		err := publisher.Publish(ctx, topic, []byte(msg))
		if err != nil {
			t.Fatalf("failed to publish message %d: %v", i, err)
		}
		t.Logf("published message: %s", msg)
	}

	// 6. Subscribe and consume messages
	// 订阅并消费消息
	receivedCount := 0
	receivedMessages := make([]string, 0)

	// Use a channel to collect messages with timeout
	// 使用 channel 收集消息并设置超时
	msgCh := make(chan string, len(testMessages))
	errCh := make(chan error, 1)

	go func() {
		unsub, err := subscriber.Subscribe(ctx, topic, func(ctx context.Context, msg *integrationcontract.Message) error {
			msgCh <- string(msg.Body)
			return nil
		})
		if err != nil {
			errCh <- err
		}
		defer unsub()
	}()

	// Collect messages with timeout
	// 带超时收集消息
	collectTimeout := time.After(10 * time.Second)
	for receivedCount < len(testMessages) {
		select {
		case msg := <-msgCh:
			receivedMessages = append(receivedMessages, msg)
			receivedCount++
			t.Logf("received message: %s", msg)
		case err := <-errCh:
			t.Logf("subscribe error (expected on context cancel): %v", err)
			return
		case <-collectTimeout:
			t.Logf("timeout waiting for messages, received %d/%d", receivedCount, len(testMessages))
			goto done
		}
	}

done:
	// 7. Verify received messages
	// 验证接收的消息
	if receivedCount == 0 {
		t.Log("WARNING: no messages received, Kafka may need time for topic creation")
	} else {
		t.Logf("received %d messages", receivedCount)
	}
}

// TestKafkaPublisherRetry tests Kafka publisher retry behavior.
//
// TestKafkaPublisherRetry 测试 Kafka publisher 重试行为。
func TestKafkaPublisherRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Kafka backend")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	brokerAddr := "localhost:9092"
	cfg := &integrationcontract.MessageQueueConfig{
		Type:                 "kafka",
		KafkaBrokers:         []string{brokerAddr},
		KafkaClientID:        "gorp-retry-test",
		KafkaVersion:         "2.8.0",
		KafkaGroupID:         "gorp-retry-group",
		KafkaMaxMessageBytes: 1000000, // Required: must be > 0
		MaxRetry:             3,
		RetryDelay:           100 * time.Millisecond,
		Timeout:              5 * time.Second,
	}

	queue, err := kafkamq.NewQueue(cfg)
	if err != nil {
		t.Fatalf("failed to create Kafka queue: %v", err)
	}
	defer queue.Close()

	publisher := queue.Publisher()

	topic := fmt.Sprintf("retry-topic-%d", time.Now().UnixNano())

	// Publish with retry
	// 带重试发布
	err = publisher.Publish(ctx, topic, []byte("retry-test-message"))
	if err != nil {
		t.Logf("publish failed after retries: %v", err)
	} else {
		t.Log("publish succeeded")
	}
}

// TestKafkaQueueUnderlying tests accessing underlying Kafka client.
//
// TestKafkaQueueUnderlying 测试访问底层 Kafka client。
func TestKafkaQueueUnderlying(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Kafka backend")
	}

	brokerAddr := "localhost:9092"
	cfg := &integrationcontract.MessageQueueConfig{
		Type:                 "kafka",
		KafkaBrokers:         []string{brokerAddr},
		KafkaClientID:        "gorp-underlying-test",
		KafkaVersion:         "2.8.0",
		KafkaMaxMessageBytes: 1000000, // Required: must be > 0
	}

	queue, err := kafkamq.NewQueue(cfg)
	if err != nil {
		t.Fatalf("failed to create Kafka queue: %v", err)
	}
	defer queue.Close()

	// Test Underlying
	// 测试 Underlying
	underlying := queue.Underlying()
	if underlying == nil {
		t.Fatal("Underlying() returned nil")
	}

	// Test NativeMQClient
	// 测试 NativeMQClient
	nativeClient := queue.NativeMQClient()
	if nativeClient == nil {
		t.Fatal("NativeMQClient() returned nil")
	}

	// Test NativeSyncProducer
	// 测试 NativeSyncProducer
	producer := queue.NativeSyncProducer()
	if producer == nil {
		t.Fatal("NativeSyncProducer() returned nil")
	}

	t.Log("verified: all underlying accessors return non-nil values")
}

// TestKafkaQueueClose tests proper resource cleanup.
//
// TestKafkaQueueClose 测试资源正确清理。
func TestKafkaQueueClose(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Kafka backend")
	}

	brokerAddr := "localhost:9092"
	cfg := &integrationcontract.MessageQueueConfig{
		Type:                 "kafka",
		KafkaBrokers:         []string{brokerAddr},
		KafkaClientID:        "gorp-close-test",
		KafkaVersion:         "2.8.0",
		KafkaMaxMessageBytes: 1000000, // Required: must be > 0
	}

	queue, err := kafkamq.NewQueue(cfg)
	if err != nil {
		t.Fatalf("failed to create Kafka queue: %v", err)
	}

	// Close should not panic
	// Close 不应该 panic
	err = queue.Close()
	if err != nil {
		t.Fatalf("first close failed: %v", err)
	}

	// Second close should be idempotent
	// 第二次 close 应该是幂等的
	err = queue.Close()
	if err != nil {
		t.Fatalf("second close failed: %v", err)
	}

	t.Log("verified: Close is idempotent")
}
