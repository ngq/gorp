package cmd

import "github.com/spf13/cobra"

// docCmd 是文档相关命令组。
//
// 中文说明：
// - 这里承载“生成/校验运维文档、配置文档、CLI 参考”等子命令。
// - 当前主要子命令是 `gorp doc gen`，用于把项目当前状态导出为 docs/manual/ 下的 markdown。
var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "Documentation generation tools",
}

func init() {
	rootCmd.AddCommand(docCmd)
}
