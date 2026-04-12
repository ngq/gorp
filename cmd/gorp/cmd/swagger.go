package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// swaggerCmd 是 Swagger/OpenAPI2 工具命令组。
//
// 中文说明：
// - 当前主要提供 `swagger gen`，用于基于 swag 生成 swagger2 文档。
// - 后续 openapi3 转换则由 `openapi gen` 负责。
var swaggerCmd = &cobra.Command{
	Use:   "swagger",
	Short: "Swagger/OpenAPI tools",
}

var swaggerGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate swagger2 docs using swag",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use `swag` binary (installed via go get github.com/swaggo/swag/cmd/swag).
		// Generate under ./docs by default.
		outDir := filepath.Join("docs")
		_ = os.MkdirAll(outDir, 0o755)

		c := exec.Command("swag", "init", "-g", filepath.Join("cmd", "gorp", "main.go"), "-o", outDir)
		c.Stdout = cmd.OutOrStdout()
		c.Stderr = cmd.ErrOrStderr()
		if err := c.Run(); err != nil {
			return fmt.Errorf("swag init: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "swagger2 generated at %s\n", outDir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(swaggerCmd)
	swaggerCmd.AddCommand(swaggerGenCmd)
}