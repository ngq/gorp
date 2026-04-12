package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// 配置结构对齐 config/app*.yaml：
// ssh:
//   timeout_sec: 10
//   hosts:
//     web-01:
//       host: 127.0.0.1
//       port: 22
//       username: root
//       password: ""            # 可选：密码认证
//       key_path: "~/.ssh/id_rsa"  # 可选：key 认证
//       known_hosts: "~/.ssh/known_hosts"

// hostConfig 描述单台远程主机的 SSH 连接参数。
//
// 中文说明：
// - Host/Port/Username 是建立 SSH 连接的基础三元组。
// - Password 与 KeyPath 二选一即可，两者都存在时会同时加入认证方法列表。
// - KnownHosts 用于开启主机指纹校验；不配置时会退化为忽略校验。
// - 该结构直接映射配置文件中的 ssh.hosts.<name> 节点。
type hostConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	KeyPath    string `mapstructure:"key_path"`
	KnownHosts string `mapstructure:"known_hosts"`
}

// sshConfig 是 ssh 顶层配置结构。
//
// 中文说明：
// - TimeoutSec 控制整体拨号超时时间。
// - Hosts 按名称维护多个远程主机，deploy 等命令通过名称获取连接。
// - 这里不缓存配置内容，而是每次获取客户端时重新从配置中心读取，便于热更新后生效。
type sshConfig struct {
	TimeoutSec int                   `mapstructure:"timeout_sec"`
	Hosts      map[string]hostConfig `mapstructure:"hosts"`
}

// Service 提供 SSH 客户端的延迟创建与连接复用能力。
//
// 中文说明：
// - c 用于从容器中获取 config 等依赖。
// - clients 会缓存已经建立好的 SSH 连接，避免同一主机重复拨号。
// - mu 保护 clients，避免并发场景下出现 map 竞争或重复建连。
type Service struct {
	c contract.Container

	mu      sync.Mutex
	clients map[string]*ssh.Client
}

// NewService 创建 SSH 服务实例。
//
// 中文说明：
// - 这里只初始化缓存 map，不会主动建立任何网络连接。
// - 真正的拨号动作发生在首次调用 GetClient 时。
func NewService(c contract.Container) (*Service, error) {
	return &Service{c: c, clients: map[string]*ssh.Client{}}, nil
}

// GetClient 按主机名获取一个可复用的 SSH 客户端。
//
// 核心流程：
// 1. 先检查 hostName 是否为空，避免后续查配置时出现歧义。
// 2. 优先从本地缓存读取，命中后直接返回已建立连接。
// 3. 未命中时，从配置中心读取 ssh 配置并解析目标主机。
// 4. 调用 dial 建立新连接。
// 5. 建连成功后再次加锁做竞态兜底，确保只缓存一个连接实例。
//
// 注意：
// - 这里的缓存没有做健康检查；如果远端断开，后续使用方在 session/exec 时可能才会感知错误。
// - 若未来要增强健壮性，可以在这里补充 ping/重连策略，但当前实现保持最小复杂度。
func (s *Service) GetClient(hostName string) (*ssh.Client, error) {
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

	cfgAny, err := s.c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg := cfgAny.(contract.Config)

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

	client, err := dial(h, timeout)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	// 竞态兜底：如果另一个 goroutine 已经建立连接，就关闭新建连接复用旧的。
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
	// expand 用于把 ~/xxx 形式的路径展开为用户家目录绝对路径。
	// 这样配置文件里既能保持可读性，又不会要求调用方自己做路径标准化。
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

	// auth 会按顺序收集所有可用认证方式。
	// 当前支持：
	// - 密码认证
	// - 私钥认证
	// ssh.ClientConfig 会依次尝试这些方法，直到成功或全部失败。
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

	// Host key verification
	//
	// 中文说明：
	// - 默认使用 InsecureIgnoreHostKey，意味着只要能连上就接受对端主机指纹。
	// - 如果配置了 known_hosts，则切换为标准 known_hosts 校验，安全性更高。
	// - deploy 到生产环境时，建议务必配置 known_hosts，避免中间人攻击风险。
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

	// 这里使用 tcp + host:port 发起底层连接，ssh.Dial 内部会完成握手与认证。
	// 如果 host 不可达、认证失败、host key 校验失败，错误都会在这里向上返回。
	addr := net.JoinHostPort(h.Host, fmt.Sprintf("%d", h.Port))
	return ssh.Dial("tcp", addr, clientCfg)
}
