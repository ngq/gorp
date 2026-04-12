//go:build !windows

package cmd

import (
	"os"
	"syscall"
)

func isProcessAlive(pid int) (bool, error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}
	// Signal 0 checks existence.
	//
	// 中文说明：
	// - Unix 下向进程发送 signal 0 不会真正发信号，只用于探测进程是否存在。
	err = p.Signal(syscall.Signal(0))
	if err == nil {
		return true, nil
	}
	// ESRCH means no such process.
	if err == syscall.ESRCH {
		return false, nil
	}
	return false, err
}

func sendStopSignal(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	// We choose SIGTERM for graceful shutdown.
	return p.Signal(syscall.SIGTERM)
}
