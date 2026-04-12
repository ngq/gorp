//go:build !windows

package cmd

import (
	"os"
	"syscall"
)

func gracefulStopSignal() os.Signal {
	// 中文说明：
	// - Unix 下优先使用 SIGTERM 作为优雅退出信号。
	return syscall.SIGTERM
}
