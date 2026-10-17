// Application scenarios:
// - Safely locate configuration files, data models, or assets regardless of the current working directory.
// - Eliminate brittle relative-path assumptions when running binaries from subdirectories, containers, or test runners.
// - Search across APP_BASE_PATH, executable directories, and parent directories.
//
// 适用场景：
// - 无论当前工作目录如何，安全定位配置文件、数据模型或静态资源。
// - 消除在子目录、容器或测试运行器中执行二进制时的脆弱相对路径假设。
// - 跨 APP_BASE_PATH、可执行文件目录以及父级目录进行智能探测。
package support

import (
	"os"
	"path/filepath"
	"strings"
)

// ResolveResourcePath safely searches for a file path across standard runtime locations:
// 1. As-is if it exists (absolute or relative to current working directory).
// 2. Relative to APP_BASE_PATH environment variable (if set).
// 3. Relative to the running executable's directory.
// 4. Relative to caller-provided extra search directories.
// 5. Searching up to 3 parent directories (../, ../../, ../../../).
//
// Returns the resolved absolute path if found, or empty string if not found.
func ResolveResourcePath(filePath string, extraSearchDirs ...string) string {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return ""
	}

	// 1. Check direct path (absolute or relative to current working directory)
	if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
		if abs, err := filepath.Abs(filePath); err == nil {
			return abs
		}
		return filepath.Clean(filePath)
	}

	// 2. Check APP_BASE_PATH
	if base := strings.TrimSpace(os.Getenv("APP_BASE_PATH")); base != "" {
		candidate := filepath.Join(base, filePath)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			if abs, err := filepath.Abs(candidate); err == nil {
				return abs
			}
			return filepath.Clean(candidate)
		}
	}

	// 3. Check extra search directories
	for _, dir := range extraSearchDirs {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		candidate := filepath.Join(dir, filePath)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			if abs, err := filepath.Abs(candidate); err == nil {
				return abs
			}
			return filepath.Clean(candidate)
		}
	}

	// 4. Check executable directory
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		candidate := filepath.Join(execDir, filePath)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			if abs, err := filepath.Abs(candidate); err == nil {
				return abs
			}
			return filepath.Clean(candidate)
		}
		// Also check parent of executable directory (e.g. bin/../config)
		candidateParent := filepath.Join(filepath.Dir(execDir), filePath)
		if info, err := os.Stat(candidateParent); err == nil && !info.IsDir() {
			if abs, err := filepath.Abs(candidateParent); err == nil {
				return abs
			}
			return filepath.Clean(candidateParent)
		}
	}

	// 5. Check up to 3 parent directories of current working directory
	wd, err := os.Getwd()
	if err == nil {
		curr := wd
		for i := 0; i < 3; i++ {
			parent := filepath.Dir(curr)
			if parent == curr {
				break
			}
			candidate := filepath.Join(parent, filePath)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				if abs, err := filepath.Abs(candidate); err == nil {
					return abs
				}
				return filepath.Clean(candidate)
			}
			curr = parent
		}
	}

	return ""
}
