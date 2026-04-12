package contract

import "golang.org/x/crypto/ssh"

const SSHKey = "framework.ssh"

// SSHService 提供 SSH 客户端连接能力。
//
// 中文说明：
// - deploy 等命令需要 SSH 连接远端执行命令/上传文件。
// - 将 SSH 抽象为服务（provider）后，可以：
//   1) 统一读取配置（ssh.hosts.*）
//   2) 统一支持 password/key 两种认证
//   3) 支持连接复用（同一 hostName 多次 GetClient 不重复 Dial）
type SSHService interface {
	// GetClient 获取一个 SSH client。
	// hostName 对应配置：ssh.hosts.<hostName>
	GetClient(hostName string) (*ssh.Client, error)
}
