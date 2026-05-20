// Package polaris provides polaris service registry implementation for gorp.
// Import this package to register the polaris provider with bootstrap.
//
// polaris 注册中心 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/registry/polaris"
package polaris

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterDiscoveryProviderFactory("polaris", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}