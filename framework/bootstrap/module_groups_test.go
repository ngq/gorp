// Package bootstrap_test provides unit tests for foundation providers and module group stability.
//
// 适用场景：
// - 验证 FoundationProviders 返回的 6 个基础 provider 列表稳定。
// - 确保 module groups 定义在运行时不被意外修改。
package bootstrap

import (
	"testing"

	"github.com/ngq/gorp/framework/provider/app"
	"github.com/ngq/gorp/framework/provider/config"
	"github.com/ngq/gorp/framework/provider/cron"
	"github.com/ngq/gorp/framework/provider/gin"
	"github.com/ngq/gorp/framework/provider/host"
	"github.com/ngq/gorp/framework/provider/log"
)

func TestFoundationProvidersRemainDefaultRuntimeSkeleton(t *testing.T) {
	providers := FoundationProviders()
	if len(providers) != 6 {
		t.Fatalf("foundation providers len=%d, want 6", len(providers))
	}

	want := []string{
		app.NewProvider().Name(),
		config.NewProvider().Name(),
		log.NewProvider().Name(),
		gin.NewProvider().Name(),
		host.NewProvider().Name(),
		cron.NewProvider().Name(),
	}
	for i, p := range providers {
		if p.Name() != want[i] {
			t.Fatalf("foundation provider[%d]=%s, want %s", i, p.Name(), want[i])
		}
	}
}

func TestCoreProvidersCurrentlyAliasFoundationProviders(t *testing.T) {
	core := CoreProviders()
	foundation := FoundationProviders()
	if len(core) != len(foundation) {
		t.Fatalf("core providers len=%d, want %d", len(core), len(foundation))
	}
	for i := range foundation {
		if core[i].Name() != foundation[i].Name() {
			t.Fatalf("core provider[%d]=%s, want foundation %s", i, core[i].Name(), foundation[i].Name())
		}
	}
}

func TestDefaultProvidersStartFromFoundationProviders(t *testing.T) {
	providers := DefaultProviders()
	foundation := FoundationProviders()
	if len(providers) < len(foundation) {
		t.Fatalf("default providers len=%d, want at least %d", len(providers), len(foundation))
	}
	for i := range foundation {
		if providers[i].Name() != foundation[i].Name() {
			t.Fatalf("default provider[%d]=%s, want foundation %s", i, providers[i].Name(), foundation[i].Name())
		}
	}
}
