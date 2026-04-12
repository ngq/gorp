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
