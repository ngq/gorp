package container

import "sort"

// ProviderNames 返回当前容器中已注册的 provider 名称列表（按字典序排序）。
//
// 中文说明：
// - 这是一个“仅供调试/脚手架命令使用”的辅助方法；
// - 它刻意没有放进公共 `contract.Container` 接口中，避免把“列出所有 provider”
//   这种偏管理侧的能力暴露成框架运行时的通用契约；
// - 当前主要被 `gorp provider list` 这类命令使用。
func (c *Container) ProviderNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]string, 0, len(c.providersByName))
	for name := range c.providersByName {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
