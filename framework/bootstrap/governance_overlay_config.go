// Application scenarios:
// - Overlay governance disable/provider overrides on top of the shared config contract.
// - Let startup options reuse the same selector path as file-based config overrides.
// - Keep config-driven and code-driven governance overrides on one effective view.
//
// 适用场景：
// - 在共享配置契约之上叠加治理关闭项和 provider 覆盖项。
// - 让启动 option 复用与配置文件相同的 selector 逻辑路径。
// - 将配置驱动与代码驱动的治理覆盖统一到同一个生效视图上。
package bootstrap

import (
	"context"

	datacontract "github.com/ngq/gorp/framework/contract/data"
)

type governanceOverlayConfig struct {
	base              datacontract.Config
	governanceDisable []string
	governanceProviders map[string]string
}

func overlayGovernanceConfig(base datacontract.Config, disabled []string, providers map[string]string) datacontract.Config {
	if len(disabled) == 0 && len(providers) == 0 {
		return base
	}
	return &governanceOverlayConfig{
		base:                base,
		governanceDisable:   append([]string(nil), disabled...),
		governanceProviders: cloneGovernanceProviderMap(providers),
	}
}

func (c *governanceOverlayConfig) Env() string {
	if c.base == nil {
		return ""
	}
	return c.base.Env()
}

func (c *governanceOverlayConfig) Get(key string) any {
	switch key {
	case "governance.disable":
		if len(c.governanceDisable) > 0 {
			return append([]string(nil), c.governanceDisable...)
		}
	}
	if value, ok := c.lookupGovernanceProviderKey(key); ok {
		return value
	}
	if c.base == nil {
		return nil
	}
	return c.base.Get(key)
}

func (c *governanceOverlayConfig) GetString(key string) string {
	if value, ok := c.lookupGovernanceProviderKey(key); ok {
		return value
	}
	if c.base == nil {
		return ""
	}
	return c.base.GetString(key)
}

func (c *governanceOverlayConfig) GetInt(key string) int {
	if c.base == nil {
		return 0
	}
	return c.base.GetInt(key)
}

func (c *governanceOverlayConfig) GetBool(key string) bool {
	if c.base == nil {
		return false
	}
	return c.base.GetBool(key)
}

func (c *governanceOverlayConfig) GetFloat(key string) float64 {
	if c.base == nil {
		return 0
	}
	return c.base.GetFloat(key)
}

func (c *governanceOverlayConfig) Unmarshal(key string, out any) error {
	if c.base == nil {
		return nil
	}
	return c.base.Unmarshal(key, out)
}

func (c *governanceOverlayConfig) Watch(ctx context.Context, key string) (datacontract.ConfigWatcher, error) {
	if c.base == nil {
		return nil, nil
	}
	return c.base.Watch(ctx, key)
}

func (c *governanceOverlayConfig) Reload(ctx context.Context) error {
	if c.base == nil {
		return nil
	}
	return c.base.Reload(ctx)
}

func (c *governanceOverlayConfig) lookupGovernanceProviderKey(key string) (string, bool) {
	const prefix = "governance.providers."
	if len(c.governanceProviders) == 0 || len(key) <= len(prefix) || key[:len(prefix)] != prefix {
		return "", false
	}
	value, ok := c.governanceProviders[key[len(prefix):]]
	return value, ok
}

func cloneGovernanceProviderMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(src))
	for key, value := range src {
		cloned[key] = value
	}
	return cloned
}
