// Package etcd provides etcd configuration center provider for the gorp framework.
// 本文件通过 init() 将 etcd provider 注册到 bootstrap 工厂表，
// 使得业务方通过 import _ "github.com/ngq/gorp/contrib/configsource/etcd" 即可启用。
package etcd

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterConfigSourceProviderFactory("etcd", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}
