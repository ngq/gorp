// Package rocketmq provides rocketmq message queue implementation for gorp.
// Import this package to register the rocketmq provider with bootstrap.
//
// rocketmq 消息队列 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/messagequeue/rocketmq"
package rocketmq

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterMessageQueueProviderFactory("rocketmq", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}