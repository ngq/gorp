package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/goroutine"
	"github.com/ngq/gorp/framework/provider/app"
	"github.com/ngq/gorp/framework/provider/host"

	"github.com/spf13/cobra"
)

// grpcCmd 是 legacy gRPC runtime 命令组。
//
// 中文说明：
// - 该命令组当前仍保留，主要用于兼容旧的 gRPC runtime CLI 路径；
// - starter 项目的公开推荐路径应优先走项目自己的启动入口，而不是这里的 runtime 命令。
var grpcCmd = &cobra.Command{
	Use:   "grpc",
	Short: "Legacy runtime commands for gRPC services",
}

var (
	grpcStartDaemon bool
	grpcPID         string
	grpcAddr        string
)

func init() {
	grpcStartCmd.Flags().BoolVarP(&grpcStartDaemon, "daemon", "d", false, "run in background")
	grpcStartCmd.Flags().StringVar(&grpcPID, "pid", "", "pid file path (default app runtime path/gorp-grpc.pid)")
	grpcStartCmd.Flags().StringVar(&grpcAddr, "addr", ":9090", "listen address")
}

var grpcStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a gRPC server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if grpcStartDaemon {
			exe, err := os.Executable()
			if err != nil {
				return err
			}
			childArgs := []string{"grpc", "start"}
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
			outF, errF, cleanup, err := openDaemonLogFiles(bc, "gorp-grpc")
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

		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			return err
		}
		builderAny, err := c.Make(contract.GRPCRuntimeBuilderKey)
		if err != nil {
			return err
		}
		builder := builderAny.(contract.GRPCRuntimeBuilder)
		srv := builder.BuildGRPCServer()

		appAny, err := c.Make(app.AppKey)
		if err != nil {
			return err
		}
		appSvc := appAny.(app.App)
		if err := ensureRuntimeDirs(appSvc); err != nil {
			return err
		}
		pidPath, err := writeGrpcPID(c, grpcPID, os.Getpid())
		if err != nil {
			return err
		}
		defer os.Remove(pidPath)

		hostAny, err := c.Make(contract.HostKey)
		if err == nil {
			h := hostAny.(contract.Host)
			// 中文说明：
			// - gRPC 优先纳入 Host 管理，与 HTTP / Cron 共享同一套生命周期抽象；
			// - 命令层只负责 listener 与 runtime builder，不再重复维护 shutdown 模板；
			// - 这样后续 framework 级通用服务启动抽象可以继续向单一入口收口。
			grpcHostable := host.NewGRPCService("grpc", srv, lis)
			if err := h.RegisterService("grpc", grpcHostable); err != nil {
				return err
			}
			logger.Info("grpc server starting", contract.Field{Key: "addr", Value: lis.Addr().String()})
			return runHostManaged(h, logger, "", "grpc shutdown signal received", "grpc server stopped gracefully")
		}

		// 中文说明：
		// - Host 不可用时，继续保留 direct mode 作为兼容路径；
		// - 这样当前阶段不会因为抽象深化而破坏旧链路可用性。
		errCh := make(chan error, 1)
		goroutine.SafeGo(cmd.Context(), c, func(ctx context.Context) {
			logger.Info("grpc server starting", contract.Field{Key: "addr", Value: lis.Addr().String()})
			errCh <- srv.Serve(lis)
		})

		sigs := []os.Signal{os.Interrupt}
		if runtime.GOOS != "windows" {
			sigs = append(sigs, syscall.SIGTERM)
		}
		ctx, stop := signal.NotifyContext(context.Background(), sigs...)
		defer stop()

		select {
		case <-ctx.Done():
			logger.Info("grpc shutdown signal received")
			srv.GracefulStop()
			return nil
		case err := <-errCh:
			return err
		}
	},
}

func init() {
	rootCmd.AddCommand(grpcCmd)
	grpcCmd.AddCommand(grpcStartCmd)
	grpcCmd.AddCommand(grpcStateCmd)
	grpcCmd.AddCommand(grpcStopCmd)
	grpcCmd.AddCommand(grpcRestartCmd)
}
