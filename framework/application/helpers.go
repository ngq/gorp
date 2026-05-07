// Application scenarios:
// - Centralize small internal helpers used by the application startup flow.
// - Normalize option handling and startup context checks before runtime construction.
// - Keep helper logic separate from exported APIs so the public entrypoints stay easy to scan.
//
// 适用场景：
// - 集中管理 application 启动流程中使用的小型内部辅助逻辑。
// - 在 runtime 构建前统一处理选项归一化与启动 context 检查。
// - 将辅助逻辑与导出 API 分离，让公开入口保持更易浏览的结构。
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
