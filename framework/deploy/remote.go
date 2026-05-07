// Application scenarios:
// - Provide a lightweight SSH command execution helper for deploy workflows.
// - Hide session lifecycle management behind one small utility function.
// - Return merged stdout/stderr output for deployment logging and diagnostics.
//
// 适用场景：
// - 为部署流程提供轻量级 SSH 命令执行 helper。
// - 把 session 生命周期管理隐藏在一个小工具函数后面。
// - 返回合并后的 stdout/stderr 输出，方便部署日志和排障。
package deploy

import (
	"golang.org/x/crypto/ssh"
)

// RunRemote runs one shell command on the remote host and returns the merged output.
//
// RunRemote 在远端机器上执行一条 shell 命令，并返回合并输出。
//
// 中文说明：
// - 每次调用都会新建一个 ssh session，用完即关。
// - 返回结果为 stdout/stderr 合并输出，便于上层直接记录部署日志。
func RunRemote(client *ssh.Client, cmd string) (string, error) {
	sess, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	out, err := sess.CombinedOutput(cmd)
	return string(out), err
}
