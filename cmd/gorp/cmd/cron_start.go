package cmd

import (
	"context"
	"fmt"
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
	cronStartDaemon bool
	cronPID         string
)

func init() {
	cronCmd.AddCommand(cronStartCmd)
	cronCmd.AddCommand(cronListCmd)
	cronCmd.AddCommand(cronStateCmd)
	cronCmd.AddCommand(cronStopCmd)
	cronCmd.AddCommand(cronRestartCmd)

	cronStartCmd.Flags().BoolVarP(&cronStartDaemon, "daemon", "d", false, "run in background")
	cronStartCmd.Flags().StringVar(&cronPID, "pid", "", "pid file path (default app runtime path/gorp-cron.pid)")
}

var cronStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start cron worker",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cronStartDaemon {
			exe, err := os.Executable()
			if err != nil {
				return err
			}
			childArgs := []string{"cron", "start"}
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
			outF, errF, cleanup, err := openDaemonLogFiles(bc, "gorp-cron")
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

		logAny, err := c.Make(contract.LogKey)
		if err != nil {
			return err
		}
		logger := logAny.(contract.Logger)

		jobsAny, err := c.Make(contract.CronRuntimeConfiguratorKey)
		if err != nil {
			return err
		}
		configurator := jobsAny.(contract.CronRuntimeConfigurator)
		jobCount, err := configurator.ConfigureCronRuntime(c)
		if err != nil {
			return err
		}

		cronAny, err := c.Make(contract.CronKey)
		if err != nil {
			return err
		}
		cr := cronAny.(contract.Cron)

		appAny, err := c.Make(app.AppKey)
		if err != nil {
			return err
		}
		appSvc := appAny.(app.App)
		if err := ensureRuntimeDirs(appSvc); err != nil {
			return err
		}
		pidPath, err := writeCronPID(c, cronPID, os.Getpid())
		if err != nil {
			return err
		}
		defer os.Remove(pidPath)

		logger.Info("cron worker starting", contract.Field{Key: "jobs", Value: jobCount})

		hostAny, err := c.Make(contract.HostKey)
		if err == nil {
			h := hostAny.(contract.Host)
			// 中文说明：
			// - 优先走 Host 模式，把 cron worker 纳入统一生命周期管理；
			// - 这样 cron 与 HTTP / gRPC 的启动/关闭路径可以继续收口到同一套框架骨架；
			// - 命令层不再重复维护 signal / shutdown 细节。
			cronHostable := host.NewCronService("cron", cr)
			if err := h.RegisterService("cron", cronHostable); err != nil {
				return err
			}
			return runHostManaged(h, logger, "", "cron shutdown signal received", "cron worker stopped")
		}

		// 中文说明：
		// - Host 不可用时，保留 direct mode 作为兼容回退路径；
		// - 这样 framework 主线可以渐进收口，而不强迫所有环境同步迁移。
		cr.Start()
		sigs := []os.Signal{os.Interrupt}
		if runtime.GOOS != "windows" {
			sigs = append(sigs, syscall.SIGTERM)
		}
		ctx, stop := signal.NotifyContext(context.Background(), sigs...)
		defer stop()
		<-ctx.Done()
		logger.Info("cron shutdown signal received")

		stopped := cr.Stop()
		select {
		case <-stopped.Done():
			logger.Info("cron worker stopped")
			return nil
		case <-time.After(10 * time.Second):
			return fmt.Errorf("timeout waiting for cron to stop")
		}
	},
}
