//go:build windows

package cmd

import "os"

func gracefulStopSignal() os.Signal {
	// 中文说明：
	// - Windows 下没有 SIGTERM，这里退回到 os.Interrupt 语义。
	return os.Interrupt
}
