// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file centralizes governance-mode detection for bootstrap and startup helpers.
// Keeps monolith and microservice default-provider behavior explicit in one place.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件集中处理 bootstrap 选择与启动 helper 使用的治理模式判断。
// 将 monolith 与 microservice 的默认 provider 语义显式收口到一个位置。
package bootstrap

import (
	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// DetectGovernanceMode resolves the runtime governance mode from config.
//
// DetectGovernanceMode 从配置中解析运行时治理模式。
func DetectGovernanceMode(cfg datacontract.Config) resiliencecontract.GovernanceMode {
	if cfg == nil {
		return resiliencecontract.GovernanceModeMono
	}

	for _, key := range []string{
		"governance.mode",
		"app.mode",
		"runtime.mode",
		"service.mode",
	} {
		switch cfg.GetString(key) {
		case string(resiliencecontract.GovernanceModeMicro):
			return resiliencecontract.GovernanceModeMicro
		case string(resiliencecontract.GovernanceModeMono):
			return resiliencecontract.GovernanceModeMono
		}
	}

	return resiliencecontract.GovernanceModeMono
}

// NormalizeGovernanceMode returns a supported governance mode, defaulting to mono.
//
// NormalizeGovernanceMode 返回受支持的治理模式；未命中时默认回退为 mono。
func NormalizeGovernanceMode(mode resiliencecontract.GovernanceMode) resiliencecontract.GovernanceMode {
	switch mode {
	case resiliencecontract.GovernanceModeMicro:
		return resiliencecontract.GovernanceModeMicro
	default:
		return resiliencecontract.GovernanceModeMono
	}
}

// IsMicroMode reports whether the given mode represents the microservice mainline.
//
// IsMicroMode 返回当前模式是否代表微服务主线。
func IsMicroMode(mode resiliencecontract.GovernanceMode) bool {
	return NormalizeGovernanceMode(mode) == resiliencecontract.GovernanceModeMicro
}

// NormalizeHTTPMode returns a supported HTTP mode, defaulting to contract.
//
// NormalizeHTTPMode 返回受支持的 HTTP 模式；未命中时默认回退为 contract。
func NormalizeHTTPMode(mode resiliencecontract.HTTPMode) resiliencecontract.HTTPMode {
	switch mode {
	case resiliencecontract.HTTPModeGin:
		return resiliencecontract.HTTPModeGin
	case resiliencecontract.HTTPModeContract:
		return resiliencecontract.HTTPModeContract
	default:
		return resiliencecontract.HTTPModeContract
	}
}

// IsGinHTTPMode reports whether the HTTP mode uses native gin.Context.
//
// IsGinHTTPMode 返回 HTTP 模式是否使用原生 gin.Context。
func IsGinHTTPMode(httpMode resiliencecontract.HTTPMode) bool {
	return NormalizeHTTPMode(httpMode) == resiliencecontract.HTTPModeGin
}
