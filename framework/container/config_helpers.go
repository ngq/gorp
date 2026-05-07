// Application scenarios:
// - Expose typed config access helpers on top of the runtime container.
// - Keep config capability lookup consistent across bootstrap and application code.
// - Avoid repeating container-to-config assertions in callers.
//
// 适用场景：
// - 在运行时容器之上暴露强类型配置访问 helper。
// - 让 bootstrap 与 application 代码获取配置能力时保持一致。
// - 避免调用方重复编写 container 到 config 的断言逻辑。
package container

import (
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// MakeConfig resolves the config capability from the container.
//
// MakeConfig 从容器中解析配置能力。
func MakeConfig(c runtimecontract.Container) (datacontract.Config, error) {
	v, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	return v.(datacontract.Config), nil
}
