// Application scenarios:
// - Provide the legacy direct SSH dial helper used by older deploy flows.
// - Support key-based SSH connection setup with known_hosts verification.
// - Keep this compatibility helper available while the provider-based SSH capability becomes the main path.
//
// 适用场景：
// - 为旧部署流程提供直接 SSH 连接 helper。
// - 支持基于私钥和 known_hosts 校验的 SSH 建连。
// - 在 provider 化 SSH 能力成为主路径之前，保留这层兼容 helper。
package deploy

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Deprecated: use framework/provider/ssh (contract.SSHKey) instead.
//
// 中文说明：
// - 本文件提供的 DialSSH 是直接 SSH 连接能力。
// - 当前已将 SSH 抽象为 Provider，支持 password/key 两种认证以及连接复用。
// - 新代码应优先使用 framework/provider/ssh；本文件仅用于兼容过渡。

// SSHHost describes one SSH host entry.
//
// SSHHost 描述一条 SSH 主机配置。
type SSHHost struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Username   string `mapstructure:"username"`
	KeyPath    string `mapstructure:"key_path"`
	KnownHosts string `mapstructure:"known_hosts"`
}

// SSHConfig is the legacy SSH config model used by deploy helpers.
//
// SSHConfig 是 deploy 旧路径使用的 SSH 配置结构。
//
// 中文说明：
// - 新代码应优先使用 framework/provider/ssh 中的配置结构和服务。
// - 这里仅用于兼容旧调用路径。
type SSHConfig struct {
	TimeoutSec int                `mapstructure:"timeout_sec"`
	Hosts      map[string]SSHHost `mapstructure:"hosts"`
}

// DialSSH dials one SSH host with key-based authentication.
//
// DialSSH 使用私钥认证连接一个 SSH 主机。
func DialSSH(h SSHHost, timeout time.Duration) (*ssh.Client, error) {
	keyBytes, err := os.ReadFile(h.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("read key_path: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	hostKeyCallback, err := knownhosts.New(h.KnownHosts)
	if err != nil {
		return nil, fmt.Errorf("known_hosts: %w", err)
	}

	cfg := &ssh.ClientConfig{
		User:            h.Username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback,
		Timeout:         timeout,
	}

	addr := net.JoinHostPort(h.Host, fmt.Sprintf("%d", h.Port))
	c, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, err
	}
	return c, nil
}
