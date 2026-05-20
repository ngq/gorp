// Package mtls provides mtls service auth implementation for gorp.
// Import this package to register the mtls provider with bootstrap.
//
// mtls 服务认证 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/serviceauth/mtls"
package mtls

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterServiceAuthProviderFactory("mtls", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}