package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/spf13/cobra"
)

// providerNewCmd 创建一个新的 provider 骨架目录。
//
// 中文说明：
// - 会同时生成 contract.go / provider.go / service.go 三个基础文件。
// - 适合快速起一个新的业务 provider，再由开发者继续补真实逻辑。
// - 生成完成后，还需要手动在 bootstrap 或其他装配入口中挂载该 provider。
var providerNewCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"create", "init"},
	Short:   "Create a new provider skeleton under app/provider/",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := repoRootFromCWD()
		if err != nil {
			return err
		}

		_, c, err := bootstrap()
		if err != nil {
			return err
		}

		name, err := promptString(cmd.InOrStdin(), cmd.OutOrStdout(), "请输入服务名称(服务凭证)：", "", true)
		if err != nil {
			return err
		}
		if err := requireIdent(name, "provider name"); err != nil {
			return err
		}

		// check duplicate provider name
		if lister, ok := c.(interface{ ProviderNames() []string }); ok {
			for _, existing := range lister.ProviderNames() {
				if existing == name {
					return fmt.Errorf("provider already exists: %s", name)
				}
			}
		}

		folder, err := promptString(cmd.InOrStdin(), cmd.OutOrStdout(), "请输入服务所在目录名称(默认: 同服务名称)：", name, false)
		if err != nil {
			return err
		}
		if err := requireIdent(folder, "provider folder"); err != nil {
			return err
		}

		targetDir := absJoin(root, "app", "provider", folder)
		if dirExists(targetDir) {
			return fmt.Errorf("target folder already exists: %s", targetDir)
		}
		if err := ensureDir(targetDir); err != nil {
			return err
		}

		pub := toPublicGoName(name)

		contractSrc := fmt.Sprintf(`package %s

const %sKey = %q

type Service interface {
	Foo() string
}
`, name, pub, name)

		providerSrc := fmt.Sprintf(`package %s

import "github.com/ngq/gorp/framework/contract"

type %sProvider struct{}

func NewProvider() *%sProvider { return &%sProvider{} }

func (p *%sProvider) Name() string { return %q }
func (p *%sProvider) IsDefer() bool { return false }
func (p *%sProvider) Provides() []string { return []string{%sKey} }

func (p *%sProvider) Register(c contract.Container) error {
	c.Bind(%sKey, func(contract.Container) (any, error) {
		return NewService(), nil
	}, true)
	return nil
}

func (p *%sProvider) Boot(contract.Container) error { return nil }
`, name, pub, pub, pub, pub, name, pub, pub, pub, pub, pub, pub)

		serviceSrc := fmt.Sprintf(`package %s

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) Foo() string { return "" }
`, name)

		if err := writeGoFile(filepath.Join(targetDir, "contract.go"), contractSrc); err != nil {
			return err
		}
		if err := writeGoFile(filepath.Join(targetDir, "provider.go"), providerSrc); err != nil {
			return err
		}
		if err := writeGoFile(filepath.Join(targetDir, "service.go"), serviceSrc); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "创建服务成功, 文件夹地址: %s\n", targetDir)
		fmt.Fprintln(cmd.OutOrStdout(), "请不要忘记挂载新创建的服务")
		return nil
	},
}

func init() {
	providerCmd.AddCommand(providerNewCmd)
}

func repoRootFromCWD() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 中文说明：
	// - 从当前工作目录开始向上查找仓库根目录；
	// - 当前使用 `go.mod` + `cmd/gorp` 作为最小特征，避免把任意有 go.mod 的上层目录误判为本项目根目录；
	// - 如果一直找不到，则退回当前目录，让调用方仍可继续使用/手工覆盖。
	cur := wd
	for {
		if fileExists(filepath.Join(cur, "go.mod")) && dirExists(filepath.Join(cur, "cmd", "gorp")) {
			return cur, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return wd, nil
}
