// Package application provides application startup entrypoints for gorp framework.
// This file defines stable startup and assembly error sentinels.
// Let callers use errors.Is against well-known bootstrap and runtime failures.
//
// 应用启动包提供 gorp 框架的应用启动入口。
// 本文件定义稳定的启动与装配错误哨兵值。
// 让调用方可以通过 errors.Is 判断已知的启动与运行时失败。
package application

import "errors"

var (
	// ErrNoServiceDeclared indicates that no runnable service has been declared.
	//
	// ErrNoServiceDeclared 表示没有声明可运行服务。
	ErrNoServiceDeclared = errors.New("application: no service declared")

	// ErrHTTPRouteRegistrationFailed indicates that HTTP route registration failed.
	//
	// ErrHTTPRouteRegistrationFailed 表示 HTTP 路由注册失败。
	ErrHTTPRouteRegistrationFailed = errors.New("application: http route registration failed")

	// ErrHTTPRuntimeUnavailable indicates that the HTTP runtime is unavailable during route setup.
	//
	// ErrHTTPRuntimeUnavailable 表示 HTTP 路由注册阶段缺少可用 runtime。
	ErrHTTPRuntimeUnavailable = errors.New("application: http runtime unavailable")

	// ErrSetupFailed indicates that the setup callback failed.
	//
	// ErrSetupFailed 表示 setup 回调执行失败。
	ErrSetupFailed = errors.New("application: setup failed")

	// ErrMigrateFailed indicates that the migrate callback failed.
	//
	// ErrMigrateFailed 表示 migrate 回调执行失败。
	ErrMigrateFailed = errors.New("application: migrate failed")

	// ErrStartupCanceled indicates that startup was canceled before boot completed.
	//
	// ErrStartupCanceled 表示启动前 context 已取消。
	ErrStartupCanceled = errors.New("application: startup canceled")

	// ErrHTTPServiceRunFailed indicates that booting the default HTTP service failed.
	//
	// ErrHTTPServiceRunFailed 表示 HTTP 服务启动失败。
	ErrHTTPServiceRunFailed = errors.New("application: http service run failed")

	// ErrHTTPRuntimeBuildFailed indicates that building the HTTP runtime failed.
	//
	// ErrHTTPRuntimeBuildFailed 表示 HTTP runtime 构建失败。
	ErrHTTPRuntimeBuildFailed = errors.New("application: http runtime build failed")

	// ErrGRPCServiceRunFailed indicates that booting the gRPC service failed.
	//
	// ErrGRPCServiceRunFailed 表示 gRPC 服务启动失败。
	ErrGRPCServiceRunFailed = errors.New("application: grpc service run failed")

	// ErrGRPCRuntimeBuildFailed indicates that building the gRPC runtime failed.
	//
	// ErrGRPCRuntimeBuildFailed 表示 gRPC runtime 构建失败。
	ErrGRPCRuntimeBuildFailed = errors.New("application: grpc runtime build failed")
)
