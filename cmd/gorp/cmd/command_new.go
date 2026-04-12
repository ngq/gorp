package cmd

import (
	"fmt"
	"path/filepath"
	"github.com/spf13/cobra"
)

// commandNewCmd 创建一个新的 console command 骨架。
//
// 中文说明：
// - 生成结果位于 `app/console/command/<folder>/`。
// - 当前只生成最小 cobra.Command 定义，方便后续再挂接到命令注册入口。
var commandNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new console command skeleton under app/console/command/",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := repoRootFromCWD()
		if err != nil {
			return err
		}

		name, err := promptString(cmd.InOrStdin(), cmd.OutOrStdout(), "请输入命令名称：", "", true)
		if err != nil {
			return err
		}
		if err := requireIdent(name, "command name"); err != nil {
			return err
		}

		folder, err := promptString(cmd.InOrStdin(), cmd.OutOrStdout(), "请输入命令所在目录名称(默认: 同命令名称)：", name, false)
		if err != nil {
			return err
		}
		if err := requireIdent(folder, "command folder"); err != nil {
			return err
		}

		targetDir := absJoin(root, "app", "console", "command", folder)
		if dirExists(targetDir) {
			return fmt.Errorf("target folder already exists: %s", targetDir)
		}
		if err := ensureDir(targetDir); err != nil {
			return err
		}

		pub := toPublicGoName(name)
		src := fmt.Sprintf(`package %s

import (
	"fmt"

	"github.com/spf13/cobra"
)

var %sCommand = &cobra.Command{
	Use:   %q,
	Short: %q,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "todo")
		return nil
	},
}
`, name, pub, name, name)

	if err := writeGoFile(filepath.Join(targetDir, folder+".go"), src); err != nil {
		return err
	}

		fmt.Fprintf(cmd.OutOrStdout(), "创建命令成功, 文件夹地址: %s\n", targetDir)
		fmt.Fprintln(cmd.OutOrStdout(), "请不要忘记挂载新创建的命令")
		return nil
	},
}

func init() {
	commandCmd.AddCommand(commandNewCmd)
}
