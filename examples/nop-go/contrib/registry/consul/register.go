// Package consul 提供 Consul 服务注册中心的自注册入口。
//
// 导入此包即可将 consul provider 自动注册到 bootstrap。
//
// 使用方式：
//
//	import _ "github.com/ngq/gorp/contrib/registry/consul"
package consul

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterDiscoveryProviderFactory("consul", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}
