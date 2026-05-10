// Package ssh provides SSH service implementation.
// Manages SSH connections with configurable authentication.
// Supports password and key-based authentication with known_hosts verification.
//
// SSH 包提供 SSH 服务实现。
// 管理带可配置认证的 SSH 连接。
// 支持密码和密钥认证，带 known_hosts 验证。
package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type hostConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	KeyPath    string `mapstructure:"key_path"`
	KnownHosts string `mapstructure:"known_hosts"`
}

type sshConfig struct {
	TimeoutSec int                   `mapstructure:"timeout_sec"`
	Hosts      map[string]hostConfig `mapstructure:"hosts"`
}

// Service manages SSH connections to multiple hosts.
// Core logic: Cache connections, dial on demand, handle authentication.
//
// Service 管理到多主机的 SSH 连接。
// 核心逻辑：缓存连接、按需拨号、处理认证。
type Service struct {
	c runtimecontract.Container

	mu      sync.Mutex
	clients map[string]*clientHandle
}

type clientHandle struct {
	raw *ssh.Client
}

func (c *clientHandle) NewSession() (integrationcontract.SSHSession, error) {
	if c == nil || c.raw == nil {
		return nil, fmt.Errorf("ssh client is nil")
	}
	session, err := c.raw.NewSession()
	if err != nil {
		return nil, err
	}
	return &sessionHandle{raw: session}, nil
}

func (c *clientHandle) Close() error {
	if c == nil || c.raw == nil {
		return nil
	}
	return c.raw.Close()
}

func (c *clientHandle) NativeSSHClient() *ssh.Client {
	if c == nil {
		return nil
	}
	return c.raw
}

type sessionHandle struct {
	raw *ssh.Session
}

func (s *sessionHandle) CombinedOutput(cmd string) ([]byte, error) {
	if s == nil || s.raw == nil {
		return nil, fmt.Errorf("ssh session is nil")
	}
	return s.raw.CombinedOutput(cmd)
}

func (s *sessionHandle) Run(cmd string) error {
	if s == nil || s.raw == nil {
		return fmt.Errorf("ssh session is nil")
	}
	return s.raw.Run(cmd)
}

func (s *sessionHandle) Close() error {
	if s == nil || s.raw == nil {
		return nil
	}
	return s.raw.Close()
}

// NewService creates a new SSH service with container reference.
// Core logic: Initialize empty client cache.
//
// NewService 创建新的 SSH 服务，携带容器引用。
// 核心逻辑：初始化空的客户端缓存。
func NewService(c runtimecontract.Container) (*Service, error) {
	return &Service{c: c, clients: map[string]*clientHandle{}}, nil
}

// Client returns SSH client for specified host name.
// Core logic: Check cache, dial if not cached, store in cache.
//
// Client 返回指定主机名的 SSH 客户端。
// 核心逻辑：检查缓存、若未缓存则拨号、存入缓存。
func (s *Service) Client(hostName string) (integrationcontract.SSHClient, error) {
	hostName = strings.TrimSpace(hostName)
	if hostName == "" {
		return nil, fmt.Errorf("ssh hostName is required")
	}

	s.mu.Lock()
	if cli, ok := s.clients[hostName]; ok {
		s.mu.Unlock()
		return cli, nil
	}
	s.mu.Unlock()

	cfgAny, err := s.c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg := cfgAny.(datacontract.Config)

	var sc sshConfig
	if err := cfg.Unmarshal("ssh", &sc); err != nil {
		return nil, err
	}
	h, ok := sc.Hosts[hostName]
	if !ok {
		return nil, fmt.Errorf("ssh.hosts.%s not found", hostName)
	}
	if h.Port == 0 {
		h.Port = 22
	}

	timeout := 10 * time.Second
	if sc.TimeoutSec > 0 {
		timeout = time.Duration(sc.TimeoutSec) * time.Second
	}

	rawClient, err := dial(h, timeout)
	if err != nil {
		return nil, err
	}
	client := &clientHandle{raw: rawClient}

	s.mu.Lock()
	if old, ok := s.clients[hostName]; ok {
		s.mu.Unlock()
		_ = client.Close()
		return old, nil
	}
	s.clients[hostName] = client
	s.mu.Unlock()

	return client, nil
}

// dial establishes SSH connection with host configuration.
// Core logic: Expand paths, configure authentication, establish connection.
//
// dial 使用主机配置建立 SSH 连接。
// 核心逻辑：扩展路径、配置认证、建立连接。
func dial(h hostConfig, timeout time.Duration) (*ssh.Client, error) {
	expand := func(p string) string {
		p = strings.TrimSpace(p)
		if p == "" {
			return ""
		}
		if strings.HasPrefix(p, "~/") {
			home, _ := os.UserHomeDir()
			if home != "" {
				return filepath.Join(home, p[2:])
			}
		}
		return p
	}

	h.KeyPath = expand(h.KeyPath)
	h.KnownHosts = expand(h.KnownHosts)

	var auth []ssh.AuthMethod
	if strings.TrimSpace(h.Password) != "" {
		auth = append(auth, ssh.Password(h.Password))
	}
	if h.KeyPath != "" {
		keyBytes, err := os.ReadFile(h.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("read key_path: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}
	if len(auth) == 0 {
		return nil, fmt.Errorf("ssh auth not configured: need password or key_path")
	}

	cb := ssh.InsecureIgnoreHostKey()
	if h.KnownHosts != "" {
		hcb, err := knownhosts.New(h.KnownHosts)
		if err != nil {
			return nil, fmt.Errorf("known_hosts: %w", err)
		}
		cb = hcb
	}

	clientCfg := &ssh.ClientConfig{
		User:            h.Username,
		Auth:            auth,
		HostKeyCallback: cb,
		Timeout:         timeout,
	}

	addr := net.JoinHostPort(h.Host, fmt.Sprintf("%d", h.Port))
	return ssh.Dial("tcp", addr, clientCfg)
}
