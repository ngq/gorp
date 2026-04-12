package model

import (
	"fmt"

	"github.com/ngq/gorp/framework/contract"

	"github.com/spf13/cobra"
)

// modelTestCmd 验证数据库连通性，并列出当前库中的表。
//
// 中文说明：
// - 这是使用 model gen/api 之前最推荐先跑的一步。
// - 如果这里失败，说明配置、驱动或数据库可达性仍有问题，后续生成命令大概率也会失败。
var modelTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test DB connectivity and list tables",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, c, err := bootstrap()
		if err != nil {
			return err
		}
		insAny, err := c.Make(contract.DBInspectorKey)
		if err != nil {
			return err
		}
		ins := insAny.(contract.DBInspector)

		if err := ins.Ping(cmd.Context()); err != nil {
			return fmt.Errorf("db ping: %w", err)
		}
		tables, err := ins.Tables(cmd.Context())
		if err != nil {
			return err
		}
		for _, t := range tables {
			fmt.Fprintln(cmd.OutOrStdout(), t.Name)
		}
		return nil
	},
}
