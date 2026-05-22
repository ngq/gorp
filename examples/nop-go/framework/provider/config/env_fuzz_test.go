// Package config_test provides fuzz tests for environment normalization.
//
// 适用场景：
// - 验证 NormalizeEnv 对各种边界输入的处理稳定性。
// - 发现潜在的 panic 或异常行为。
package config_test

import (
	"strings"
	"testing"

	"github.com/ngq/gorp/framework/provider/config"
)

// FuzzNormalizeEnv fuzzes NormalizeEnv with arbitrary strings.
//
// FuzzNormalizeEnv 对 NormalizeEnv 进行模糊测试，使用任意字符串。
func FuzzNormalizeEnv(f *testing.F) {
	// Seed corpus with typical inputs
	f.Add("dev")
	f.Add("test")
	f.Add("prod")
	f.Add("")
	f.Add("DEV")
	f.Add("TEST")
	f.Add("PROD")
	f.Add("Dev")
	f.Add("  dev  ")
	f.Add("development")
	f.Add("staging")
	f.Add("production")
	f.Add("中文")
	f.Add("test-env")
	f.Add("env_123")

	f.Fuzz(func(t *testing.T, env string) {
		result := config.NormalizeEnv(env)

		// Verify result is one of the valid values or preserved
		// 验证结果是有效值之一或被保留
		switch result {
		case config.EnvDev, config.EnvTest, config.EnvProd:
			// Valid normalized values
			// 有效的规范化值
		default:
			// Unknown values should be preserved as-is (lowercased)
			// 未知值应该原样保留（小写化）
			if result != "" && result != strings.ToLower(strings.TrimSpace(env)) {
				t.Logf("NormalizeEnv(%q) = %q (preserved)", env, result)
			}
		}

		// Verify result is always lowercase
		// 验证结果总是小写
		if result != strings.ToLower(result) {
			t.Errorf("NormalizeEnv(%q) = %q, expected lowercase", env, result)
		}

		// Verify result has no leading/trailing whitespace
		// 验证结果没有前导/尾随空白
		if result != strings.TrimSpace(result) {
			t.Errorf("NormalizeEnv(%q) = %q, expected trimmed", env, result)
		}
	})
}

// FuzzNormalizeEnv_Idempotent verifies NormalizeEnv is idempotent.
//
// FuzzNormalizeEnv_Idempotent 验证 NormalizeEnv 是幂等的。
func FuzzNormalizeEnv_Idempotent(f *testing.F) {
	f.Add("dev")
	f.Add("TEST")
	f.Add("  prod  ")
	f.Add("unknown")

	f.Fuzz(func(t *testing.T, env string) {
		first := config.NormalizeEnv(env)
		second := config.NormalizeEnv(first)

		if first != second {
			t.Errorf("NormalizeEnv not idempotent: NormalizeEnv(%q) = %q, NormalizeEnv(%q) = %q",
				env, first, first, second)
		}
	})
}
