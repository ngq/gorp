// Package otel provides otel tracing implementation for gorp.
// Import this package to register the otel provider with bootstrap.
//
// otel 链路追踪 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/tracing/otel"
package otel

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterTracingProviderFactory("otel", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}