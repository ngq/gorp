package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/deploy"
	"github.com/ngq/gorp/framework/provider/app"

	"github.com/spf13/cobra"
)

// deployFrontendCmd 部署前端静态资源（Vue/Vite dist）。
//
// 中文说明：
// - 本仓库的远端发布模型是：releases/<version> + current 软链 + shared/storage。
// - 仅部署 frontend 时，如果直接上传到新版本目录，会导致后端二进制缺失。
// - 因此在远端先把 current 目录复制到 releases/<version>（best-effort），再覆盖上传 frontend/dist。
//   这样 backend-only/frontend-only 发布都不会互相”擦掉对方文件”。
var deployFrontendCmd = &cobra.Command{
	Use:   "frontend",
	Short: "Deploy frontend assets (frontend/dist) to remote servers over SSH",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		ver := resolveDeployVersion()

		// locate local dist
		appAny, err := c.Make(app.AppKey)
		if err != nil {
			return err
		}
		appSvc := appAny.(app.App)
		localDist := filepath.Join(appSvc.BasePath(), "frontend", "dist")
		st, err := os.Stat(localDist)
		if err != nil || !st.IsDir() {
			return fmt.Errorf("frontend dist not found: %s (please build frontend first)", localDist)
		}

		for _, name := range depCfg.Connections {
			fmt.Fprintf(cmd.OutOrStdout(), "connecting: %s\n", name)
			client, err := sshSvc.GetClient(name)
			if err != nil {
				return err
			}
			defer client.Close()

			remoteRoot := strings.TrimRight(depCfg.RemoteFolder, "/")
			releasesDir := path.Join(remoteRoot, "releases")
			sharedDir := path.Join(remoteRoot, "shared")
			relDir := path.Join(releasesDir, ver)

			// 中文说明：
			// - 前端部署同样遵循 releases/current/shared 目录模型。
			// - 关键点在于“先复制 current，再覆盖 dist”，这样不会把旧版本里的后端文件抹掉。
			// - 因而 frontend-only 发布可以和 backend-only 发布独立进行。
			// ensure base dirs
			_, err = deploy.RunRemote(client, fmt.Sprintf("mkdir -p %s %s", releasesDir, sharedDir))
			if err != nil {
				return err
			}
			_, err = deploy.RunRemote(client, fmt.Sprintf("mkdir -p %s", path.Join(sharedDir, "storage")))
			if err != nil {
				return err
			}

			// pre actions
			for _, a := range depCfg.Frontend.PreAction {
				out, err := deploy.RunRemote(client, a)
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] pre: %s\n%s\n", name, a, strings.TrimSpace(out))
				if err != nil {
					return err
				}
			}

			// prepare release dir
			_, err = deploy.RunRemote(client, fmt.Sprintf("mkdir -p %s", relDir))
			if err != nil {
				return err
			}
			// copy current -> new release (best-effort)
			_, _ = deploy.RunRemote(client, fmt.Sprintf("if [ -d %s/current ]; then cp -a %s/current/. %s/; fi", remoteRoot, remoteRoot, relDir))

			// upload dist overlay
			remoteDist := path.Join(relDir, "frontend", "dist")
			fmt.Fprintf(cmd.OutOrStdout(), "[%s] uploading frontend dist to %s\n", name, remoteDist)
			if err := deploy.UploadDir(client, localDist, remoteDist); err != nil {
				return err
			}

			// link shared storage inside release
			_, err = deploy.RunRemote(client, fmt.Sprintf("ln -sfn %s %s", path.Join(sharedDir, "storage"), path.Join(relDir, "storage")))
			if err != nil {
				return err
			}

			// stop old (best-effort)
			_, _ = deploy.RunRemote(client, fmt.Sprintf("cd %s/current && ./gorp app stop", remoteRoot))

			// switch current
			_, err = deploy.RunRemote(client, fmt.Sprintf("ln -sfn %s %s", relDir, path.Join(remoteRoot, "current")))
			if err != nil {
				return err
			}

			// ensure binary executable (if exists)
			_, _ = deploy.RunRemote(client, fmt.Sprintf("chmod +x %s || true", path.Join(remoteRoot, "current", "gorp")))

			// start new (may fail if backend never deployed)
			out, err := deploy.RunRemote(client, fmt.Sprintf("cd %s/current && ./gorp app start -d", remoteRoot))
			fmt.Fprintf(cmd.OutOrStdout(), "[%s] start: %s\n", name, strings.TrimSpace(out))
			if err != nil {
				return err
			}

			// post actions
			for _, a := range depCfg.Frontend.PostAction {
				out, err := deploy.RunRemote(client, a)
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] post: %s\n%s\n", name, a, strings.TrimSpace(out))
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
}

func init() {
	deployCmd.AddCommand(deployFrontendCmd)
	deployFrontendCmd.Flags().StringVar(&deployEnv, "env", "", "use config/app.<env>.yaml overlay")
	deployFrontendCmd.Flags().StringVar(&deployVersion, "version", "", "deploy version (default timestamp)")
	deployFrontendCmd.Flags().StringVar(&deployRemoteFolder, "remote-folder", "", "override deploy.remote_folder")
}
