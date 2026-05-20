// Package consul provides Consul configuration center provider for the gorp framework.
// 本文件通过 init() 将 Consul provider 注册到 bootstrap 工厂表，
// 使得业务方通过 import _ "github.com/ngq/gorp/contrib/configsource/consul" 即可启用。
package consul

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterConfigSourceProviderFactory("consul", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}
