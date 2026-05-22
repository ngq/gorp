// Package testing provides testing utilities for gorp framework.
// This file provides environment variable helpers for tests.
// Sets env and returns restore func for cleanup.
//
// 测试包提供 gorp 框架的测试工具能力。
// 本文件提供环境变量测试 helper。
// 设置环境变量并返回 restore 函数用于清理。
package testing

import (
	"os"
)

// SetEnv sets env and returns a restore func.
func SetEnv(key, value string) func() {
	old, had := os.LookupEnv(key)
	_ = os.Setenv(key, value)
	return func() {
		if !had {
			_ = os.Unsetenv(key)
			return
		}
		_ = os.Setenv(key, old)
	}
}
