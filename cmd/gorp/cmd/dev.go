package cmd

import "github.com/spf13/cobra"

// devCmd 为调试辅助命令组（对齐教程 Ch19-20）。
//
// 策略说明：
// - `dev backend`：监听 Go 文件变更，变更后重启后端（当前内部实现仍通过 `go run ./cmd/gorp app start` 启动兼容链路）
// - `dev frontend`：启动 `npm run dev`（不代理，前端自己占端口）
// - `dev all`：同时启动前后端；并启动一个 proxy 端口（dev.port），实现：
//   - 先转发 backend，backend 返回 404 再转发 frontend（教程同款策略）
//
// 说明：这里的”proxy”是额外的 dev gateway，不会影响生产模式。
//
// 中文说明：
// - 它面向母仓与本地开发场景，不服务于 starter 项目的公开主路径；
// - 子命令会组合文件监听、子进程重启、前后端代理等能力；
// - 顶层 `gorp dev` 默认只展示帮助，避免误执行某个重型流程。
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Auxiliary local development tools (watch and proxy)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(devCmd)
}
