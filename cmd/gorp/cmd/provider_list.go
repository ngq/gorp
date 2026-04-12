package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// providerListCmd 列出当前 bootstrap 后可见的 provider 名称。
//
// 中文说明：
// - 该命令会先执行一轮标准 bootstrap，再从具体容器实现中读取 provider 名列表。
// - 这里故意没有把 ProviderNames 放进 contract.Container，而是保持为实现细节能力。
var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered providers (by name)",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, c, err := bootstrap()
		if err != nil {
			return err
		}

		// container implementation exposes provider names via internal method.
		// 中文说明：通过具体类型获取 provider 名列表，保持 contract 最小化。
		if lister, ok := c.(interface{ ProviderNames() []string }); ok {
			for _, name := range lister.ProviderNames() {
				fmt.Fprintln(cmd.OutOrStdout(), name)
			}
			return nil
		}
		return fmt.Errorf("container does not support provider listing")
	},
}

func init() {
	providerCmd.AddCommand(providerListCmd)
}
