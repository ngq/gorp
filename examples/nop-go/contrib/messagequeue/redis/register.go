// Package redis provides redis message queue implementation for gorp.
// Import this package to register the redis provider with bootstrap.
//
// redis 消息队列 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/messagequeue/redis"
package redis

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterMessageQueueProviderFactory("redis", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}