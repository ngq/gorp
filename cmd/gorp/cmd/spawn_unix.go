//go:build !windows

package cmd

import (
	"os/exec"
	"syscall"
)

func prepareDetached(cmd *exec.Cmd) {
	// 中文说明：
	// - Unix 下通过 Setsid=true 让子进程进入新的 session，脱离当前终端控制。
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
