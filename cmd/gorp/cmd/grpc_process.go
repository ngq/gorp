package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

const grpcPidFileName = "gorp-grpc.pid"

// grpcPIDPath 返回 gRPC worker 的 pidfile 路径。
func grpcPIDPath(c any, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	appSvc, err := getAppService(c)
	if err != nil {
		return "", err
	}
	return filepath.Join(appSvc.RuntimePath(), grpcPidFileName), nil
}

func writeGrpcPID(c any, override string, pid int) (string, error) {
	pidPath, err := grpcPIDPath(c, override)
	if err != nil {
		return "", err
	}
	return pidPath, writePID(pidPath, pid)
}

func readGrpcPID(c any, override string) (int, string, error) {
	pidPath, err := grpcPIDPath(c, override)
	if err != nil {
		return 0, "", err
	}
	pid, err := readPID(pidPath)
	return pid, pidPath, err
}

// grpcStateCmd 查看 gRPC worker 当前状态。
var grpcStateCmd = &cobra.Command{
	Use:   "state",
	Short: "Show gRPC process state (via pid file)",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, c, err := bootstrap()
		if err != nil {
			return err
		}
		pid, pidPath, err := readGrpcPID(c, grpcPID)
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

var grpcStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop gRPC process (best-effort)",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, c, err := bootstrap()
		if err != nil {
			return err
		}
		pid, pidPath, err := readGrpcPID(c, grpcPID)
		if err != nil {
			return fmt.Errorf("pidfile not found")
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
				_ = os.Remove(pidPath)
				fmt.Fprintln(cmd.OutOrStdout(), "stopped")
				return nil
			}
			time.Sleep(250 * time.Millisecond)
		}
		return fmt.Errorf("timeout waiting for process to stop: pid=%d", pid)
	},
}

var grpcRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart gRPC process",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := grpcStopCmd.RunE(cmd, args); err != nil {
			return err
		}
		time.Sleep(300 * time.Millisecond)
		return grpcStartCmd.RunE(cmd, args)
	},
}
