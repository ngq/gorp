package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

var swaggerGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate swagger2 docs using swag",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use `swag` binary (installed via go get github.com/swaggo/swag/cmd/swag).
		// Generate under ./docs by default.
		outDir := filepath.Join("docs")
		_ = os.MkdirAll(outDir, 0o755)

		c := exec.Command("swag", "init", "-g", filepath.Join("cmd", "app", "main.go"), "-o", outDir)
		c.Stdout = cmd.OutOrStdout()
		c.Stderr = cmd.ErrOrStderr()
		if err := c.Run(); err != nil {
			return fmt.Errorf("swag init: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "swagger2 generated at %s\n", outDir)
		return nil
	},
}

var swaggerCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Scan Go handlers and validate Swagger comment annotations (@Summary, @Param, @Router)",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		var warnings []string
		validMethods := map[string]bool{
			"get": true, "post": true, "put": true, "delete": true,
			"patch": true, "options": true, "head": true,
		}

		err = filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.Contains(path, "vendor") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(wd, path)
			lines := strings.Split(string(content), "\n")
			for lineNo, l := range lines {
				trimmed := strings.TrimSpace(l)
				if !strings.HasPrefix(trimmed, "//") {
					continue
				}
				comment := strings.TrimSpace(strings.TrimPrefix(trimmed, "//"))
				if strings.HasPrefix(comment, "@Router") {
					parts := strings.Fields(comment)
					if len(parts) < 2 {
						warnings = append(warnings, fmt.Sprintf("%s:%d: @Router missing path and method", rel, lineNo+1))
					} else {
						pathAndMethod := parts[1]
						if !strings.HasPrefix(pathAndMethod, "/") {
							warnings = append(warnings, fmt.Sprintf("%s:%d: @Router path should start with '/'", rel, lineNo+1))
						}
						if len(parts) >= 3 {
							method := strings.Trim(strings.ToLower(parts[2]), "[]")
							if !validMethods[method] {
								warnings = append(warnings, fmt.Sprintf("%s:%d: invalid HTTP method %q in @Router", rel, lineNo+1, method))
							}
						}
					}
				} else if strings.HasPrefix(comment, "@Param") {
					parts := strings.Fields(comment)
					if len(parts) < 5 {
						warnings = append(warnings, fmt.Sprintf("%s:%d: @Param requires at least 4 arguments (name in type dataType)", rel, lineNo+1))
					}
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		if len(warnings) > 0 {
			fmt.Fprintf(cmd.ErrOrStderr(), "Found %d swagger annotation issues:\n", len(warnings))
			for _, w := range warnings {
				fmt.Fprintln(cmd.ErrOrStderr(), "  - "+w)
			}
			return fmt.Errorf("swagger check failed with %d warnings", len(warnings))
		}

		fmt.Fprintln(cmd.OutOrStdout(), "swagger annotations check passed (0 errors)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(swaggerCmd)
	swaggerCmd.AddCommand(swaggerGenCmd)
	swaggerCmd.AddCommand(swaggerCheckCmd)
}
