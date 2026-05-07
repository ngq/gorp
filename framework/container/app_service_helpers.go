// Application scenarios:
// - Provide strongly typed access helpers for app-service style bindings.
// - Reduce repetitive container type assertions in business and bootstrap code.
// - Keep generic service resolution ergonomics inside the container package.
//
// 适用场景：
// - 为 app-service 风格绑定提供强类型访问 helper。
// - 减少业务代码和 bootstrap 代码中重复的容器类型断言。
// - 把泛型服务解析的人体工学能力收口在 container 包内。
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
