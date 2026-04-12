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
