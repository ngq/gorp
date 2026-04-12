package cmd

import "github.com/spf13/cobra"

// deployCmd 是部署命令组。
//
// 中文说明：
// - 当前采用 SSH 直连远端机器的发布模型。
// - 子命令分别覆盖 backend / frontend / all / rollback 等典型发布动作。
// - 发布目录约定为 releases/<version> + current 软链 + shared 持久目录。
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy tools (SSH-based)",
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
