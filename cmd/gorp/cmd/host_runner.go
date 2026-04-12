package cmd

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// runHostManaged 统一处理 Host 模式下的启动、信号监听与优雅关闭。
//
// 中文说明：
// - 这是 framework 通用服务启动抽象第二轮的命令层共用 helper；
// - HTTP / Cron / gRPC 在注册完 Hostable 后，都可以复用这段逻辑；
// - 这样命令层不再为每种服务各自重复写 signal/shutdown 模板。
func runHostManaged(h contract.Host, logger contract.Logger, startMessage, shutdownMessage, stoppedMessage string) error {
	if startMessage != "" {
		logger.Info(startMessage)
	}
	if err := h.Start(context.Background()); err != nil {
		return err
	}

	sigs := []os.Signal{os.Interrupt}
	if runtime.GOOS != "windows" {
		sigs = append(sigs, syscall.SIGTERM)
	}
	ctx, stop := signal.NotifyContext(context.Background(), sigs...)
	defer stop()
	<-ctx.Done()

	if shutdownMessage != "" {
		logger.Info(shutdownMessage)
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := h.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if stoppedMessage != "" {
		logger.Info(stoppedMessage)
	}
	return nil
}
