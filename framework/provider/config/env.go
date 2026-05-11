// Package config provides environment normalization utilities for gorp framework.
// Standardizes environment names: dev/test/prod.
//
// 配置包提供 gorp 框架的环境名规范化工具。
// 统一环境名约定：dev/test/prod。
package config

import "strings"

const (
	EnvDev  = "dev"
	EnvTest = "test"
	EnvProd = "prod"
)

// NormalizeEnv normalizes environment name to framework standard.
// Accepts dev/test/prod and empty string (defaults to dev).
// Core logic: Trim whitespace, lowercase, map empty to dev.
//
// NormalizeEnv 将环境名规范化为框架统一标准。
// 接受 dev/test/prod 和空字符串（默认为 dev）。
// 核心逻辑：去除空白、小写化、空值映射到 dev。
func NormalizeEnv(env string) string {
	env = strings.TrimSpace(strings.ToLower(env))
	switch env {
	case "", EnvDev:
		return EnvDev
	case EnvTest:
		return EnvTest
	case EnvProd:
		return EnvProd
	default:
		return env
	}
}
