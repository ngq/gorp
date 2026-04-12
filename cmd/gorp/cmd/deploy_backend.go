package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/deploy"
	"github.com/ngq/gorp/framework/provider/app"

	"github.com/spf13/cobra"
)

// deployBackendConfig 描述后端发布阶段的专属配置。
//
// 中文说明：
// - Goos/Goarch 用于本地交叉编译远端目标二进制。
// - PreAction/PostAction 用于在远端发布前后插入自定义 shell 动作。
type deployBackendConfig struct {
	Goos       string   `mapstructure:"goos"`
	Goarch     string   `mapstructure:"goarch"`
	PreAction  []string `mapstructure:"pre_action"`
	PostAction []string `mapstructure:"post_action"`
}

// deployFrontendConfig 描述前端发布阶段的钩子动作配置。
type deployFrontendConfig struct {
	PreAction  []string `mapstructure:"pre_action"`
	PostAction []string `mapstructure:"post_action"`
}

// deployConfig 是 deploy 顶层配置结构。
//
// 中文说明：
// - RemoteFolder 是远端项目根目录。
// - Connections 是需要发布到的 SSH 主机名列表。
// - Frontend/Backend 则分别描述各自的附加发布动作。
type deployConfig struct {
	RemoteFolder string               `mapstructure:"remote_folder"`
	Connections  []string             `mapstructure:"connections"`
	Frontend     deployFrontendConfig `mapstructure:"frontend"`
	Backend      deployBackendConfig  `mapstructure:"backend"`
}

var (
	deployEnv          string
	deployVersion      string
	deployRemoteFolder string
)

var deployBackendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Deploy backend to remote servers over SSH",
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

		// 中文说明：
		// - backend 部署流程分两段：先在本地打包，再逐台远端上传并切换 current 软链。
		// - 这样可以保证所有机器拿到的是同一份本地构建产物。
		// - 同时也为后续灰度/回滚保留了 releases/<version> 历史目录。
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

		// package locally
		appAny, err := c.Make(app.AppKey)
		if err != nil {
			return err
		}
		appSvc := appAny.(app.App)
		pkgDir, err := buildDeployPackage(appSvc.BasePath(), depCfg.Backend, ver)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "packaged: %s\n", pkgDir)

		// deploy to each connection
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

			// ensure base dirs
			_, err = deploy.RunRemote(client, fmt.Sprintf("mkdir -p %s %s", releasesDir, sharedDir))
			if err != nil {
				return err
			}
			// ensure shared storage
			_, err = deploy.RunRemote(client, fmt.Sprintf("mkdir -p %s", path.Join(sharedDir, "storage")))
			if err != nil {
				return err
			}

			// pre actions
			for _, a := range depCfg.Backend.PreAction {
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
			// Copy current -> new release (best-effort) so that backend-only deploy does not wipe frontend assets.
			_, _ = deploy.RunRemote(client, fmt.Sprintf("if [ -d %s/current ]; then cp -a %s/current/. %s/; fi", remoteRoot, remoteRoot, relDir))

			// upload backend overlay (binary/config/.env)
			fmt.Fprintf(cmd.OutOrStdout(), "[%s] uploading backend overlay to %s\n", name, relDir)
			if err := deploy.UploadDir(client, pkgDir, relDir); err != nil {
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

			// ensure binary executable
			_, err = deploy.RunRemote(client, fmt.Sprintf("chmod +x %s", path.Join(remoteRoot, "current", "gorp")))
			if err != nil {
				return err
			}

			// start new
			out, err := deploy.RunRemote(client, fmt.Sprintf("cd %s/current && ./gorp app start -d", remoteRoot))
			fmt.Fprintf(cmd.OutOrStdout(), "[%s] start: %s\n", name, strings.TrimSpace(out))
			if err != nil {
				return err
			}

			// post actions
			for _, a := range depCfg.Backend.PostAction {
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

var deployAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Deploy all components (backend + frontend)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure a single version is used for backend + frontend.
		//
		// 中文说明：
		// - all 本质上只是顺序执行 backend + frontend。
		// - 但两者必须共享同一个版本号，否则会被发布到不同 release 目录。
		_ = resolveDeployVersion()

		if err := deployBackendCmd.RunE(cmd, args); err != nil {
			return err
		}
		return deployFrontendCmd.RunE(cmd, args)
	},
}

func init() {
	deployCmd.AddCommand(deployBackendCmd)
	deployCmd.AddCommand(deployAllCmd)

	for _, c := range []*cobra.Command{deployBackendCmd, deployAllCmd} {
		c.Flags().StringVar(&deployEnv, "env", "", "use config/app.<env>.yaml overlay")
		c.Flags().StringVar(&deployVersion, "version", "", "deploy version (default timestamp)")
		c.Flags().StringVar(&deployRemoteFolder, "remote-folder", "", "override deploy.remote_folder")
	}
}

func buildDeployPackage(base string, backend deployBackendConfig, version string) (string, error) {
	outDir := filepath.Join(base, "deploy", version)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}

	// 中文说明：
	// - 这里生成的是“后端发布覆盖包”，而不是完整仓库镜像。
	// - 当前最小集合包括：gorp 二进制、config 目录、可选 .env。
	// - 前端静态资源由 deploy frontend 单独处理。
	// Build gorp binary
	exe := filepath.Join(outDir, "gorp")
	goos := backend.Goos
	if goos == "" {
		goos = "linux"
	}
	goarch := backend.Goarch
	if goarch == "" {
		goarch = "amd64"
	}

	exeAbs, err := filepath.Abs(exe)
	if err != nil {
		return "", err
	}

	c := exec.Command("go", "build", "-o", exeAbs, "./cmd/gorp")
	c.Dir = base
	c.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return "", err
	}

	// Copy config dir
	if err := copyDir(filepath.Join(base, "config"), filepath.Join(outDir, "config")); err != nil {
		return "", err
	}

	// Copy optional .env
	if _, err := os.Stat(filepath.Join(base, ".env")); err == nil {
		_ = copyFile(filepath.Join(base, ".env"), filepath.Join(outDir, ".env"))
	}

	return outDir, nil
}

func copyDir(src, dst string) error {
	// 中文说明：
	// - 这里按相对路径递归复制整个目录树。
	// - 当前主要用于把 config 目录打包进 deploy/<version>/。
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(p, target)
	})
}

func copyFile(src, dst string) error {
	// 中文说明：
	// - 这里采用直接整文件读写，适合配置文件、小型发布产物复制。
	// - 当前发布包体量不大，没必要引入更复杂的流式复制抽象。
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

// resolveDeployVersion returns the deploy version used by all deploy subcommands.
//
// 中文说明：
// - `deploy backend/frontend/all` 可能被连续调用（比如 deploy all 先后端再前端）。
// - 为了保证一次发布“同一个版本目录”，我们把 version 缓存在全局 deployVersion 中。
func resolveDeployVersion() string {
	if deployVersion == "" {
		deployVersion = time.Now().Format("20060102150405")
	}
	return deployVersion
}
