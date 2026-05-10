// Package container provides runtime dependency injection container for gorp framework.
// This file provides strongly typed access helpers for app-service style bindings.
// Reduces repetitive container type assertions in business and bootstrap code.
//
// 容器包提供 gorp 框架的运行时依赖注入容器实现。
// 本文件为 app-service 风格绑定提供强类型访问 helper。
// 减少业务代码和 bootstrap 代码中重复的容器类型断言。
package container

import (
	"fmt"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// MakeAppService resolves an app service by key and casts it to the requested type.
//
// MakeAppService 按 key 解析 app service，并转换成目标类型。
func MakeAppService[T any](c runtimecontract.Container, key string) (T, error) {
	var zero T
	v, err := c.Make(key)
	if err != nil {
		return zero, err
	}
	svc, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("app service type mismatch: key=%s, got=%T", key, v)
	}
	return svc, nil
}

// MustMakeAppService resolves an app service by key and panics on failure.
//
// MustMakeAppService 按 key 解析 app service，失败时 panic。
func MustMakeAppService[T any](c runtimecontract.Container, key string) T {
	v := c.MustMake(key)
	return v.(T)
}
