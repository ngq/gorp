// Package app_test provides unit tests for app provider binding key and path resolution.
//
// 适用场景：
// - 验证 app provider 的绑定 key 和默认路径解析行为。
// - 防止配置驱动的路径覆盖能力回归。
// - 通过聚焦型单测固化路径推导语义。
package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

func TestProvider_ProvidesRootKey(t *testing.T) {
	p := NewProvider()
	if got := p.Provides(); len(got) != 1 || got[0] != runtimecontract.RootKey {
		t.Fatalf("unexpected provides keys: %v", got)
	}
}

type stubConfig struct {
	values map[string]string
}

func (s *stubConfig) Env() string                 { return "testing" }
func (s *stubConfig) Get(key string) any          { return s.values[key] }
func (s *stubConfig) GetString(key string) string { return s.values[key] }
func (s *stubConfig) GetInt(string) int           { return 0 }
func (s *stubConfig) GetBool(string) bool         { return false }
func (s *stubConfig) GetFloat(string) float64     { return 0 }
func (s *stubConfig) Unmarshal(string, any) error { return nil }
func (s *stubConfig) Watch(context.Context, string) (datacontract.ConfigWatcher, error) {
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
	c.Bind(datacontract.ConfigKey, func(runtimecontract.Container) (any, error) {
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
