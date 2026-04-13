package cmd

import "github.com/spf13/cobra"

// appCmd 是 legacy HTTP runtime 命令组。
//
// 中文说明：
// - 该命令组当前仍保留，主要用于兼容旧的 runtime CLI 路径；
// - starter 项目的公开推荐启动方式已经回到项目自己的 `cmd/*/main.go`；
// - 因此这里不再作为 CLI 的主心智入口。
var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Legacy runtime commands for HTTP services",
}

func init() {
	appCmd.AddCommand(appStartCmd)
}
