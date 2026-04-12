package bootstrap

import (
	"github.com/ngq/gorp/framework/contract"
)

// RegisterAppServices 批量注册轻量应用服务绑定。
//
// 中文说明：
// - 用于不依赖模板的业务项目快速注册自己的 app service；
// - 每个服务仍通过显式 key 注册，避免隐藏装配规则；
// - 相比手工多次 Bind，这里提供一个更轻的统一入口。
func RegisterAppServices(c contract.Container, bindings map[string]contract.Factory) {
	for key, factory := range bindings {
		if factory == nil {
			continue
		}
		c.Bind(key, factory, true)
	}
}
