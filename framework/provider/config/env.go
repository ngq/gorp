// Package config provides environment normalization utilities for gorp framework.
// Standardizes environment names: dev/test/prod, with backward compatibility for legacy aliases.
//
// 配置包提供 gorp 框架的环境名规范化工具。
// 统一环境名约定：dev/test/prod，并兼容历史别名 development/testing/production。
package config

import "strings"

const (
	EnvDev        = "dev"
	EnvTest       = "test"
	EnvProd       = "prod"
	LegacyDev     = "development"
	LegacyTest    = "testing"
	LegacyProd    = "production"
)

// NormalizeEnv normalizes environment name to framework standard.
// Converts legacy aliases to short names: development->dev, testing->test, production->prod.
// Core logic: Trim whitespace, lowercase, map to canonical name.
//
// NormalizeEnv 将环境名规范化为框架统一标准。
// 将历史别名转换为短名：development->dev, testing->test, production->prod。
// 核心逻辑：去除空白、小写化、映射到标准名称。
func NormalizeEnv(env string) string {
	env = strings.TrimSpace(strings.ToLower(env))
	switch env {
	case "", EnvDev, LegacyDev:
		return EnvDev
	case EnvTest, LegacyTest:
		return EnvTest
	case EnvProd, LegacyProd:
		return EnvProd
	default:
		return env
	}
}
