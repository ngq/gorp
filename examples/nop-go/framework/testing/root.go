// Package testing provides testing utilities for gorp framework.
// This file provides working directory helpers for tests.
// Changes working directory to repo root where go.mod lives.
//
// 测试包提供 gorp 框架的测试工具能力。
// 本文件提供工作目录测试 helper。
// 切换工作目录到 go.mod 所在的仓库根目录。
package testing

import (
	"os"
	"path/filepath"
	"runtime"
)

// ChdirRepoRoot changes working directory to the repo root (where go.mod lives).
func ChdirRepoRoot() error {
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		return nil
	}
	// here = .../framework/testing/root.go
	root := filepath.Dir(filepath.Dir(filepath.Dir(here)))
	// root should be repo root
	return os.Chdir(root)
}
