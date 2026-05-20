// Package zookeeper provides zookeeper service registry implementation for gorp.
// Import this package to register the zookeeper provider with bootstrap.
//
// zookeeper 注册中心 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/registry/zookeeper"
package zookeeper

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterDiscoveryProviderFactory("zookeeper", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}