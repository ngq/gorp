// Package token provides token service auth implementation for gorp.
// Import this package to register the token provider with bootstrap.
//
// token 服务认证 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/serviceauth/token"
package token

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterServiceAuthProviderFactory("token", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}