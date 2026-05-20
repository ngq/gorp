// Package redis provides Redis-based message queue implementation.
// This provider supports Pub/Sub and list-based queue operations.
//
// 本包提供基于 Redis 的消息队列实现。
// 支持 Pub/Sub 和列表队列操作。
package redis

import (
	"errors"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider implements runtimecontract.ServiceProvider for Redis message queue.
type Provider struct {
	BaseMQProvider
}

// NewProvider creates a new Redis message queue provider.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "messagequeue.redis"
	p.GetConfig = getMQConfig
	p.NewQueue = func(cfg *integrationcontract.MessageQueueConfig) (integrationcontract.MessageQueue, error) {
		return NewQueue(cfg)
	}
	return p
}

// getMQConfig extracts Redis configuration from the container.
func getMQConfig(c runtimecontract.Container) (*integrationcontract.MessageQueueConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("messagequeue: invalid config service")
	}

	mqCfg := &integrationcontract.MessageQueueConfig{
		Type:           "redis",
		MaxRetry:       3,
		RetryDelay:     time.Second,
		Timeout:        5 * time.Second,
		ConsumerBuffer: 100,
	}
	if addr := cfg.GetString("message_queue.redis.addr"); addr != "" {
		mqCfg.RedisAddr = addr
	} else {
		mqCfg.RedisAddr = "localhost:6379"
	}
	if password := cfg.GetString("message_queue.redis.password"); password != "" {
		mqCfg.RedisPassword = password
	}
	if db := cfg.GetInt("message_queue.redis.db"); db > 0 {
		mqCfg.RedisDB = db
	}
	if maxRetry := cfg.GetInt("message_queue.max_retry"); maxRetry > 0 {
		mqCfg.MaxRetry = maxRetry
	}
	return mqCfg, nil
}
