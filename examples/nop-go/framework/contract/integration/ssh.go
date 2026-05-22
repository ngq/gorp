// Application scenarios:
// - Define the SSH integration contract used by deployment, ops, or remote execution features.
// - Keep SSH client/session access provider-neutral while allowing native-client escape hatches.
// - Standardize host-level SSH client lookup and command execution semantics.
//
// 适用场景：
// - 定义部署、运维或远程执行功能使用的 SSH 集成契约。
// - 保持 SSH client/session 访问与具体 provider 解耦，同时保留 native client 下探能力。
// - 统一按主机名获取 SSH client 与执行命令的语义。
package integration

import "golang.org/x/crypto/ssh"

// SSHKey is the container key for the SSH capability.
//
// SSHKey 是 SSH 能力的容器键。
const SSHKey = "framework.ssh"

// SSHService defines the SSH capability exposed by the framework.
//
// SSHService 定义框架对外暴露的 SSH 能力。
type SSHService interface {
	Client(hostName string) (SSHClient, error)
}

// SSHClient defines the SSH client contract.
//
// SSHClient 定义 SSH 客户端契约。
type SSHClient interface {
	NewSession() (SSHSession, error)
	Close() error
}

// SSHSession defines the SSH session contract.
//
// SSHSession 定义 SSH 会话契约。
type SSHSession interface {
	CombinedOutput(cmd string) ([]byte, error)
	Run(cmd string) error
	Close() error
}

// NativeSSHClient exposes the underlying native ssh.Client when available.
//
// NativeSSHClient 在可用时暴露底层原生 ssh.Client。
type NativeSSHClient interface {
	NativeSSHClient() *ssh.Client
}
