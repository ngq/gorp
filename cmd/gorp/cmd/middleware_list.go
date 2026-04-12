package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

// middlewareListCmd 列出当前项目中的”业务中间件”。
//
// 中文说明：
// - 业务中间件位于 `app/http/middleware/` 下，每一个子目录代表一个中间件。
// - 这里只列目录名，不解析具体 Go 文件内容。
var middlewareListCmd = &cobra.Command{
	Use:   "list",
	Short: "List business middlewares under app/http/middleware/",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := repoRootFromCWD()
		if err != nil {
			return err
		}
		base := filepath.Join(root, "app", "http", "middleware")
		entries, err := os.ReadDir(base)
		if err != nil {
			return err
		}

		items := make([]string, 0)
		for _, e := range entries {
			if e.IsDir() {
				items = append(items, e.Name())
			}
		}
		sort.Strings(items)
		for _, name := range items {
			fmt.Fprintln(cmd.OutOrStdout(), name)
		}
		return nil
	},
}

func init() {
	middlewareCmd.AddCommand(middlewareListCmd)
}
