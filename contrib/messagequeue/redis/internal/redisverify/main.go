// Package main 是 Redis 消息队列集成验证，验证 MQ-14 修复：
// 毒丸消息（永远处理失败）不应零间隔热循环，而应在重试上限后进入死信队列。
//
// 用法：在 contrib/messagequeue/redis 模块目录下执行
//   go run ./internal/redisverify
package main

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/ngq/gorp/contrib/messagequeue/redis"
	redis9 "github.com/redis/go-redis/v9"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
)

func main() {
	cfg := &integrationcontract.MessageQueueConfig{
		RedisAddr: "localhost:6380",
		Timeout:    5 * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	queue, err := redis.NewQueue(cfg)
	if err != nil {
		fmt.Printf("FAIL NewQueue: %v\n", err)
		os.Exit(1)
	}
	defer queue.Close()

	queueName := fmt.Sprintf("verify-queue-%d", time.Now().UnixNano())
	deadLetter := queueName + ":dead"
	client := queue.Underlying().(*redis9.Client)

	// 先清空可能的残留
	_ = client.Del(ctx, queueName, deadLetter)

	// Consume 走 Redis list（BLPop），而 Publish 走 Pub/Sub——这里直接用
	// RPush 投递一条毒丸消息到 list，验证 Consume 路径的毒丸→死信。
	if err := client.RPush(ctx, queueName, "poison-message").Err(); err != nil {
		fmt.Printf("FAIL push poison message: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK pushed poison message to list queue")

	// 消费：handler 永远失败。
	var attempts int64
	sub := queue.Subscriber()
	consumeDone := make(chan error, 1)
	go func() {
		consumeDone <- sub.Consume(ctx, queueName, func(ctx context.Context, msg *integrationcontract.Message) error {
			atomic.AddInt64(&attempts, 1)
			return fmt.Errorf("always fail")
		})
	}()

	// 等待毒丸消息经过重试退避（5 次上限，退避 1+2+4+8=15s 内）+ 进入死信。
	time.Sleep(35 * time.Second)

	// 检查死信队列是否收到消息
	dlLen, err := queue.Underlying().(*redis9.Client).LLen(ctx, deadLetter).Result()
	if err != nil {
		fmt.Printf("FAIL check dead letter len: %v\n", err)
		os.Exit(1)
	}
	attemptsVal := atomic.LoadInt64(&attempts)
	fmt.Printf("handler attempts: %d, dead letter len: %d\n", attemptsVal, dlLen)

	if dlLen < 1 {
		fmt.Println("FAIL poison message did NOT reach dead letter queue")
		os.Exit(1)
	}
	if attemptsVal > 60 {
		// 若没有退避，20s 内会热循环几百上千次；退避后约 5 次。
		fmt.Printf("FAIL attempts %d too high — likely no backoff (hot loop)\n", attemptsVal)
		os.Exit(1)
	}
	fmt.Printf("OK poison message moved to dead letter after %d attempts (backoff worked)\n", attemptsVal)

	// 原队列应为空（毒丸不再回推）
	qLen, _ := queue.Underlying().(*redis9.Client).LLen(ctx, queueName).Result()
	if qLen != 0 {
		fmt.Printf("WARN original queue len %d (may still be draining)\n", qLen)
	} else {
		fmt.Println("OK original queue empty after poison moved to dead letter")
	}

	// 清理
	_ = queue.Underlying().(*redis9.Client).Del(ctx, queueName, deadLetter)
	fmt.Println("ALL PASS")
}
