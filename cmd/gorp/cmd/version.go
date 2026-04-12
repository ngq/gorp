package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// versionCmd 打印当前 CLI 的构建版本信息。
//
// 中文说明：
// - 这三个字段通常由构建脚本在编译时通过 ldflags 注入。
// - 本地直接 `go run` 时会看到默认值（dev/none/unknown）。
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "version=%s\ncommit=%s\nbuildDate=%s\n", Version, Commit, BuildDate)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
