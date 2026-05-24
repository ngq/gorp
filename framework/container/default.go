// Package container provides runtime dependency injection container for gorp framework.
// This file manages the process-wide default container used by facade packages.
// Keeps default container reads and writes concurrency-safe with noop fallback.
//
// 容器包提供 gorp 框架的运行时依赖注入容器实现。
// 本文件管理 facade 包使用的进程级默认容器。
// 保证默认容器的读写具备并发安全，提供 noop 回退。
package container

import (
	"context"
	"errors"
	"io"
	"sync"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
)

var (
	defaultContainer   runtimecontract.Container
	defaultContainerMu sync.RWMutex
)

// ErrDefaultContainerNotSet 全局默认 Container 未设置时 Make 返回此错误。
// 业务代码在 gorp.Run 启动前调用 facade 会触发此错误。
var ErrDefaultContainerNotSet = errors.New("container: default container has not been set, please call gorp.Run first")

// SetDefault 设置全局默认 Container。
// 仅在 bootstrap 阶段调用一次，启动前所有 facade 调用将返回 ErrDefaultContainerNotSet。
//
// SetDefault replaces the process-wide default container.
func SetDefault(c runtimecontract.Container) {
	defaultContainerMu.Lock()
	defer defaultContainerMu.Unlock()
	defaultContainer = c
}

// Default 返回全局默认 Container。
// 如果尚未设置（gorp.Run 未调用），返回 noopContainer，其 Make 返回 ErrDefaultContainerNotSet。
//
// Default returns the process-wide default container.
// If not set, returns a noopContainer whose Make returns ErrDefaultContainerNotSet.
func Default() runtimecontract.Container {
	defaultContainerMu.RLock()
	defer defaultContainerMu.RUnlock()
	if defaultContainer != nil {
		return defaultContainer
	}
	return noopContainer{}
}

// Resolve 从 ctx 中提取 Container，提取不到则使用全局默认 Container。
// 优先级：ctx 中的 Container > 全局默认 Container > noopContainer。
// 传入 context.Background() 也能工作（fallback 到全局默认）。
//
// Resolve extracts Container from ctx; falls back to the global default.
// Priority: ctx Container > global default > noopContainer.
// context.Background() works by falling back to the global default.
func Resolve(ctx context.Context) runtimecontract.Container {
	if ctx != nil {
		if c, ok := supportcontract.FromContainerContext(ctx); ok {
			if cont, ok := c.(runtimecontract.Container); ok {
				return cont
			}
		}
	}
	return Default()
}

// noopContainer 在全局默认 Container 未设置时作为安全回退。
// Make/MakeNamed 返回 ErrDefaultContainerNotSet，其余方法为空操作。
//
// noopContainer is a safe fallback when no default container has been set.
// Make/MakeNamed return ErrDefaultContainerNotSet; all other methods are no-ops.
type noopContainer struct{}

var _ runtimecontract.Container = noopContainer{}

func (noopContainer) Bind(string, runtimecontract.Factory, bool)              {}
func (noopContainer) NamedBind(string, string, runtimecontract.Factory, bool) {}
func (noopContainer) IsBind(string) bool                                      { return false }
func (noopContainer) IsBindNamed(string, string) bool                         { return false }

func (noopContainer) Make(string) (any, error)       { return nil, ErrDefaultContainerNotSet }
func (noopContainer) MakeNamed(string, string) (any, error) { return nil, ErrDefaultContainerNotSet }

func (noopContainer) MustMake(key string) any {
	panic(ErrDefaultContainerNotSet)
}

func (noopContainer) MustMakeNamed(name, key string) any {
	panic(ErrDefaultContainerNotSet)
}

func (noopContainer) RegisterCloser(string, io.Closer)                    {}
func (noopContainer) Destroy() error                                      { return nil }
func (noopContainer) RegisterProvider(runtimecontract.ServiceProvider) error { return nil }
func (noopContainer) RegisterProviders(...runtimecontract.ServiceProvider) error { return nil }
func (noopContainer) RegisteredProviders() []runtimecontract.ProviderInfo { return nil }
func (noopContainer) DebugPrint() string                                 { return "" }
func (noopContainer) ProviderDAG() runtimecontract.ProviderDAG {
	return runtimecontract.ProviderDAG{}
}