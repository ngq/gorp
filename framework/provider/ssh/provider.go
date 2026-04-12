package ssh

import "github.com/ngq/gorp/framework/contract"

// Provider 将 SSHService 注入容器。
//
// 中文说明：
// - SSH 并不是 app 启动必须的服务，多数时候只在 deploy 命令里使用。
// - 因此 IsDefer=true，只有在 Make(contract.SSHKey) 时才真正初始化。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "ssh" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string { return []string{contract.SSHKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.SSHKey, func(c contract.Container) (any, error) {
		return NewService(c)
	}, true)
	return nil
}

func (p *Provider) Boot(contract.Container) error { return nil }
