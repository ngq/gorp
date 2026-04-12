package cmd

import "github.com/spf13/cobra"

// providerCmd 是 provider 脚手架命令组。
//
// 中文说明：
// - 用于承载 provider 相关的创建与查看能力。
// - 当前主要包括：
//   1. `gorp provider list`：列出当前已注册 provider
//   2. `gorp provider new`：生成新的业务 provider 骨架
var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Provider scaffolding tools",
}

func init() {
	rootCmd.AddCommand(providerCmd)
}
