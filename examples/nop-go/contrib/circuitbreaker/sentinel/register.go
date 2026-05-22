// Package sentinel provides sentinel circuit breaker implementation for gorp.
// Import this package to register the sentinel provider with bootstrap.
//
// sentinel 熔断降级 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/circuitbreaker/sentinel"
package sentinel

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterCircuitBreakerProviderFactory("sentinel", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}