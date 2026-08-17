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
// 注意：进程级副作用，同一进程内并行测试会互相影响；尽量用
// ChdirRepoRootRestore 并在 cleanup 中恢复。
func ChdirRepoRoot() error {
	root, err := RepoRoot()
	if err != nil {
		return err
	}
	return os.Chdir(root)
}

// RepoRoot 返回 go.mod 所在的仓库根目录绝对路径。
func RepoRoot() (string, error) {
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		return "", nil
	}
	// here = .../framework/testing/root.go
	return filepath.Dir(filepath.Dir(filepath.Dir(here))), nil
}

// ChdirRepoRootRestore 切到仓库根目录并返回恢复函数（切回原工作目录）。
// 避免 ChdirRepoRoot 的进程级副作用污染同进程内的其他测试。
func ChdirRepoRootRestore() (func(), error) {
	prev, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if err := ChdirRepoRoot(); err != nil {
		return nil, err
	}
	return func() { _ = os.Chdir(prev) }, nil
}
