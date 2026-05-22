// Package container provides runtime dependency injection container for gorp framework.
// This file offers provider introspection and diagnostics helper methods.
// Supports CLI or tooling queries without polluting core container contract.
//
// 容器包提供 gorp 框架的运行时依赖注入容器实现。
// 本文件提供 provider 维度的辅助查询和诊断能力。
// 支持 CLI 或工具侧查询，而不污染核心运行时容器契约。
package container

import "sort"

// ProviderNames returns the registered provider names in lexicographical order.
//
// ProviderNames 按字典序返回当前已注册的 provider 名称列表。
//
// 中文说明：
// - 这是一个仅供调试与框架命令使用的辅助方法。
// - 它没有被放入公共 `contract.Container` 接口，避免把管理侧能力暴露成运行时通用契约。
// - 当前主要服务于类似 `gorp provider list` 这类工具场景。
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
