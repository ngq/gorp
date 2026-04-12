package model

import "github.com/spf13/cobra"

// Cmd 是 model 命令组入口。
//
// 中文说明：
// - 它承载与数据库表结构生成相关的几个子命令：test / gen / api。
// - 当前这套命令更偏脚手架能力，而不是运行时业务逻辑。
var Cmd = &cobra.Command{
	Use:   "model",
	Short: "Model/API code generation",
}

func init() {
	Cmd.AddCommand(modelTestCmd)
	Cmd.AddCommand(modelGenCmd)
	Cmd.AddCommand(modelAPICmd)
}
