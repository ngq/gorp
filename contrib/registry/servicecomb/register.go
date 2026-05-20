// Package servicecomb provides servicecomb service registry implementation for gorp.
// Import this package to register the servicecomb provider with bootstrap.
//
// servicecomb 注册中心 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/registry/servicecomb"
package servicecomb

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterDiscoveryProviderFactory("servicecomb", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}