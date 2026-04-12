package cmd

import (
	"github.com/ngq/gorp/cmd/gorp/cmd/model"
)

func init() {
	// 中文说明：
	// - model 子命令树实际定义在 cmd/model 子包中。
	// - 这里仅负责把它挂到根命令上，保持主命令目录结构清晰。
	rootCmd.AddCommand(model.Cmd)
}
