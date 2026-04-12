package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

// commandListCmd 列出当前 CLI 下所有一级命令。
//
// 中文说明：
// - 它读取的是 rootCmd.Commands()，因此列出来的是当前 CLI 已挂载命令，而不是 app/console 目录文件。
// - 这更适合检查“命令树装配结果”，而不是做源码扫描。
var commandListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all top-level CLI commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		children := rootCmd.Commands()
		out := make([]*cobra.Command, 0, len(children))
		for _, c := range children {
			if c.Hidden {
				continue
			}
			out = append(out, c)
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })

		for _, c := range out {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", c.Name(), c.Short)
		}
		return nil
	},
}

func init() {
	commandCmd.AddCommand(commandListCmd)
}
