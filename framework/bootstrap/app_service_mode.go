// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file registers app-service-specific runtime bindings during bootstrap.
// Keeps app-service mode assembly compact while preserving singleton startup behavior.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件在 bootstrap 阶段注册 app-service 模式专用的运行时绑定。
// 在保持单例启动语义的同时，让 app-service 模式装配保持精简。
package bootstrap

import runtimecontract "github.com/ngq/gorp/framework/contract/runtime"

// RegisterAppServices registers app-service runtime bindings as singletons.
//
// RegisterAppServices 以单例形式注册 app-service 运行时绑定。
func RegisterAppServices(c runtimecontract.Container, bindings map[string]runtimecontract.Factory) {
	for key, factory := range bindings {
		if factory == nil {
			// Skip nil factories so partially assembled binding maps stay safe to consume.
			// 跳过空 factory，保证部分装配完成的绑定表也能安全消费。
			continue
		}
		c.Bind(key, factory, true)
	}
}
