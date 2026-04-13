package cmd

import (
	"github.com/spf13/cobra"
)

// cronCmd 是 legacy cron runtime 命令组。
//
// 中文说明：
// - 该命令组当前仍保留，主要用于兼容旧的 cron worker CLI 路径；
// - starter 项目的公开推荐路径应优先走项目自己的启动入口，而不是这里的 runtime 命令。
var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Legacy runtime commands for cron workers",
}

func init() {
	rootCmd.AddCommand(cronCmd)
}
