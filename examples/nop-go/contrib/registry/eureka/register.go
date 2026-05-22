// Package eureka provides eureka service registry implementation for gorp.
// Import this package to register the eureka provider with bootstrap.
//
// eureka 注册中心 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/registry/eureka"
package eureka

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterDiscoveryProviderFactory("eureka", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}