package cmd

import "github.com/spf13/cobra"

// buildBackendCmd 编译后端。
//
// 中文说明：
// - 在本仓库中，后端二进制就是 `gorp` 本身（`./cmd/gorp`）。
// - 因此 build backend 直接复用 build self。
var buildBackendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Build backend (same as build self)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return buildSelfCmd.RunE(cmd, args)
	},
}

func init() {
	buildCmd.AddCommand(buildBackendCmd)
}
