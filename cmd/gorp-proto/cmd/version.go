// Package cmd version 子命令
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// 版本信息（编译时注入）
	version   = "dev"
	gitCommit = "unknown"
	buildDate = "unknown"
)

// versionCmd version 子命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  `显示 gorp-proto CLI 工具的版本信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gorp-proto %s\n", version)
		fmt.Printf("  Git Commit: %s\n", gitCommit)
		fmt.Printf("  Build Date: %s\n", buildDate)
		fmt.Printf("  Go Version: %s\n", runtime.Version())
		fmt.Printf("  Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// SetVersionInfo 设置版本信息（编译时调用）
func SetVersionInfo(v, commit, date string) {
	version = v
	gitCommit = commit
	buildDate = date
}