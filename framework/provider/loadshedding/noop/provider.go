// Package noop 提供 LoadShedding 的空实现。
//
// 适用场景：
// - 单体/Gin-first 模式默认使用此 provider；
// - 所有请求都允许通过，不进行过载保护；
// - 作为 governance.disable=loadshedding 时的回退 provider。
package noop

import (
	"context"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider 提供 noop 过载保护实现。
//
// 中文说明：
// - 单体项目默认使用此 provider；
// - 所有请求都允许通过；
// - 不进行实际的过载保护。
type Provider struct{}

// NewProvider 创建 noop LoadShedding provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 provider 唯一名称。
func (p *Provider) Name() string { return "loadshedding.noop" }

// IsDefer 标记此 provider 延迟装载。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回该 provider 提供的容器 key 列表。
func (p *Provider) Provides() []string {
	return []string{resiliencecontract.LoadShedderKey}
}

// DependsOn returns the keys this provider depends on.
// Noop loadshedding has no dependencies.
//
// DependsOn 返回该 provider 依赖的 key。
// Noop loadshedding 无依赖。
func (p *Provider) DependsOn() []string { return nil }

// Requires 返回该 provider 依赖的容器 key 列表（无外部依赖）。
func (p *Provider) Requires() []string { return nil }

// Register 将 noop LoadShedder 实例注册到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(resiliencecontract.LoadShedderKey, func(c runtimecontract.Container) (any, error) {
		return &noopLoadShedder{}, nil
	}, true)
	return nil
}

// Boot 启动期初始化（无额外操作）。
func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

// noopLoadShedder 是 LoadShedder 的空实现。
// 所有请求都允许通过，不进行过载保护。
type noopLoadShedder struct{}

// Allow 总是允许请求通过（空操作）。
func (ls *noopLoadShedder) Allow(ctx context.Context, resource string) error {
	return nil
}

// Done 空操作（不释放任何资源）。
func (ls *noopLoadShedder) Done(ctx context.Context, resource string, err error) {}
