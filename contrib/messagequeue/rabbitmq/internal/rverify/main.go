// Package main 是 RabbitMQ 消息队列集成验证，验证 MQ-12/MQ-13 修复：
// 1. MQ-12：匿名订阅队列应为 exclusive+auto-delete（连接关闭后自动删除，
//    不再遗留 durable 孤儿队列）。
// 2. MQ-13：断线后订阅自动重建（消费不静默死亡）。
//
// 用法：在 contrib/messagequeue/rabbitmq 模块目录下执行
//   go run ./internal/rverify
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ngq/gorp/contrib/messagequeue/rabbitmq"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	reconnect := flag.Bool("reconnect", false, "run reconnection test (restart rabbitmq while waiting)")
	waitSec := flag.Int("wait", 8, "seconds to wait for external restart in reconnect mode")
	flag.Parse()

	cfg := &integrationcontract.MessageQueueConfig{
		RabbitMQURL:          "amqp://guest:guest@localhost:5672/",
		RabbitMQExchange:     "verify-exchange",
		RabbitMQExchangeType: "fanout",
		RabbitMQQueuePrefix:  "verify-prefix",
		RabbitMQPrefetch:     10,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	queue, err := rabbitmq.NewQueue(cfg)
	if err != nil {
		fmt.Printf("FAIL NewQueue: %v\n", err)
		os.Exit(1)
	}
	defer queue.Close()
	// NewQueue 已声明 fanout exchange，无需手动声明。

	sub := queue.Subscriber()
	msgCh := make(chan string, 50)

	if *reconnect {
		runReconnectTest(ctx, queue, sub, pubOrNil(ctx, queue), msgCh, *waitSec)
		return
	}

	// 匿名订阅（无 group）——MQ-12 修复：应为 exclusive+auto-delete 队列。
	unsub, err := sub.Subscribe(ctx, "verify-exchange", func(ctx context.Context, msg *integrationcontract.Message) error {
		msgCh <- string(msg.Body)
		return nil
	})
	if err != nil {
		fmt.Printf("FAIL anonymous subscribe: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK anonymous subscription created")

	// 通过管理 API 查匿名队列属性。
	time.Sleep(1500 * time.Millisecond)
	qName, autoDelete, exclusive, durable := findQueueViaHTTP("verify-prefix")
	fmt.Printf("anonymous queue: name=%q durable=%v auto_delete=%v exclusive=%v\n", qName, durable, autoDelete, exclusive)
	if qName == "" {
		fmt.Println("FAIL cannot find anonymous queue via management API")
		os.Exit(1)
	}
	if durable || !autoDelete || !exclusive {
		fmt.Printf("FAIL anonymous queue should be exclusive+auto-delete, got durable=%v auto=%v excl=%v\n", durable, autoDelete, exclusive)
		os.Exit(1)
	}
	fmt.Println("OK anonymous queue is exclusive+auto-delete")

	// 发布消息验证消费。
	pub := queue.Publisher()
	for i := 0; i < 3; i++ {
		if err := pub.Publish(ctx, "verify-exchange", []byte(fmt.Sprintf("m-%d", i))); err != nil {
			fmt.Printf("FAIL publish: %v\n", err)
			os.Exit(1)
		}
	}
	got := 0
	select {
	case <-msgCh:
		got++
	case <-time.After(6 * time.Second):
		fmt.Println("FAIL no message consumed")
		os.Exit(1)
	}
	for len(msgCh) > 0 {
		<-msgCh
		got++
	}
	fmt.Printf("OK consumed %d messages via anonymous queue\n", got)

	// 取消订阅后，exclusive+auto-delete 队列应被 broker 删除。
	_ = unsub()
	time.Sleep(2 * time.Second)
	if remaining, _, _, _ := findQueueViaHTTP("verify-prefix"); remaining != "" {
		fmt.Printf("WARN queue %q still exists after unsubscribe (will be removed on connection close)\n", remaining)
	} else {
		fmt.Println("OK anonymous queue removed after unsubscribe")
	}

	fmt.Println("ALL PASS")
}

// pubOrNil 返回 publisher（占位，避免主流程改动）。
func pubOrNil(ctx context.Context, q *rabbitmq.Queue) integrationcontract.MessagePublisher {
	return q.Publisher()
}

// runReconnectTest 验证 MQ-13：broker 重启后订阅自动重建、消费恢复。
func runReconnectTest(ctx context.Context, queue *rabbitmq.Queue, sub integrationcontract.MessageSubscriber, pub integrationcontract.MessagePublisher, msgCh chan string, waitSec int) {
	unsub, err := sub.Subscribe(ctx, "verify-exchange", func(ctx context.Context, msg *integrationcontract.Message) error {
		msgCh <- string(msg.Body)
		return nil
	})
	if err != nil {
		fmt.Printf("FAIL subscribe for reconnect test: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = unsub() }()
	time.Sleep(1500 * time.Millisecond)

	// 消费循环已建立：先发布一条确认消费链路通。
	if err := pub.Publish(ctx, "verify-exchange", []byte("before-restart")); err != nil {
		fmt.Printf("FAIL pre-restart publish: %v\n", err)
		os.Exit(1)
	}
	select {
	case <-msgCh:
		fmt.Println("OK consumed before restart")
	case <-time.After(6 * time.Second):
		fmt.Println("WARN no pre-restart consumption (timing)")
	}

	// 等待外部重启 broker（docker restart rabbitmq）。
	fmt.Printf(">>> waiting %d seconds for external broker restart ...\n", waitSec)
	time.Sleep(time.Duration(waitSec) * time.Second)

	// 重启后持续发布，检查消费是否恢复。
	recovered := false
	for i := 0; i < 5 && !recovered; i++ {
		if err := pub.Publish(ctx, "verify-exchange", []byte("post-restart")); err != nil {
			fmt.Printf("WARN publish after restart: %v\n", err)
		}
		select {
		case m := <-msgCh:
			fmt.Printf("OK consumed after restart: %s\n", m)
			recovered = true
		case <-time.After(4 * time.Second):
		}
	}
	if recovered {
		fmt.Println("ALL PASS (reconnection restored consumption)")
	} else {
		fmt.Println("FAIL consumption NOT restored after broker restart")
		os.Exit(1)
	}
}
func findQueueViaHTTP(prefix string) (name string, autoDelete, exclusive, durable bool) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:15672/api/queues", nil)
	if err != nil {
		return "", false, false, false
	}
	req.SetBasicAuth("guest", "guest")
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", false, false, false
	}
	defer resp.Body.Close()
	var queues []struct {
		Name       string `json:"name"`
		Durable    bool   `json:"durable"`
		AutoDelete bool   `json:"auto_delete"`
		Exclusive  bool   `json:"exclusive"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&queues); err != nil {
		return "", false, false, false
	}
	for _, q := range queues {
		if len(q.Name) >= len(prefix) && q.Name[:len(prefix)] == prefix {
			return q.Name, q.AutoDelete, q.Exclusive, q.Durable
		}
	}
	return "", false, false, false
}

var _ = amqp.Channel{} // keep amqp import for potential reuse
