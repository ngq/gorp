package cmd

import (
	"fmt"
	"path"
	"strings"
	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/deploy"

	"github.com/spf13/cobra"
)

// deployRollbackCmd 把 current 软链回切到历史版本。
//
// 中文说明：
// - 该命令复用 deploy 的 releases/current 目录模型。
// - 当前第二个参数 frontend|backend|all 仅为教程兼容保留，实际逻辑统一回滚整个 release 目录。
// - 回滚过程会先尝试停止旧进程，再切换 current，最后重新启动。
var deployRollbackCmd = &cobra.Command{
	Use:   "rollback <version> <frontend|backend|all>",
	Short: "Rollback to a previous release version",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		version := args[0]
		_ = args[1] // end(frontend/backend/all) kept for tutorial compatibility
		_, c, err := bootstrap(WithAppEnv(deployEnv))
		if err != nil {
			return err
		}
		cfgAny, err := c.Make(contract.ConfigKey)
		if err != nil {
			return err
		}
		cfg := cfgAny.(contract.Config)


		sshAny, err := c.Make(contract.SSHKey)
		if err != nil {
			return err
		}
		sshSvc := sshAny.(contract.SSHService)

		var depCfg deployConfig
		if err := cfg.Unmarshal("deploy", &depCfg); err != nil {
			return err
		}
		if deployRemoteFolder != "" {
			depCfg.RemoteFolder = deployRemoteFolder
		}
		if depCfg.RemoteFolder == "" {
			return fmt.Errorf("deploy.remote_folder is required")
		}
		if len(depCfg.Connections) == 0 {
			return fmt.Errorf("deploy.connections is required")
		}

		for _, name := range depCfg.Connections {
			fmt.Fprintf(cmd.OutOrStdout(), "connecting: %s\n", name)
			client, err := sshSvc.GetClient(name)
			if err != nil {
				return err
			}
			defer client.Close()

			recordRoot := strings.TrimRight(depCfg.RemoteFolder, "/")
			relDir := path.Join(recordRoot, "releases", version)

			// Ensure target release exists
			_, err = deploy.RunRemote(client, fmt.Sprintf("test -d %s", relDir))
			if err != nil {
				return fmt.Errorf("release not found on remote: %s", relDir)
			}

			// stop old (best-effort)
			_, _ = deploy.RunRemote(client, fmt.Sprintf("cd %s/current && ./gorp app stop", recordRoot))
			// switch current
			_, err = deploy.RunRemote(client, fmt.Sprintf("ln -sfn %s %s", relDir, path.Join(recordRoot, "current")))
			if err != nil {
				return err
			}
			// ensure binary executable
			_, err = deploy.RunRemote(client, fmt.Sprintf("chmod +x %s", path.Join(recordRoot, "current", "gorp")))
			if err != nil {
				return err
			}
			// start new
			out, err := deploy.RunRemote(client, fmt.Sprintf("cd %s/current && ./gorp app start -d", recordRoot))
			fmt.Fprintf(cmd.OutOrStdout(), "[%s] start: %s\n", name, strings.TrimSpace(out))
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	deployCmd.AddCommand(deployRollbackCmd)
	deployRollbackCmd.Flags().StringVar(&deployEnv, "env", "", "use config/app.<env>.yaml overlay")
	deployRollbackCmd.Flags().StringVar(&deployRemoteFolder, "remote-folder", "", "override deploy.remote_folder")
}
