package cmd

import "github.com/spf13/cobra"

// middlewareCmd 是 middleware 命令组（一级命令）：
//
// 用于承载中间件相关的脚手架能力：
// - `gorp middleware list`：列出已有业务中间件
// - `gorp middleware new`：创建业务中间件骨架
// - `gorp middleware migrate`：迁移 gin-contrib 中间件到本项目（并修正 import）
var middlewareCmd = &cobra.Command{
	Use:   "middleware",
	Short: "Middleware scaffolding tools",
}

func init() {
	rootCmd.AddCommand(middlewareCmd)
}
