// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file builds the governance defaults table for the view=defaults inspect endpoint.
// Projects mode-aware feature defaults, provider defaults, and HTTP/RPC option defaults
// into one serializable, inspection-friendly snapshot.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件为 view=defaults 诊断视图构建治理默认值表。
// 将按模式生效的 feature 默认值、provider 默认值、HTTP/RPC 选项默认值
// 投影成一份可序列化、可检查的快照。
package bootstrap

import (
	"fmt"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/ngq/gorp/framework/http/middleware"
)

// BuildGovernanceDefaultsTable builds the full governance defaults table for one mode.
// This is a pure function that only reads static defaults — no runtime config involved.
//
// BuildGovernanceDefaultsTable 构建某个治理模式下完整的默认值表。
// 这是一个纯函数，只读取静态默认值，不涉及运行时配置。
func BuildGovernanceDefaultsTable(mode resiliencecontract.GovernanceMode) *resiliencecontract.GovernanceDefaultsTable {
	mode = NormalizeGovernanceMode(mode)

	// 特性默认值：从 DefaultGovernanceFeatureSet 投影
	features := resiliencecontract.DefaultGovernanceFeatureSet(mode)
	featureDefaults := map[string]bool{
		"request_identity": features.RequestIdentity,
		"logging":          features.Logging,
		"recovery":         features.Recovery,
		"timeout":          features.Timeout,
		"metrics":          features.Metrics,
		"metadata":         features.MetadataPropagation,
		"tracing":          features.Tracing,
		"selector":         features.Selector,
		"serviceauth":      features.ServiceAuth,
		"circuitbreaker":   features.CircuitBreaker,
		"retry":            features.Retry,
		"loadshedding":     features.LoadShedding,
		"discovery":        features.Discovery,
	}

	// Provider 默认值：从 DefaultGovernanceProviderDefaults 投影
	providerDefaults := governanceProviderMap(DefaultGovernanceProviderDefaults(mode))

	// HTTP 中间件默认值：从 DefaultHTTPServiceGovernanceDefaults 投影
	httpDefaults := middleware.DefaultHTTPServiceGovernanceDefaults()
	middlewareDefaults := resiliencecontract.MiddlewareDefaults{
		Timeout:           httpDefaults.API.Timeout.String(),
		BodyLimit:         formatBytes(httpDefaults.API.BodyLimitBytes),
		MaxConcurrent:     httpDefaults.MaxConcurrent,
		EnableMetrics:     httpDefaults.API.EnableMetrics,
		EnableCompression: httpDefaults.API.EnableCompression,
	}

	// CORS 默认值（启用 CORS 时生效的默认配置）
	corsDefaults := middleware.DefaultCORSOptions()
	middlewareDefaults.CORS = resiliencecontract.CORSDefaults{
		AllowOrigins:  corsDefaults.AllowOrigins,
		MaxAgeSeconds: corsDefaults.MaxAgeSeconds,
	}

	// 安全头默认值
	securityDefaults := middleware.DefaultSecurityHeadersOptions()
	middlewareDefaults.SecurityHeaders = resiliencecontract.SecurityHeaderDefaults{
		XFrameOptions:       securityDefaults.XFrameOptions,
		XContentTypeOptions: securityDefaults.XContentTypeOptions,
		ReferrerPolicy:      securityDefaults.ReferrerPolicy,
	}

	// 本地化默认值
	localeDefaults := middleware.DefaultLocaleOptions()
	middlewareDefaults.Locale = resiliencecontract.LocaleDefaults{
		Supported: localeDefaults.Supported,
		Default:   localeDefaults.Default,
		QueryKeys: localeDefaults.QueryKeys,
	}

	// RPC 客户端默认值
	rpcClientDefaults := resiliencecontract.RPCClientDefaults{
		Timeout: "0s",
	}

	return &resiliencecontract.GovernanceDefaultsTable{
		Mode:               mode,
		FeatureDefaults:    featureDefaults,
		ProviderDefaults:   providerDefaults,
		MiddlewareDefaults: middlewareDefaults,
		RPCClientDefaults:  rpcClientDefaults,
	}
}

// formatBytes converts a byte count into a human-readable string (e.g. "2MB", "1.5MB").
//
// formatBytes 将字节数转换为人类可读字符串（如 "2MB"、"1.5MB"）。
func formatBytes(bytes int64) string {
	if bytes <= 0 {
		return "0"
	}
	const mb = 1024 * 1024
	if bytes%mb == 0 {
		return fmt.Sprintf("%dMB", bytes/mb)
	}
	return fmt.Sprintf("%.1fMB", float64(bytes)/float64(mb))
}
