// Package container_test provides unit tests for generic safe API functions.
//
// 适用场景：
// - 验证 MakeWith[T] / MustMakeWith[T] 的类型安全解析。
// - 验证 MakeNamedWith[T] / MustMakeNamedWith[T] 的类型安全解析。
// - 验证类型不匹配时返回错误而非 panic。
package container

import (
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

// TestMakeWith_Success verifies that MakeWith[T] correctly resolves and types a service.
//
// TestMakeWith_Success 验证 MakeWith[T] 正确解析并类型化服务。
func TestMakeWith_Success(t *testing.T) {
	c := New()
	c.Bind("greeting", func(runtimecontract.Container) (any, error) {
		return "hello", nil
	}, true)

	v, err := MakeWith[string](c, "greeting")
	require.NoError(t, err)
	require.Equal(t, "hello", v)
}

// TestMakeWith_TypeMismatch verifies that MakeWith[T] returns an error
// when the resolved value does not match the requested type.
//
// TestMakeWith_TypeMismatch 验证 MakeWith[T] 在类型不匹配时返回错误。
func TestMakeWith_TypeMismatch(t *testing.T) {
	c := New()
	c.Bind("greeting", func(runtimecontract.Container) (any, error) {
		return "hello", nil
	}, true)

	_, err := MakeWith[int](c, "greeting")
	require.Error(t, err)
	require.Contains(t, err.Error(), "type mismatch")
}

// TestMakeWith_UnboundKey verifies that MakeWith[T] returns an error for unbound keys.
//
// TestMakeWith_UnboundKey 验证 MakeWith[T] 对未绑定 key 返回错误。
func TestMakeWith_UnboundKey(t *testing.T) {
	c := New()
	_, err := MakeWith[string](c, "missing")
	require.Error(t, err)
}

// TestMustMakeWith_Success verifies that MustMakeWith[T] resolves and types correctly.
//
// TestMustMakeWith_Success 验证 MustMakeWith[T] 正确解析并类型化。
func TestMustMakeWith_Success(t *testing.T) {
	c := New()
	c.Bind("greeting", func(runtimecontract.Container) (any, error) {
		return "hello", nil
	}, true)

	v := MustMakeWith[string](c, "greeting")
	require.Equal(t, "hello", v)
}

// TestMustMakeWith_TypeMismatchPanics verifies that MustMakeWith[T] panics on type mismatch.
//
// TestMustMakeWith_TypeMismatchPanics 验证 MustMakeWith[T] 在类型不匹配时 panic。
func TestMustMakeWith_TypeMismatchPanics(t *testing.T) {
	c := New()
	c.Bind("greeting", func(runtimecontract.Container) (any, error) {
		return "hello", nil
	}, true)

	require.Panics(t, func() {
		MustMakeWith[int](c, "greeting")
	})
}

// TestMakeNamedWith_Success verifies that MakeNamedWith[T] resolves named bindings with type safety.
//
// TestMakeNamedWith_Success 验证 MakeNamedWith[T] 类型安全地解析命名绑定。
func TestMakeNamedWith_Success(t *testing.T) {
	c := New()
	c.NamedBind("primary", "cache", func(runtimecontract.Container) (any, error) {
		return "redis", nil
	}, true)

	v, err := MakeNamedWith[string](c, "primary", "cache")
	require.NoError(t, err)
	require.Equal(t, "redis", v)
}

// TestMakeNamedWith_TypeMismatch verifies that MakeNamedWith[T] returns an error on type mismatch.
//
// TestMakeNamedWith_TypeMismatch 验证 MakeNamedWith[T] 在类型不匹配时返回错误。
func TestMakeNamedWith_TypeMismatch(t *testing.T) {
	c := New()
	c.NamedBind("primary", "cache", func(runtimecontract.Container) (any, error) {
		return "redis", nil
	}, true)

	_, err := MakeNamedWith[int](c, "primary", "cache")
	require.Error(t, err)
	require.Contains(t, err.Error(), "type mismatch")
}

// TestMustMakeNamedWith_Success verifies that MustMakeNamedWith[T] resolves named bindings.
//
// TestMustMakeNamedWith_Success 验证 MustMakeNamedWith[T] 解析命名绑定。
func TestMustMakeNamedWith_Success(t *testing.T) {
	c := New()
	c.NamedBind("primary", "cache", func(runtimecontract.Container) (any, error) {
		return "redis", nil
	}, true)

	v := MustMakeNamedWith[string](c, "primary", "cache")
	require.Equal(t, "redis", v)
}

// TestMakeWith_InterfaceType verifies that MakeWith[T] works with interface types.
//
// TestMakeWith_InterfaceType 验证 MakeWith[T] 可用于接口类型。
func TestMakeWith_InterfaceType(t *testing.T) {
	c := New()
	c.Bind("closer", func(runtimecontract.Container) (any, error) {
		return &closeFunc{}, nil
	}, true)

	v, err := MakeWith[*closeFunc](c, "closer")
	require.NoError(t, err)
	require.NotNil(t, v)
}
