package cmd

import "github.com/spf13/cobra"

// newCmd 创建一个新的业务项目脚手架。
//
// 说明：为了兼顾“离线可用”与“未来 GitHub 发布”，我们把 new 拆成两条路径：
//
// 1) `gorp new`（默认，离线）：
//    - 使用本仓库内置的模板（go:embed），生成一个最小可运行项目。
//    - go.mod 会写入 replace 指向本地框架源码路径，适合开发调试。
//
// 2) `gorp new from-release`（联网）：
//    - 从 GitHub Release 下载模板包并生成项目。
//    - 面向后续发布给外部用户使用的 starter / template 场景。
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new project (offline template by default)",
	Long: `Create a new project from starter templates.

Template recommendation:
  - golayout      : default choice for single-service / lightweight-DDD-first projects
  - golayout-wire : recommended choice for single-service production projects using Wire
  - multi-flat    : default choice for multi-service projects
  - multi-flat-wire : recommended choice for multi-service production projects using Wire
  - base          : minimal skeleton for custom structure

If you are not sure which template to pick, start with golayout or golayout-wire.`,
}

func init() {
	rootCmd.AddCommand(newCmd)
	// 子命令：
	// - offline：离线模板（go:embed），在 new_offline.go 的 init() 中挂载
	// - from-release：从 GitHub Release 下载模板（在 new_from_release.go 的 init() 中挂载）
	// 注意：不要在这里重复挂载 newOfflineCmd，避免重复
}
