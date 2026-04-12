//go:build windows

package cmd

import "syscall"

func isProcessAlive(pid int) (bool, error) {
	h, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		// The process may not exist.
		//
		// 中文说明：
		// - Windows 下这里采用能否成功打开进程句柄来判断进程是否还活着。
		return false, nil
	}
	defer syscall.CloseHandle(h)
	return true, nil
}

func sendStopSignal(pid int) error {
	h, err := syscall.OpenProcess(syscall.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(h)
	return syscall.TerminateProcess(h, 0)
}
