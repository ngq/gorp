// Package nacos provides nacos service registry implementation for gorp.
// Import this package to register the nacos provider with bootstrap.
//
// nacos 注册中心 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/registry/nacos"
package nacos

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterDiscoveryProviderFactory("nacos", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}