package config

import (
	"os"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 把配置服务注册进容器，并在 Boot 阶段完成加载。
//
// 中文说明：
// - cfg 持有一个稳定的 *Service 指针，保证 Register 和 Boot 操作的是同一个实例。
// - 这样容器里拿到的 config 服务在 Boot 后就已经带有完整配置内容，无需再次替换对象。
type Provider struct {
	cfg *Service
}

func NewProvider() *Provider {
	return &Provider{cfg: NewService()}
}

// Name 返回 provider 名称。
func (p *Provider) Name() string { return "config" }

// IsDefer 表示 config provider 不走延迟加载。
func (p *Provider) IsDefer() bool { return false }

// Provides 返回 config provider 暴露的能力 key。
func (p *Provider) Provides() []string { return []string{contract.ConfigKey} }

// Register 绑定稳定的 config service 指针。
func (p *Provider) Register(c contract.Container) error {
	// Bind a stable pointer so Boot() can load into it.
	cfg := p.cfg
	c.Bind(contract.ConfigKey, func(contract.Container) (any, error) {
		return cfg, nil
	}, true)
	return nil
}

// Boot 根据 APP_ENV 装载配置内容。
func (p *Provider) Boot(contract.Container) error {
	env := NormalizeEnv(os.Getenv("APP_ENV"))
	// 中文说明：
	// - APP_ENV 是整个配置装载流程的入口变量；
	// - framework 级统一约定使用 dev / test / prod；
	// - 同时兼容 development / testing / production 的历史值。
	return p.cfg.Load(env)
}
