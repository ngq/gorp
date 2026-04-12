package cmd

import "github.com/spf13/cobra"

// rootCmd 是 gorp CLI 的根命令。
//
// 中文说明：
// - 所有一级命令（app/grpc/cron/model/doc/new/deploy 等）都从这里挂载出去；
// - 这个文件本身不承载具体业务逻辑，职责是建立“命令树入口”；
// - 当前 CLI 前缀统一为 `gorp`；
// - 新增代码、示例命令与文档都应以 `gorp` 为准，避免再次出现多套前缀并存。
var rootCmd = &cobra.Command{
	Use:   "gorp",
	Short: "Gin-based web framework CLI",
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
	// 先挂载 app 命令组；其余命令组在各自文件的 init() 中继续补充。
	rootCmd.AddCommand(appCmd)
}
