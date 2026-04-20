package cmd

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed all:templates/release all:templates/multi-flat all:templates/multi-flat-wire
var releaseTemplateFS embed.FS

var templatePackOut string
var templatePackName string

// templatePackCmd 生成可供 `gorp new from-release` 消费的模板 zip 包。
//
// 中文说明：
// - 它会把 release 模板目录打成一个标准 zip，默认文件名为 `gorp-template.zip`；
// - zip 内部会保留 `templates/project/**` 目录结构，以便 `new from-release` 直接消费；
// - 当前 release 路径覆盖公开 starter 模板，可按 template 选择单服务或多服务模板；
// - 这是维护者/发布者使用的补充交付工具，不是普通 starter 创建流程的默认入口。
var templatePackCmd = &cobra.Command{
	Use:   "pack",
	Short: "Pack release starter template into gorp-template.zip",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateReleaseStarterTemplate(templatePackName); err != nil {
			return err
		}
		out := templatePackOut
		if out == "" {
			out = filepath.Join(".tmp", defaultReleaseTemplateAsset(templatePackName))
		}
		srcFS, srcRoot := releaseTemplateSource(templatePackName)
		if err := zipDirectoryFromFS(srcFS, srcRoot, out, "templates/project"); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "created template asset[%s]: %s\n", templateDisplayName(templatePackName), out)
		fmt.Fprintf(cmd.OutOrStdout(), "hint: upload this zip to GitHub Release as %s\n", filepath.Base(out))
		return nil
	},
}

func init() {
	templatePackCmd.Flags().StringVar(&templatePackOut, "out", "", "output zip path (default depends on --template)")
	templatePackCmd.Flags().StringVar(&templatePackName, "template", starterTemplateBase, "release starter template: base, golayout, golayout-wire, multi-flat, multi-flat-wire")
	templateCmd.AddCommand(templatePackCmd)
}
