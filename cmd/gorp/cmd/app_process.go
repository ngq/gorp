package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// appStateCmd 通过 pidfile 查看 HTTP app 进程状态。
//
// 中文说明：
// - 它不会主动探测端口，而是以 pidfile + 进程存活检查为准。
// - 如果发现 pidfile 已陈旧，会顺手删除，避免误导后续 stop/restart。
var appStateCmd = &cobra.Command{
	Use:   "state",
	Short: "Show app process state (via pid file)",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, c, err := bootstrap()
		if err != nil {
			return err
		}

		appSvc, err := getAppService(c)
		if err != nil {
			return err
		}
		pidPath := defaultPIDPath(appSvc)
		pid, err := readPID(pidPath)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "not running (pidfile not found): %s\n", pidPath)
			return nil
		}

		ok, err := isProcessAlive(pid)
		if err != nil {
			return err
		}
		if !ok {
			_ = os.Remove(pidPath)
			fmt.Fprintf(cmd.OutOrStdout(), "not running (stale pidfile removed): pid=%d\n", pid)
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "running: pid=%d (pidfile=%s)\n", pid, pidPath)
		return nil
	},
}

// appStopCmd 停止 HTTP app 进程，并等待其退出。
var appStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop app process (best-effort)",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, c, err := bootstrap()
		if err != nil {
			return err
		}

		appSvc, err := getAppService(c)
		if err != nil {
			return err
		}
		pidPath := defaultPIDPath(appSvc)
		pid, err := readPID(pidPath)
		if err != nil {
			return fmt.Errorf("pidfile not found: %s", pidPath)
		}

		if err := sendStopSignal(pid); err != nil {
			return err
		}

		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			alive, err := isProcessAlive(pid)
			if err != nil {
				return err
			}
			if !alive {
				fmt.Fprintln(cmd.OutOrStdout(), "stopped")
				return nil
			}
			time.Sleep(250 * time.Millisecond)
		}

		return fmt.Errorf("timeout waiting for process to stop: pid=%d", pid)
	},
}

var appRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart app process",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := appStopCmd.RunE(cmd, args); err != nil {
			return err
		}
		// small delay to avoid port reuse issues
		time.Sleep(300 * time.Millisecond)
		return appStartCmd.RunE(cmd, args)
	},
}

func init() {
	appCmd.AddCommand(appStateCmd)
	appCmd.AddCommand(appStopCmd)
	appCmd.AddCommand(appRestartCmd)
}
