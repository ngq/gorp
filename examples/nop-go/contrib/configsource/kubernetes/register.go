// Package kubernetes provides Kubernetes ConfigMap configuration source provider for the gorp framework.
// 本文件通过 init() 将 Kubernetes provider 注册到 bootstrap 工厂表，
// 使得业务方通过 import _ "github.com/ngq/gorp/contrib/configsource/kubernetes" 即可启用。
package kubernetes

import (
	"github.com/ngq/gorp/framework/bootstrap"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func init() {
	bootstrap.RegisterConfigSourceProviderFactory("kubernetes", func() runtimecontract.ServiceProvider {
		return NewProvider()
	})
}
