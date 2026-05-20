// Package rabbitmq provides rabbitmq message queue implementation for gorp.
// Import this package to register the rabbitmq provider with bootstrap.
//
// rabbitmq 消息队列 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/messagequeue/rabbitmq"
package rabbitmq

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterMessageQueueProviderFactory("rabbitmq", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}