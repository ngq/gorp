// Package polaris provides polaris config source implementation for gorp.
// Import this package to register the polaris provider with bootstrap.
//
// polaris 配置中心 Provider，通过 init() 自动注册到 bootstrap。
//
// Example:
//
//	import _ "github.com/ngq/gorp/contrib/configsource/polaris"
package polaris

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterConfigSourceProviderFactory("polaris", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}