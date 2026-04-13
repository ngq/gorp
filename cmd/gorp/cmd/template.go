package cmd

import "github.com/spf13/cobra"

// templateCmd 是模板资产相关命令组。
//
// 中文说明：
// - 这组命令主要面向“模板生产者/维护者”，而不是普通业务开发者；
// - 当前包含：
//   1. `gorp template pack`：生成 release 模板资产 zip
//   2. `gorp template diff`：比较当前项目与模板之间的差异（只读）
//   3. `gorp template upgrade`：预览/说明模板升级路径（当前仍偏治理辅助）
//   4. `gorp template version`：查看当前 CLI 内嵌模板类型
// - 该命令组的定位是“模板治理与模板资产维护”，不是业务运行时命令。
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Template governance and release asset tools",
}

func init() {
	rootCmd.AddCommand(templateCmd)
}
