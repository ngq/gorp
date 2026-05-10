// Package application provides application startup entrypoints for gorp framework.
// This file centralizes internal helpers for startup flow normalization.
// Normalizes option handling and startup context checks before runtime construction.
//
// 应用启动包提供 gorp 框架的应用启动入口。
// 本文件集中管理启动流程的内部辅助逻辑。
// 在 runtime 构建前统一处理选项归一化与启动 context 检查。
package application

import (
	"context"
	"strings"
)

// resolveRunConfig resolves and normalizes startup options.
//
// resolveRunConfig 解析并归一化启动配置。
func resolveRunConfig(serviceName string, options ...Option) (runConfig, error) {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return runConfig{}, ErrServiceNameRequired
	}

	cfg := runConfig{httpEnabled: true}
	for _, opt := range options {
		if opt != nil {
			opt.apply(&cfg)
		}
	}
	if !cfg.httpEnabled {
		return runConfig{}, ErrNoServiceDeclared
	}
	return cfg, nil
}

// ensureStartupContext validates the context state before startup.
//
// ensureStartupContext 在启动前校验 context 状态。
func ensureStartupContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

// composeSetup chains two setup callbacks in declaration order.
//
// composeSetup 以声明顺序串联两个 setup 回调。
func composeSetup(prev, next SetupFunc) SetupFunc {
	switch {
	case prev == nil:
		return next
	case next == nil:
		return prev
	default:
		return func(rt *HTTPRuntime) error {
			if err := prev(rt); err != nil {
				return err
			}
			return next(rt)
		}
	}
}
