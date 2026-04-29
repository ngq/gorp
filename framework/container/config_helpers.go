package container

import "github.com/ngq/gorp/framework/contract"

// MakeConfig 从容器获取 Config 服务，失败返回 error。
//
// 中文说明：
// - 与 MustMakeConfig 相比，失败时返回 error 而不 panic；
// - 适用于需要优雅处理错误的场景；
// - 这是业务与框架代码读取统一配置服务的标准 helper 入口。
func MakeConfig(c contract.Container) (contract.Config, error) {
	v, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	return v.(contract.Config), nil
}
