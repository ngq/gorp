package cmd

import "github.com/spf13/cobra"

// commandCmd 是控制台命令脚手架命令组。
//
// 中文说明：
// - 用于承载 console command 相关的创建与查看能力。
// - 这类命令通常对应 app/console/command 下的业务命令骨架。
// - 这组命令属于按需进入的脚手架工具，不是默认起步入口。
var commandCmd = &cobra.Command{
	Use:     "command",
	Short:   "On-demand console tools",
	GroupID: commandGroupAdvanced,
	Long: `On-demand console tools.

Use this command group only after the default starter path has stopped matching your current task.
It is not the default starter path.`,
}

func init() {
	rootCmd.AddCommand(commandCmd)
}
