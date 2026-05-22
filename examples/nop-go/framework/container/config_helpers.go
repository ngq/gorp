// Package container provides runtime dependency injection container for gorp framework.
// This file exposes typed config access helpers on top of runtime container.
// Keeps config capability lookup consistent across bootstrap and application code.
//
// 容器包提供 gorp 框架的运行时依赖注入容器实现。
// 本文件在运行时容器之上暴露强类型配置访问 helper。
// 让 bootstrap 与 application 代码获取配置能力时保持一致。
package container

import (
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// MakeConfig resolves the config capability from the container.
//
// MakeConfig 从容器中解析配置能力。
func MakeConfig(c runtimecontract.Container) (datacontract.Config, error) {
	return MakeWith[datacontract.Config](c, datacontract.ConfigKey)
}
