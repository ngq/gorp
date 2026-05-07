// Application scenarios:
// - Expose the main startup entrypoints of the application package.
// - Support both direct boot and runtime-only build workflows for HTTP services.
// - Keep service-name validation, startup cancellation, and runtime construction semantics stable.
//
// 适用场景：
// - 暴露 application 包的主启动入口。
// - 同时支持直接启动和仅构建 runtime 的 HTTP 服务工作流。
// - 稳定维护服务名校验、启动取消和 runtime 构建语义。
package application

import (
	"context"
	"errors"
	"strings"
)

// Run boots the default HTTP mainline with application options.
//
// 说明：业务运行入口仍是项目自己的 main；application 只提供启动装配 helper。
// Run 启动默认 HTTP 主线。
//
// Example:
//
//	err := application.Run(
//	    "user-service",
//	    application.HTTP(),
//	    application.WithProviders(myProvider),
//	    application.WithHTTPRoutes(func(router transportcontract.HTTPRouter, c runtimecontract.Container) error {
//	        registerRoutes(router)
//	        return nil
//	    }),
//	)
func Run(serviceName string, options ...Option) error {
	return RunContext(context.Background(), serviceName, options...)
}

// Start is an alias of Run.
//
// Start 是 Run 的同义入口。
func Start(serviceName string, options ...Option) error {
	return Run(serviceName, options...)
}

// RunContext boots the default mainline with an explicit context.
//
// 当前语义：仅在启动前检查取消，运行中关闭流程仍由 bootstrap 处理。
// RunContext 使用显式 context 启动默认主线。
func RunContext(ctx context.Context, serviceName string, options ...Option) error {
	if err := ensureStartupContext(ctx); err != nil {
		return errors.Join(ErrStartupCanceled, err)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	normalizedServiceName := strings.TrimSpace(serviceName)
	cfg, err := resolveRunConfig(normalizedServiceName, options...)
	if err != nil {
		return err
	}
	if err := bootHTTPService(normalizedServiceName, cfg.httpOpts, cfg.migrate, cfg.setup); err != nil {
		return errors.Join(ErrHTTPServiceRunFailed, err)
	}
	return nil
}

// BuildHTTPRuntime builds the startup runtime without starting listeners.
//
// BuildHTTPRuntime 仅构建启动上下文，不启动监听。
//
// Example:
//
//	rt, err := application.BuildHTTPRuntime("user-service", application.HTTP())
//	if err != nil {
//	    return err
//	}
//	defer rt.Container.Close()
func BuildHTTPRuntime(serviceName string, options ...Option) (*HTTPRuntime, error) {
	normalizedServiceName := strings.TrimSpace(serviceName)
	cfg, err := resolveRunConfig(normalizedServiceName, options...)
	if err != nil {
		return nil, err
	}
	rt, err := newHTTPRuntimeFunc(normalizedServiceName, cfg.httpOpts)
	if err != nil {
		return nil, errors.Join(ErrHTTPRuntimeBuildFailed, err)
	}
	return rt, nil
}

// Build is an alias of BuildHTTPRuntime.
//
// Build 是 BuildHTTPRuntime 的同义入口。
func Build(serviceName string, options ...Option) (*HTTPRuntime, error) {
	return BuildHTTPRuntime(serviceName, options...)
}
