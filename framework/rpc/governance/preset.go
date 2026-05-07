package governance

import "time"

// DefaultClientPresetOptions controls the default outbound RPC governance preset.
//
// DefaultClientPresetOptions 用于控制默认出站 RPC 治理预设。
type DefaultClientPresetOptions struct {
	Timeout time.Duration
}

// DefaultClientPresetOrder returns the stable logical order of the outbound governance chain.
//
// DefaultClientPresetOrder 返回出站治理链的稳定逻辑顺序。
func DefaultClientPresetOrder() []string {
	return []string{
		"selector",
		"timeout",
		"tracing",
		"metadata",
		"serviceauth",
		"breaker",
		"retry",
	}
}
