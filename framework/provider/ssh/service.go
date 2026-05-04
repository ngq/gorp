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

func NewService(c runtimecontract.Container) (*Service, error) {
	return &Service{c: c, clients: map[string]*clientHandle{}}, nil
}

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
