// Package dtmsdk provides dtm distributed transaction implementation for gorp.
// Import this package to register the dtm provider with bootstrap.
//
// dtm 分布式事务 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/dtm/dtmsdk"
package dtmsdk

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterDTMProviderFactory("dtmsdk", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}