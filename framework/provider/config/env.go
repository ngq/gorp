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

// NormalizeEnv 统一环境名。
//
// 中文说明：
// - framework 级统一约定使用 `dev / test / prod`；
// - 兼容历史值 `development / testing / production`；
// - 这样业务项目不必自己再约定一套环境命名。
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
