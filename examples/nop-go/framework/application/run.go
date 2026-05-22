// Package application provides application startup entrypoints for gorp framework.
// This file exposes the main startup entrypoints: Run and Build.
// Supports direct boot and runtime-only build workflows for HTTP services.
//
// 应用启动包提供 gorp 框架的应用启动入口。
// 本文件暴露主启动入口：Run 和 Build。
// 支持直接启动和仅构建 runtime 的 HTTP 服务工作流。
package application

import (
	"context"
	"errors"
)

// Run boots the default HTTP mainline with application options.
// Service name is read from config file (app.name) automatically.
//
// 说明：业务运行入口仍是项目自己的 main；application 只提供启动装配 helper。
// Run 启动默认 HTTP 主线，服务名自动从配置文件 app.name 读取。
//
// Example:
//
//	err := application.Run(
//	    application.HTTP(),
//	    application.WithProviders(myProvider),
//	    application.WithHTTPRoutes(func(router transportcontract.Router, c runtimecontract.Container) error {
//	        registerRoutes(router)
//	        return nil
//	    }),
//	)
func Run(options ...Option) error {
	return RunContext(context.Background(), options...)
}

// Start is an alias of Run.
//
// Start 是 Run 的同义入口。
func Start(options ...Option) error {
	return Run(options...)
}

// RunContext boots the default mainline with an explicit context.
// Service name is read from config file (app.name) automatically.
//
// 当前语义：仅在启动前检查取消，运行中关闭流程仍由 bootstrap 处理。
// RunContext 使用显式 context 启动默认主线，服务名从配置自动读取。
func RunContext(ctx context.Context, options ...Option) error {
	if err := ensureStartupContext(ctx); err != nil {
		return errors.Join(ErrStartupCanceled, err)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	cfg, err := resolveRunConfig(options...)
	if err != nil {
		return err
	}
	// serviceName 从配置读取，在 bootHTTPService 内部处理
	if err := bootHTTPService(cfg.httpOpts, cfg.migrate, cfg.setup); err != nil {
		return errors.Join(ErrHTTPServiceRunFailed, err)
	}
	return nil
}

// BuildHTTPRuntime builds the startup runtime without starting listeners.
// Service name is read from config (app.name) automatically.
//
// BuildHTTPRuntime 仅构建启动上下文，不启动监听。
// 服务名自动从配置 app.name 读取。
//
// Example:
//
//	rt, err := application.BuildHTTPRuntime(application.HTTP())
//	if err != nil {
//	    return err
//	}
//	defer rt.Container.Close()
func BuildHTTPRuntime(options ...Option) (*HTTPRuntime, error) {
	cfg, err := resolveRunConfig(options...)
	if err != nil {
		return nil, err
	}
	rt, err := newHTTPRuntimeFunc(cfg.httpOpts)
	if err != nil {
		return nil, errors.Join(ErrHTTPRuntimeBuildFailed, err)
	}
	return rt, nil
}

// Build is an alias of BuildHTTPRuntime.
//
// Build 是 BuildHTTPRuntime 的同义入口。
func Build(options ...Option) (*HTTPRuntime, error) {
	return BuildHTTPRuntime(options...)
}
