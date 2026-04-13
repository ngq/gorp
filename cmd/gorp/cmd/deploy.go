package cmd

import "github.com/spf13/cobra"

// deployCmd 是部署辅助命令组。
//
// 中文说明：
// - 当前采用 SSH 直连远端机器的发布模型；
// - 子命令分别覆盖 backend / frontend / all / rollback 等典型发布动作；
// - 发布目录约定为 releases/<version> + current 软链 + shared 持久目录；
// - 这组命令更偏母仓/工程辅助能力，不是 starter 项目的公开推荐工作流。
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Auxiliary deployment tools (SSH-based)",
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
