package cmd

import "github.com/spf13/cobra"

// buildCmd 为编译命令组（对齐教程 Ch18-20）。
//
// 提供统一命令入口：
// - `gorp build self`：编译当前 CLI（./cmd/gorp）到 ./.tmp/gorp.exe（Windows）或 ./.tmp/gorp（Unix）
// - `gorp build backend`：等价于 build self（本仓库 backend 二进制就是 gorp）
// - `gorp build frontend`：若存在 frontend/ 且 package.json 存在，则执行 `npm run build`（优先 npm；后续可加 pnpm/bun）
// - `gorp build all`：frontend + backend
// buildCmd 是统一构建入口。
//
// 中文说明：
// - 目标是把不同构建动作收口到一个命令树下，降低日常使用心智负担。
// - 实现刻意保持轻量：只调用现有 Go / npm 构建链路，不额外引入复杂打包编排器。
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build tools (frontend/backend/self)",
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
