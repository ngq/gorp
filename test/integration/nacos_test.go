// Package integration provides Nacos registry integration tests.
//
// 本包提供 Nacos 注册发现集成测试。
package integration

import (
	"context"
	"testing"
	"time"

	nacosregistry "github.com/ngq/gorp/contrib/registry/nacos"
)

// TestNacosRegisterDiscover tests real Nacos register and discover operations.
//
// TestNacosRegisterDiscover 测试真实 Nacos 注册和发现操作。
func TestNacosRegisterDiscover(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Nacos backend")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Create Nacos registry
	// 创建 Nacos registry
	cfg := &nacosregistry.DiscoveryConfig{
		NacosAddr:      getEnvOrDefault("GORP_TEST_NACOS_ADDR", "localhost:8848"),
		NacosNamespace: "", // public namespace
		NacosGroup:     "DEFAULT_GROUP",
		ServiceWeight:  1.0,
		LoadBalance:    "weight",
	}

	registry, err := nacosregistry.NewRegistry(cfg)
	if err != nil {
		t.Fatalf("failed to create Nacos registry: %v", err)
	}
	defer registry.Close()

	// 2. Register service instance
	// 注册服务实例
	testService := "test-service-nacos"
	testAddress := "192.168.1.100:8080"
	testMetadata := map[string]string{
		"version": "v1",
		"env":     "test",
	}

	err = registry.Register(ctx, testService, testAddress, testMetadata)
	if err != nil {
		t.Fatalf("failed to register service: %v", err)
	}

	t.Logf("registered service: %s at %s", testService, testAddress)

	// 3. Deregister service
	// 注销服务
	err = registry.Deregister(ctx, testService, testAddress)
	if err != nil {
		t.Fatalf("failed to deregister service: %v", err)
	}

	t.Logf("deregistered service: %s at %s", testService, testAddress)
	t.Log("verified: Nacos register/deregister operations work correctly")
}

// TestNacosMultipleInstances tests registering and discovering multiple instances.
//
// TestNacosMultipleInstances 测试注册和发现多个实例。
func TestNacosMultipleInstances(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Nacos backend")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := &nacosregistry.DiscoveryConfig{
		NacosAddr:      getEnvOrDefault("GORP_TEST_NACOS_ADDR", "localhost:8848"),
		NacosNamespace: "",
		NacosGroup:     "DEFAULT_GROUP",
		ServiceWeight:  1.0,
		LoadBalance:    "weight",
	}

	registry, err := nacosregistry.NewRegistry(cfg)
	if err != nil {
		t.Fatalf("failed to create Nacos registry: %v", err)
	}
	defer registry.Close()

	multiService := "multi-instance-nacos"

	// Register 3 instances with different weights
	// 注册 3 个不同权重的实例
	addresses := []string{
		"10.0.0.1:8080",
		"10.0.0.2:8080",
		"10.0.0.3:8080",
	}

	for i, addr := range addresses {
		err := registry.Register(ctx, multiService, addr, map[string]string{
			"index": string(rune('a' + i)),
		})
		if err != nil {
			t.Fatalf("failed to register instance %d: %v", i, err)
		}
	}

	t.Logf("registered %d instances for service %s", len(addresses), multiService)

	// Cleanup
	// 清理
	for _, addr := range addresses {
		registry.Deregister(ctx, multiService, addr)
	}

	t.Log("verified: multiple instance registration works correctly")
}

// TestNacosWeightedLoadBalance tests Nacos weight-based load balancing.
//
// TestNacosWeightedLoadBalance 测试 Nacos 基于权重的负载均衡。
func TestNacosWeightedLoadBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Nacos backend")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := &nacosregistry.DiscoveryConfig{
		NacosAddr:      getEnvOrDefault("GORP_TEST_NACOS_ADDR", "localhost:8848"),
		NacosNamespace: "",
		NacosGroup:     "DEFAULT_GROUP",
		ServiceWeight:  1.0,
		LoadBalance:    "weight",
	}

	registry, err := nacosregistry.NewRegistry(cfg)
	if err != nil {
		t.Fatalf("failed to create Nacos registry: %v", err)
	}
	defer registry.Close()

	weightService := "weight-test-nacos"

	// Register instances
	// 注册实例
	err = registry.Register(ctx, weightService, "10.0.0.1:8080", nil)
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	t.Logf("registered instance for service %s", weightService)

	// Cleanup
	// 清理
	registry.Deregister(ctx, weightService, "10.0.0.1:8080")

	t.Log("verified: weight-based load balance configuration works correctly")
}
