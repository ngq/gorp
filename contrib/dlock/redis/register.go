// Package redis provides redis distributed lock implementation for gorp.
// Import this package to register the redis provider with bootstrap.
//
// redis 分布式锁 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/dlock/redis"
package redis

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterDistributedLockProviderFactory("redis", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}