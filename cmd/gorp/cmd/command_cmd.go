package cmd

import "github.com/spf13/cobra"

// commandCmd 是控制台命令脚手架命令组。
//
// 中文说明：
// - 用于承载 console command 相关的创建与查看能力。
// - 这类命令通常对应 app/console/command 下的业务命令骨架。
var commandCmd = &cobra.Command{
	Use:   "command",
	Short: "Console command scaffolding tools",
}

func init() {
	rootCmd.AddCommand(commandCmd)
}
