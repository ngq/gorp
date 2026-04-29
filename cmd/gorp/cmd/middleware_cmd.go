package cmd

import "github.com/spf13/cobra"

// middlewareCmd 是 middleware 命令组（一级命令）：
//
// 用于承载中间件相关的脚手架能力：
// - `gorp middleware list`：列出已有业务中间件
// - `gorp middleware new`：创建业务中间件骨架
// - `gorp middleware migrate`：迁移 gin-contrib 中间件到本项目（并修正 import）
// - 这组命令属于按需进入的脚手架工具，不是默认起步入口。
var middlewareCmd = &cobra.Command{
	Use:     "middleware",
	Short:   "On-demand middleware tools",
	GroupID: commandGroupAdvanced,
	Long: `On-demand middleware tools.

Use this command group only after the default starter path has stopped matching your current task.
It is not the default starter path.`,
}

func init() {
	rootCmd.AddCommand(middlewareCmd)
}
