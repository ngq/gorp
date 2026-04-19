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
	Use:   "new [wire|multi|multi-wire]",
	Short: "Create a new project from embedded starter templates",
	Long: `Create a new project from embedded starter templates.

Default recommendation:
  - Install gorp first, then use 'gorp new' as the primary starter path.
  - Bare 'gorp new' is the default single-service quick start.
  - Use 'gorp new wire' when you specifically want the Wire-based single-service template.
  - Use 'gorp new multi' / 'gorp new multi-wire' when you already know you want a multi-service structure.
  - Use 'gorp new from-release' only when you specifically need published release assets or fixed-version starter delivery.

High-frequency intents:
  - gorp new            : default single-service quick start
  - gorp new wire       : single-service Wire template
  - gorp new multi      : default multi-service template
  - gorp new multi-wire : multi-service Wire template

Advanced template matrix:
  - golayout        : default choice for single-service projects
  - golayout-wire   : advanced public single-service template with Wire assembly
  - multi-flat      : default choice for multi-service projects
  - multi-flat-wire : advanced public multi-service template with Wire assembly
  - base            : minimal skeleton for custom structure

Important:
  - Positional intent is the primary public path.
  - Explicit --template has higher priority than positional intent.
  - Use --template when you want custom structure or advanced composition.
  - Authentication, RBAC, admin, and other business permissions should be implemented in the generated project, not assumed by the starter template.

If you are not sure which path to pick:
  - Single service quick start: gorp new
  - Single service with Wire: gorp new wire
  - Multi-service: gorp new multi or gorp new multi-wire`,
}

func init() {
	rootCmd.AddCommand(newCmd)
	// 子命令：
	// - from-release：从 GitHub Release 下载模板（在 new_from_release.go 的 init() 中挂载）
}
