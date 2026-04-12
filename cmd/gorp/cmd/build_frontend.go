package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// buildFrontendCmd 编译前端（对齐教程 Ch18-20）。
//
// 说明：
// - 仅负责调用前端项目自身的 build 命令，不做依赖安装。
// - 默认使用 npm（可扩展支持 pnpm/bun）。
var buildFrontendCmd = &cobra.Command{
	Use:   "frontend",
	Short: "Build frontend (npm run build)",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := repoRootFromCWD()
		if err != nil {
			return err
		}
		frontendDir := filepath.Join(root, "frontend")
		pkgJSON := filepath.Join(frontendDir, "package.json")
		if !fileExists(pkgJSON) {
			return fmt.Errorf("frontend package.json not found: %s", pkgJSON)
		}

		// 中文说明：
		// - 这里只负责触发前端项目自己的构建脚本，不负责 npm install。
		// - 如果依赖未安装，错误会直接由 npm 返回，保持职责单一。
		c := exec.Command("npm", "run", "build")
		c.Dir = frontendDir
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "frontend build done")
		return nil
	},
}

func init() {
	buildCmd.AddCommand(buildFrontendCmd)
}
