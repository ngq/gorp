// Package etcd provides etcd service registry implementation for gorp.
// Import this package to register the etcd provider with bootstrap.
//
// etcd 注册中心 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/registry/etcd"
package etcd

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterDiscoveryProviderFactory("etcd", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}
