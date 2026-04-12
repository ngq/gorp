package deploy

import (
	"golang.org/x/crypto/ssh"
)

// RunRemote 在远端机器上执行一条 shell 命令，并返回合并后的输出。
//
// 中文说明：
// - 每次调用都会新建一个 ssh session，用完即关。
// - 返回的是 stdout/stderr 合并结果，便于上层在部署日志里直接展示。
func RunRemote(client *ssh.Client, cmd string) (string, error) {
	sess, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	out, err := sess.CombinedOutput(cmd)
	return string(out), err
}
