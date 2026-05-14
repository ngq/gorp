// Package rocketmq provides RocketMQ-based message queue provider for the gorp framework.
// This provider implements MessageQueue contract with apache/rocketmq-client-go SDK integration.
//
// 本包提供 gorp 框架基于 RocketMQ 的消息队列 provider。
// 本 provider 实现 MessageQueue 契约，集成 apache/rocketmq-client-go SDK。
package rocketmq

import (
	"errors"
	"time"

	"github.com/ngq/gorp/contrib/internal/basemq"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider implements runtimecontract.ServiceProvider for RocketMQ message queue.
type Provider struct {
	basemq.BaseMQProvider
}

// NewProvider creates a new RocketMQ message queue provider.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "messagequeue.rocketmq"
	p.GetConfig = getRocketMQConfig
	p.NewQueue = func(cfg *integrationcontract.MessageQueueConfig) (integrationcontract.MessageQueue, error) {
		return NewQueue(cfg)
	}
	return p
}

// getRocketMQConfig extracts RocketMQ configuration from the container.
func getRocketMQConfig(c runtimecontract.Container) (*integrationcontract.MessageQueueConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("messagequeue.rocketmq: invalid config service")
	}

	mqCfg := &integrationcontract.MessageQueueConfig{
		Type:                 "rocketmq",
		MaxRetry:             3,
		RetryDelay:           time.Second,
		Timeout:              5 * time.Second,
		ConsumerBuffer:       100,
		RocketMQNamesrvAddr:  "localhost:9876",
		RocketMQGroupName:    "default-group",
		RocketMQInstanceName: "",
		RocketMQRetryTimes:   3,
	}

	if namesrv := cfg.GetString("message_queue.rocketmq.namesrv_addr"); namesrv != "" {
		mqCfg.RocketMQNamesrvAddr = namesrv
	}
	if groupName := cfg.GetString("message_queue.rocketmq.group_name"); groupName != "" {
		mqCfg.RocketMQGroupName = groupName
	}
	if instanceName := cfg.GetString("message_queue.rocketmq.instance_name"); instanceName != "" {
		mqCfg.RocketMQInstanceName = instanceName
	}
	if retryTimes := cfg.GetInt("message_queue.rocketmq.retry_times"); retryTimes > 0 {
		mqCfg.RocketMQRetryTimes = retryTimes
	}
	if enableTLS := cfg.GetBool("message_queue.rocketmq.enable_tls"); enableTLS {
		mqCfg.RocketMQEnableTLS = enableTLS
	}

	if maxRetry := cfg.GetInt("message_queue.max_retry"); maxRetry > 0 {
		mqCfg.MaxRetry = maxRetry
	}
	if timeoutMs := cfg.GetInt("message_queue.timeout_ms"); timeoutMs > 0 {
		mqCfg.Timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	return mqCfg, nil
}
