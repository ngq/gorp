package container

import (
	"fmt"

	"github.com/ngq/gorp/framework/contract"
)

// MakeAppService 从容器中按 key 解析轻量应用服务。
//
// 中文说明：
// - 这是 framework 级轻量 CRUD / app service 模式的最小接入 helper；
// - 不要求业务项目必须理解 provider/container 全部细节；
// - 业务项目可以把自己的应用服务按 key 绑定后，通过该 helper 获取。
func MakeAppService[T any](c contract.Container, key string) (T, error) {
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

// MustMakeAppService 从容器中强制解析轻量应用服务。
func MustMakeAppService[T any](c contract.Container, key string) T {
	v := c.MustMake(key)
	return v.(T)
}
