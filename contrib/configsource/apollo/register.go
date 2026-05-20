// Package apollo provides Apollo configuration center provider for the gorp framework.
// 本文件通过 init() 将 Apollo provider 注册到 bootstrap 工厂表，
// 使得业务方通过 import _ "github.com/ngq/gorp/contrib/configsource/apollo" 即可启用。
package apollo

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterConfigSourceProviderFactory("apollo", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}
