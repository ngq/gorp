package cmd

import "github.com/spf13/cobra"

// rootCmd 是 gorp CLI 的根命令。
//
// 中文说明：
// - `gorp` 的公开主心智应是：framework + starter templates + developer toolchain；
// - 一级命令当前可按三层理解：
//   1. toolchain 主命令：`new / template / proto / model / migrate`；
//   2. scaffolding / generation / docs 辅助命令：`provider / middleware / command / doc / swagger / openapi`；
//   3. legacy runtime / auxiliary ops：`app / grpc / cron / build / dev / deploy`；
// - 运行时命令当前仍保留，但不应再作为第一印象压过主工具链命令；
// - 这个文件本身不承载具体业务逻辑，职责是建立“命令树入口”；
// - 当前 CLI 前缀统一为 `gorp`；
// - 新增代码、示例命令与文档都应以 `gorp` 为准，避免再次出现多套前缀并存。
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

func init() {
	// 中文说明：
	// - 当前一级命令仍主要在各自文件的 init() 中向 rootCmd 追加注册；
	// - 这里不再特殊对待 appCmd，避免 root.go 给出“runtime 优先挂载”的误导；
	// - 若后续进入更大一轮 Cobra 命令树重构，再评估是否统一改为集中注册。
}
