// Package container_test provides unit tests for generic safe API functions.
//
// 适用场景：
// - 验证 MakeWith[T] / MustMakeWith[T] 的类型安全解析。
// - 验证 MakeNamedWith[T] / MustMakeNamedWith[T] 的类型安全解析。
// - 验证类型不匹配时返回错误而非 panic。
package container

import (
	"context"
	"io"
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
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

// TestMakeWith_InterfaceTypeAssertion verifies that interface type assertions work correctly.
// This is a diagnostic test for the generic type assertion behavior.
//
// TestMakeWith_InterfaceTypeAssertion 验证接口类型断言行为正确。
func TestMakeWith_InterfaceTypeAssertion(t *testing.T) {
	// Test 1: concrete type to interface assertion (should work)
	var v any = "hello"
	_, ok := v.(string)
	require.True(t, ok, "string assertion should work")

	// Test 2: pointer type to interface assertion
	var iface any = &closeFunc{}
	_, ok2 := iface.(io.Closer)
	require.True(t, ok2, "*closeFunc should implement io.Closer")

	// Test 3: generic MakeWith with interface type
	c := New()
	c.Bind("closer", func(runtimecontract.Container) (any, error) {
		return &closeFunc{}, nil
	}, true)
	_, err := MakeWith[io.Closer](c, "closer")
	require.NoError(t, err, "MakeWith[io.Closer] should work for *closeFunc")
}

// TestMakeWith_ContractInterface verifies that MakeWith[T] works with contract interface types.
//
// TestMakeWith_ContractInterface 验证 MakeWith[T] 可用于契约接口类型。
func TestMakeWith_ContractInterface(t *testing.T) {
	c := New()
	c.Bind("test.key", func(runtimecontract.Container) (any, error) {
		return &testCloser{}, nil
	}, true)

	v, err := MakeWith[io.Closer](c, "test.key")
	require.NoError(t, err)
	require.NotNil(t, v)
}

type testCloser struct{}

func (t *testCloser) Close() error { return nil }

// TestMakeWith_DataContractInterface verifies that MakeWith[T] works with datacontract interfaces.
//
// TestMakeWith_DataContractInterface 验证 MakeWith[T] 可用于 datacontract 接口类型。
func TestMakeWith_DataContractInterface(t *testing.T) {
	// First verify the mock implements the interface
	var _ datacontract.Redis = (*testRedis)(nil)

	c := New()
	c.Bind(datacontract.RedisKey, func(runtimecontract.Container) (any, error) {
		return &testRedis{}, nil
	}, true)

	// Test direct type assertion
	v, err := c.Make(datacontract.RedisKey)
	require.NoError(t, err)

	// Manual type assertion should work
	redis, ok := v.(datacontract.Redis)
	require.True(t, ok, "manual type assertion should work")
	require.NotNil(t, redis)

	// Now test MakeWith
	redis2, err := MakeWith[datacontract.Redis](c, datacontract.RedisKey)
	require.NoError(t, err, "MakeWith[datacontract.Redis] should work")
	require.NotNil(t, redis2)
}

// TestMakeWith_SecurityContractInterface verifies that MakeWith[T] works with securitycontract interfaces.
//
// TestMakeWith_SecurityContractInterface 验证 MakeWith[T] 可用于 securitycontract 接口类型。
func TestMakeWith_SecurityContractInterface(t *testing.T) {
	// Verify mock implements the interface
	var _ securitycontract.JWTService = (*testJWTService)(nil)

	c := New()
	c.Bind(securitycontract.AuthJWTKey, func(runtimecontract.Container) (any, error) {
		return &testJWTService{}, nil
	}, true)

	// Test direct type assertion
	v, err := c.Make(securitycontract.AuthJWTKey)
	require.NoError(t, err)

	// Manual type assertion should work
	jwt, ok := v.(securitycontract.JWTService)
	require.True(t, ok, "manual type assertion should work for JWTService")
	require.NotNil(t, jwt)

	// Now test MakeWith
	jwt2, err := MakeWith[securitycontract.JWTService](c, securitycontract.AuthJWTKey)
	require.NoError(t, err, "MakeWith[securitycontract.JWTService] should work")
	require.NotNil(t, jwt2)
}

type testRedis struct{}

func (t *testRedis) Ping(ctx context.Context) error                                   { return nil }
func (t *testRedis) Get(ctx context.Context, key string) (string, error)              { return "value", nil }
func (t *testRedis) Set(ctx context.Context, key, value string, ttlSeconds int) error { return nil }
func (t *testRedis) Del(ctx context.Context, key string) error                        { return nil }
func (t *testRedis) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	return map[string]string{"k": "v"}, nil
}

type testJWTService struct{}

func (t *testJWTService) Sign(claims securitycontract.JWTClaims) (string, error) {
	return "token", nil
}
func (t *testJWTService) Verify(token string) (*securitycontract.JWTClaims, error) {
	return &securitycontract.JWTClaims{}, nil
}
func (t *testJWTService) NewClaims(subjectID int64, subjectType, subjectName string, roles []string, ttlSeconds int64) securitycontract.JWTClaims {
	return securitycontract.JWTClaims{}
}
