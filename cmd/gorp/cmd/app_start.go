package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/provider/app"
	"github.com/ngq/gorp/framework/provider/host"

	"github.com/spf13/cobra"
)

var (
	appStartDaemon bool
	appStartPID    string
)

func init() {
	appStartCmd.Flags().BoolVarP(&appStartDaemon, "daemon", "d", false, "run in background")
	appStartCmd.Flags().StringVar(&appStartPID, "pid", "", "pid file path (default app runtime path/gorp.pid)")
}

var appStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start HTTP server (legacy runtime path)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if appStartDaemon {
			exe, err := os.Executable()
			if err != nil {
				return err
			}
			childArgs := []string{"app", "start"}
			for _, a := range os.Args[1:] {
				if a == "--daemon" || a == "-d" {
					continue
				}
				childArgs = append(childArgs, a)
			}

			c := exec.Command(exe, childArgs...)
			_, bc, err := bootstrap()
			if err != nil {
				return err
			}
			outF, errF, cleanup, err := openDaemonLogFiles(bc, "gorp-app")
			if err != nil {
				return err
			}
			defer cleanup()
			c.Stdout = outF
			c.Stderr = errF
			prepareDetached(c)
			if err := c.Start(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "started in background (pid=%d)\n", c.Process.Pid)
			return nil
		}

		_, c, err := bootstrap()
		if err != nil {
			return err
		}

		// 中文说明：
		// - 命令层现在只调用一个统一的 app HTTP 装配入口；
		// - 具体的 migrate / route assembly 细节不再直接散落在 start 命令中；
		// - 这是 framework 抽离阶段"先把装配动作收进单独入口"的第一步。
		configAny, err := c.Make(contract.HTTPRuntimeConfiguratorKey)
		if err != nil {
			return err
		}
		configurator := configAny.(contract.HTTPRuntimeConfigurator)
		if err := configurator.ConfigureHTTPRuntime(c); err != nil {
			return err
		}

		appAny, err := c.Make(app.AppKey)
		if err != nil {
			return err
		}
		appSvc := appAny.(app.App)
		if err := ensureRuntimeDirs(appSvc); err != nil {
			return err
		}
		pidPath := defaultPIDPath(appSvc)
		if appStartPID != "" {
			pidPath = appStartPID
		}
		if err := writePID(pidPath, os.Getpid()); err != nil {
			return err
		}
		defer os.Remove(pidPath)

		// 中文说明：
		// - 使用 Host 接口统一管理服务生命周期；
		// - HTTP 服务作为 Hostable 注册到 Host；
		// - 信号处理通过 Host 的 Shutdown 方法统一管理。
		return runWithHost(c)
	},
}

// runWithHost 使用 Host 接口统一管理服务生命周期。
//
// 中文说明：
// - 获取 Host 服务并注册 HTTP 服务；
// - 通过 Host.Start 统一启动；
// - 信号处理通过 Host 的 Shutdown 方法统一管理。
func runWithHost(c contract.Container) error {
	// 获取 Host 服务
	hostAny, err := c.Make(contract.HostKey)
	if err != nil {
		// 如果 Host 不可用，回退到旧方式
		return runHTTP(c)
	}
	h := hostAny.(contract.Host)

	// 获取 HTTP 服务
	httpAny, err := c.Make(contract.HTTPKey)
	if err != nil {
		return err
	}
	httpSvc := httpAny.(contract.HTTP)

	// 获取日志服务
	logAny, err := c.Make(contract.LogKey)
	if err != nil {
		return err
	}
	logger := logAny.(contract.Logger)

	// 注册 HTTP 服务到 Host
	httpHostable := host.NewHTTPService("http", httpSvc)
	if err := h.RegisterService("http", httpHostable); err != nil {
		return err
	}

	logger.Info("starting http server")
	return runHostManaged(h, logger, "", "shutdown signal received", "http server stopped gracefully")
}

func runHTTP(c contract.Container) error {
	hAny, err := c.Make(contract.HTTPKey)
	if err != nil {
		return err
	}
	h := hAny.(contract.HTTP)

	lAny, err := c.Make(contract.LogKey)
	if err != nil {
		return err
	}
	logger := lAny.(contract.Logger)

	logger.Info("starting http server")

	errCh := make(chan error, 1)
	go func() {
		if err := h.Run(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				errCh <- nil
				return
			}
			errCh <- err
		}
	}()

	sigs := []os.Signal{os.Interrupt}
	if runtime.GOOS != "windows" {
		sigs = append(sigs, syscall.SIGTERM)
	}
	ctx, stop := signal.NotifyContext(context.Background(), sigs...)
	defer stop()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return err
		}
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := h.Shutdown(shutdownCtx); err != nil {
		return err
	}
	logger.Info("http server stopped gracefully")
	return nil
}