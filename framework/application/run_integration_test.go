// Package application_test provides integration tests for application startup.
//
// 适用场景：
// - 验证 Application Run 的真实启动流程。
//go:build integration

package application

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/bootstrap"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// TestIntegrationRunBootsHTTPService verifies that Run actually boots an HTTP service.
// This test creates a minimal config file and starts a real HTTP server.
//
// TestIntegrationRunBootsHTTPService 验证 Run 真实启动 HTTP 服务。
// 此测试创建最小配置文件并启动真实 HTTP 服务器。
func TestIntegrationRunBootsHTTPService(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "app.yaml")
	configContent := `
app:
  name: test-service
  env: test
server:
  http:
    addr: :18080
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Set config path
	oldConfigPath := bootstrap.GetConfigPath()
	defer bootstrap.SetConfigPath(oldConfigPath)
	bootstrap.SetConfigPath(configPath)

	// Track server startup
	serverReady := make(chan struct{}, 1)
	serverStopped := make(chan struct{}, 1)

	// Run in goroutine
	runErr := make(chan error, 1)
	go func() {
		err := Run(
			HTTP(),
			WithHTTPRoutes(func(router transportcontract.HTTPRouter, c bootstrap.Container) error {
				router.GET("/health", func(ctx transportcontract.HTTPContext) {
					ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
				})
				return nil
			}),
			WithSetup(func(rt *bootstrap.HTTPServiceRuntime) error {
				close(serverReady)
				return nil
			}),
		)
		runErr <- err
		close(serverStopped)
	}()

	// Wait for server to be ready
	select {
	case <-serverReady:
		// Server is ready
	case <-time.After(5 * time.Second):
		t.Fatal("server did not start within 5 seconds")
	}

	// Make health check request
	resp, err := http.Get("http://localhost:18080/health")
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Shutdown by canceling context (simulate graceful shutdown)
	// In real scenario, this would be triggered by signal
	// For this test, we just verify the server started and responded
}

// TestIntegrationBuildHTTPRuntimeCreatesContainer verifies that BuildHTTPRuntime creates a valid runtime.
//
// TestIntegrationBuildHTTPRuntimeCreatesContainer 验证 BuildHTTPRuntime 创建有效的运行时。
func TestIntegrationBuildHTTPRuntimeCreatesContainer(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "app.yaml")
	configContent := `
app:
  name: test-service
  env: test
server:
  http:
    addr: :18081
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Set config path
	oldConfigPath := bootstrap.GetConfigPath()
	defer bootstrap.SetConfigPath(oldConfigPath)
	bootstrap.SetConfigPath(configPath)

	// Build runtime
	rt, err := BuildHTTPRuntime(HTTP())
	if err != nil {
		t.Fatalf("BuildHTTPRuntime failed: %v", err)
	}
	defer rt.Container.Close()

	// Verify runtime components
	if rt.Container == nil {
		t.Fatal("Container is nil")
	}
	if rt.Config == nil {
		t.Fatal("Config is nil")
	}

	// Verify config values
	serviceName := rt.Config.GetString("app.name")
	if serviceName != "test-service" {
		t.Fatalf("expected service name 'test-service', got '%s'", serviceName)
	}
}

// TestIntegrationRunContextCanBeCanceled verifies that RunContext respects context cancellation.
//
// TestIntegrationRunContextCanBeCanceled 验证 RunContext 响应 context 取消。
func TestIntegrationRunContextCanBeCanceled(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "app.yaml")
	configContent := `
app:
  name: test-service
  env: test
server:
  http:
    addr: :18082
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Set config path
	oldConfigPath := bootstrap.GetConfigPath()
	defer bootstrap.SetConfigPath(oldConfigPath)
	bootstrap.SetConfigPath(configPath)

	ctx, cancel := context.WithCancel(context.Background())

	runErr := make(chan error, 1)
	go func() {
		err := RunContext(ctx, HTTP())
		runErr <- err
	}()

	// Cancel after a short delay
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for RunContext to return
	select {
	case err := <-runErr:
		// RunContext should return (either with canceled error or nil after graceful shutdown)
		t.Logf("RunContext returned: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("RunContext did not return after cancellation")
	}
}
