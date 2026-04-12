package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

// buildSelfCmd 编译当前 gorp CLI 自身。
//
// 中文说明：
// - 产物默认输出到仓库根目录下的 .tmp/，避免污染正式发布目录。
// - backend 子命令本质上复用的也是这条构建链路。
// - Windows 下会自动补 .exe 后缀，保证产物可直接执行。
var buildSelfCmd = &cobra.Command{
	Use:   "self",
	Short: "Build gorp CLI (self)",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := repoRootFromCWD()
		if err != nil {
			return err
		}
		outDir := filepath.Join(root, ".tmp")
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return err
		}
		out := filepath.Join(outDir, "gorp")
		if runtime.GOOS == "windows" {
			out += ".exe"
		}

		c := exec.Command("go", "build", "-o", out, "./cmd/gorp")
		c.Dir = root
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "built: %s\n", out)
		return nil
	},
}

func init() {
	buildCmd.AddCommand(buildSelfCmd)
}
