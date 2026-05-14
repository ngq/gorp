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

// MakeWith resolves a service by key and casts it to the requested type T.
// Returns an error if the service is not bound, resolution fails, or the type does not match.
// Prefer this over raw Make + type assertion for type safety.
//
// MakeWith 按 key 解析服务并转换为类型 T。
// 如果服务未绑定、解析失败或类型不匹配，返回错误。
// 优先使用此函数而非裸 Make + 类型断言，以获得类型安全。
func MakeWith[T any](c runtimecontract.Container, key string) (T, error) {
	var zero T
	v, err := c.Make(key)
	if err != nil {
		return zero, err
	}
	typed, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("service type mismatch: key=%s, expected=%T, got=%T", key, zero, v)
	}
	return typed, nil
}

// MustMakeWith resolves a service by key and casts it to type T, panicking on failure.
// Prefer this over raw MustMake + type assertion for type safety.
//
// MustMakeWith 按 key 解析服务并转换为类型 T，失败时 panic。
// 优先使用此函数而非裸 MustMake + 类型断言，以获得类型安全。
func MustMakeWith[T any](c runtimecontract.Container, key string) T {
	v, err := MakeWith[T](c, key)
	if err != nil {
		panic(err)
	}
	return v
}

// MakeNamedWith resolves a named service and casts it to type T.
// Returns an error if the service is not bound, resolution fails, or the type does not match.
//
// MakeNamedWith 解析命名服务并转换为类型 T。
// 如果服务未绑定、解析失败或类型不匹配，返回错误。
func MakeNamedWith[T any](c runtimecontract.Container, name, key string) (T, error) {
	var zero T
	v, err := c.MakeNamed(name, key)
	if err != nil {
		return zero, err
	}
	typed, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("named service type mismatch: name=%s, key=%s, expected=%T, got=%T", name, key, zero, v)
	}
	return typed, nil
}

// MustMakeNamedWith resolves a named service and casts it to type T, panicking on failure.
//
// MustMakeNamedWith 解析命名服务并转换为类型 T，失败时 panic。
func MustMakeNamedWith[T any](c runtimecontract.Container, name, key string) T {
	v, err := MakeNamedWith[T](c, name, key)
	if err != nil {
		panic(err)
	}
	return v
}

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
