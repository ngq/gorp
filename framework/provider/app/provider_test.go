package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
)

type stubConfig struct {
	values map[string]string
}

func (s *stubConfig) Env() string                  { return "testing" }
func (s *stubConfig) Get(key string) any           { return s.values[key] }
func (s *stubConfig) GetString(key string) string  { return s.values[key] }
func (s *stubConfig) GetInt(string) int            { return 0 }
func (s *stubConfig) GetBool(string) bool          { return false }
func (s *stubConfig) GetFloat(string) float64      { return 0 }
func (s *stubConfig) Unmarshal(string, any) error  { return nil }
func (s *stubConfig) Watch(context.Context, string) (contract.ConfigWatcher, error) {
	return nil, nil
}
func (s *stubConfig) Reload(context.Context) error { return nil }

func TestProvider_DefaultPathsFollowWorkingDir(t *testing.T) {
	c := container.New()
	if err := c.RegisterProvider(NewProvider()); err != nil {
		t.Fatal(err)
	}
	v, err := c.Make(AppKey)
	if err != nil {
		t.Fatal(err)
	}
	svc := v.(App)
	base := svc.BasePath()
	if svc.StoragePath() != filepath.Join(base, "storage") {
		t.Fatalf("unexpected storage path: %s", svc.StoragePath())
	}
	if svc.RuntimePath() != filepath.Join(base, "storage", "runtime") {
		t.Fatalf("unexpected runtime path: %s", svc.RuntimePath())
	}
	if svc.LogPath() != filepath.Join(base, "storage", "log") {
		t.Fatalf("unexpected log path: %s", svc.LogPath())
	}
}

func TestProvider_ConfigurablePaths(t *testing.T) {
	c := container.New()
	c.Bind(contract.ConfigKey, func(contract.Container) (any, error) {
		return &stubConfig{values: map[string]string{
			"app.paths.base":    "custom-root",
			"app.paths.storage": "var",
			"app.paths.runtime": "run",
			"app.paths.log":     "logs",
		}}, nil
	}, true)
	if err := c.RegisterProvider(NewProvider()); err != nil {
		t.Fatal(err)
	}
	v, err := c.Make(AppKey)
	if err != nil {
		t.Fatal(err)
	}
	svc := v.(App)
	if filepath.Base(svc.BasePath()) != "custom-root" {
		t.Fatalf("unexpected base path: %s", svc.BasePath())
	}
	if svc.StoragePath() != filepath.Join(svc.BasePath(), "var") {
		t.Fatalf("unexpected storage path: %s", svc.StoragePath())
	}
	if svc.RuntimePath() != filepath.Join(svc.StoragePath(), "run") {
		t.Fatalf("unexpected runtime path: %s", svc.RuntimePath())
	}
	if svc.LogPath() != filepath.Join(svc.StoragePath(), "logs") {
		t.Fatalf("unexpected log path: %s", svc.LogPath())
	}
}
