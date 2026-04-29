package cmd

import "github.com/spf13/cobra"

// newCmd 创建一个新的业务项目脚手架。
//
// 说明：为了兼顾“默认离线可用”与“未来 GitHub 发布”，我们把 new 设计成：
//
// 1) `gorp new`（默认，内置模板）：
//    - 使用本仓库内置的模板（go:embed），生成一个最小可运行项目。
//    - go.mod 会写入 replace 指向本地框架源码路径，适合开发调试。
//
// 2) `gorp new from-release`（联网）：
//    - 从 GitHub Release 下载模板包并生成项目。
//    - 面向后续发布给外部用户使用的 starter / template 场景。
var newCmd = &cobra.Command{
	Use:     "new [multi-wire]",
	Short:   "Create a new project from embedded starter templates",
	GroupID: commandGroupStarter,
	Long: `Create a new project from embedded starter templates.

Default starter path:
  - gorp new            : default single-service quick start
  - gorp new multi-wire : default multi-service quick start

Supplementary delivery path:
  - gorp new from-release : published release asset delivery path

On-demand starter selection:
  - Use --template only after gorp new and gorp new multi-wire no longer match your project shape.
  - golayout          : 单服务 / 默认起步
  - multi-flat-wire   : 多服务 / 默认微服务起步
  - multi-independent : 多服务 / 更强独立治理

Decision rule:
  - If you are not sure, start with gorp new.
  - If you already know you need multi-service structure, use gorp new multi-wire.
  - Reach for --template only after those two default paths no longer fit.
  - Authentication, RBAC, admin, and other business permissions belong in the generated project, not in the starter template.
  - After generation, start services from the generated project's own cmd/*/main.go entrypoints.`,
}

func init() {
	rootCmd.AddCommand(newCmd)
	// 子命令：
	// - from-release：从 GitHub Release 下载模板（在 new_from_release.go 的 init() 中挂载）
}
