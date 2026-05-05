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
// - 这组命令属于按需使用的文档产物链，不是默认起步入口。
var swaggerCmd = &cobra.Command{
	Use:     "swagger",
	Short:   "On-demand Swagger tools",
	GroupID: commandGroupAdvanced,
	Long: `On-demand Swagger tools.

Use this command group only after you already know you need API documentation artifacts.
It is not the default starter path.`,
}

var (
	swaggerEntryFile string
	swaggerOutputDir string
	swaggerParseInternal bool
)

var swaggerGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate swagger2 docs using swag",
	Long: `Generate swagger2 docs using swaggo/swag.

Examples:
  # 单服务项目（默认入口 cmd/main.go）
  gorp swagger gen

  # 微服务项目（指定服务入口）
  gorp swagger gen -g services/user/cmd/main.go

  # 指定输出目录
  gorp swagger gen -o services/user/docs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use `swag` binary (installed via go install github.com/swaggo/swag/cmd/swag@latest).
		if _, err := exec.LookPath("swag"); err != nil {
			return fmt.Errorf("swag not found, install: go install github.com/swaggo/swag/cmd/swag@latest")
		}

		// 默认值处理
		entry := swaggerEntryFile
		if entry == "" {
			entry = filepath.Join("cmd", "main.go")
		}

		output := swaggerOutputDir
		if output == "" {
			output = "docs"
		}

		_ = os.MkdirAll(output, 0o755)

		// 构建命令参数
		swagArgs := []string{"init", "-g", entry, "-o", output}
		if swaggerParseInternal {
			swagArgs = append(swagArgs, "--parseInternal")
		}

		c := exec.Command("swag", swagArgs...)
		c.Stdout = cmd.OutOrStdout()
		c.Stderr = cmd.ErrOrStderr()
		if err := c.Run(); err != nil {
			return fmt.Errorf("swag init: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "swagger2 generated at %s\n", output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(swaggerCmd)
	swaggerCmd.AddCommand(swaggerGenCmd)

	swaggerGenCmd.Flags().StringVarP(&swaggerEntryFile, "entry", "g", "", "entry file (default cmd/main.go)")
	swaggerGenCmd.Flags().StringVarP(&swaggerOutputDir, "output", "o", "", "output directory (default docs)")
	swaggerGenCmd.Flags().BoolVar(&swaggerParseInternal, "parse-internal", true, "parse internal packages")
}