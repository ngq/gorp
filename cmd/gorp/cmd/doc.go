package cmd

import "github.com/spf13/cobra"

// docCmd 是文档相关命令组。
//
// 中文说明：
// - 这里承载“生成/校验运维文档、配置文档、CLI 参考”等子命令。
// - 当前主要子命令是 `gorp doc gen`，用于把项目当前状态导出为 docs/manual/ 下的 markdown。
// - 这组命令属于按需查阅的文档产物链，不是默认起步入口。
var docCmd = &cobra.Command{
	Use:     "doc",
	Short:   "On-demand documentation tools",
	GroupID: commandGroupAdvanced,
	Long: `On-demand documentation tools.

Use this command group only after the default starter path and high-frequency developer tools
no longer answer the current documentation question. Read it as a reference toolchain, not as a starting path.`,
}

func init() {
	rootCmd.AddCommand(docCmd)
}
