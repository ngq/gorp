// Package kafka provides kafka message queue implementation for gorp.
// Import this package to register the kafka provider with bootstrap.
//
// kafka 消息队列 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/messagequeue/kafka"
package kafka

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterMessageQueueProviderFactory("kafka", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}