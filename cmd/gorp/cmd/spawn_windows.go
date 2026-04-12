//go:build windows

package cmd

import (
	"os/exec"
	"syscall"
)

func prepareDetached(cmd *exec.Cmd) {
	// 中文说明：
	// - Windows 下通过 CREATE_NEW_PROCESS_GROUP 把子进程放到新的进程组中。
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}
