package cmd

import "github.com/spf13/cobra"

// buildAllCmd 串行执行前后端构建。
//
// 中文说明：
// - 这里故意先 frontend 再 backend。
// - 这样如果后端发布包/运行目录需要携带最新前端产物，后续流程能直接拿到最新 dist。
var buildAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Build frontend + backend",
	RunE: func(cmd *cobra.Command, args []string) error {
		// frontend first so that backend can serve latest dist if needed
		if err := buildFrontendCmd.RunE(cmd, args); err != nil {
			return err
		}
		return buildBackendCmd.RunE(cmd, args)
	},
}

func init() {
	buildCmd.AddCommand(buildAllCmd)
}
