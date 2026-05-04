package container

import (
	"fmt"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

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

func MustMakeAppService[T any](c runtimecontract.Container, key string) T {
	v := c.MustMake(key)
	return v.(T)
}
