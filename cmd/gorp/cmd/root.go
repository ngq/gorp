package cmd

import "github.com/spf13/cobra"

// rootCmd 是 gorp CLI 的根命令。
//
// 中文说明：
// - `gorp` 的公开心智模型：framework + starter templates + developer toolchain；
// - 一级命令分为两层：
//  1. toolchain 主命令：`new / template / proto / model`；
//  2. scaffolding / generation / docs 辅助命令：`provider / middleware / command / doc / swagger / openapi`；
//
// - legacy runtime 命令（app / grpc / cron / build / dev / deploy）已退役；
// - 用户通过项目自己的 `cmd/*/main.go` 启动服务，不依赖 CLI runtime。
var rootCmd = &cobra.Command{
	Use:   "gorp",
	Short: "Framework, starter templates, and developer tooling for gorp",
}

// Execute 执行整个 Cobra 命令树。
//
// 中文说明：
// - main 函数通常只需要调用这一层；
// - 具体命令匹配、flag 解析、help 输出、RunE 执行，都由 Cobra 在这里统一调度。
func Execute() error {
	return rootCmd.Execute()
}
