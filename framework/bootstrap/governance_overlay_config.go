// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file overlays governance disable/provider overrides on shared config contract.
// Unifies config-driven and code-driven governance overrides in one effective view.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件在共享配置契约之上叠加治理关闭项和 provider 覆盖项。
// 将配置驱动与代码驱动的治理覆盖统一到同一个生效视图上。
package bootstrap

import (
	"context"

	datacontract "github.com/ngq/gorp/framework/contract/data"
)

type governanceOverlayConfig struct {
	base                datacontract.Config
	governanceDisable   []string
	governanceEnable    []string
	governanceProviders map[string]string
}

// OverlayGovernanceConfig 在基础配置之上叠加代码级治理覆盖（关闭项、开启项、provider 覆盖项）。
// 将代码驱动和配置驱动的治理覆盖统一到同一个生效视图上。
func OverlayGovernanceConfig(base datacontract.Config, disabled []string, enabled []string, providers map[string]string) datacontract.Config {
	if len(disabled) == 0 && len(enabled) == 0 && len(providers) == 0 {
		return base
	}
	return &governanceOverlayConfig{
		base:                base,
		governanceDisable:   append([]string(nil), disabled...),
		governanceEnable:    append([]string(nil), enabled...),
		governanceProviders: cloneGovernanceProviderMap(providers),
	}
}

func overlayGovernanceConfig(base datacontract.Config, disabled []string, enabled []string, providers map[string]string) datacontract.Config {
	return OverlayGovernanceConfig(base, disabled, enabled, providers)
}

func (c *governanceOverlayConfig) Env() string {
	if c.base == nil {
		return ""
	}
	return c.base.Env()
}

func (c *governanceOverlayConfig) Get(key string) any {
	// Handle overlay keys with explicit logic flow (no implicit fallthrough).
	// This is clearer than switch fallthrough and less error-prone when adding new keys.
	//
	// 使用显式逻辑流处理 overlay key（无隐式 fallthrough）。
	// 比 switch fallthrough 更清晰，新增 key 时不易出错。

	// governance.disable / governance.enable: union of overlay and base lists.
	// Merging (not replacing) keeps registration behavior consistent with the
	// governance summary, which also unions both layers — replacing base here
	// made the summary report a capability as disabled while its provider was
	// still registered.
	//
	// governance.disable / governance.enable：overlay 与 base 取并集。
	// 用合并而非替换，使注册行为与治理摘要（同为并集语义）一致——
	// 此前替换会让摘要报告某能力已禁用，实际容器中仍注册了 provider。
	if key == "governance.disable" {
		return mergeOverlayListWithBase(c.base, key, c.governanceDisable)
	}
	if key == "governance.enable" {
		return mergeOverlayListWithBase(c.base, key, c.governanceEnable)
	}

	// governance.providers.*: check overlay map first.
	if value, ok := c.lookupGovernanceProviderKey(key); ok {
		return value
	}

	// Fallback to base config for all other keys.
	if c.base == nil {
		return nil
	}
	return c.base.Get(key)
}

// mergeOverlayListWithBase 返回 overlay 列表与 base 配置同 key 列表的并集
// （overlay 在前，去重）。
func mergeOverlayListWithBase(base datacontract.Config, key string, overlay []string) []string {
	merged := make([]string, 0, len(overlay))
	seen := make(map[string]struct{}, len(overlay))
	for _, v := range overlay {
		if _, dup := seen[v]; dup {
			continue
		}
		seen[v] = struct{}{}
		merged = append(merged, v)
	}
	if base == nil {
		return merged
	}
	switch values := base.Get(key).(type) {
	case []string:
		for _, v := range values {
			if _, dup := seen[v]; !dup {
				seen[v] = struct{}{}
				merged = append(merged, v)
			}
		}
	case []any:
		for _, raw := range values {
			if v, ok := raw.(string); ok {
				if _, dup := seen[v]; !dup {
					seen[v] = struct{}{}
					merged = append(merged, v)
				}
			}
		}
	}
	return merged
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
