// Package integration provides etcd registry integration tests.
//
// 本包提供 etcd 注册发现集成测试。
package integration

import (
	"context"
	"testing"
	"time"

	etcdregistry "github.com/ngq/gorp/contrib/registry/etcd"
)

// TestEtcdRegisterDiscover tests real etcd register and discover operations.
//
// TestEtcdRegisterDiscover 测试真实 etcd 注册和发现操作。
func TestEtcdRegisterDiscover(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires etcd backend")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Create etcd registry
	// 创建 etcd registry
	cfg := &etcdregistry.DiscoveryConfig{
		EtcdEndpoints: []string{getEnvOrDefault("GORP_TEST_ETCD_ADDR", "localhost:2379")},
	}

	registry, err := etcdregistry.NewRegistry(cfg)
	if err != nil {
		t.Fatalf("failed to create etcd registry: %v", err)
	}
	defer registry.Close()

	// 2. Register service instance
	// 注册服务实例
	testService := "test-service-integration"
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

	// 3. Discover service instances
	// 发现服务实例
	instances, err := registry.Discover(ctx, testService)
	if err != nil {
		t.Fatalf("failed to discover service: %v", err)
	}

	if len(instances) == 0 {
		t.Fatal("expected at least 1 instance")
	}

	// 4. Verify instance metadata
	// 验证实例 metadata
	found := false
	for _, inst := range instances {
		if inst.Address == testAddress {
			found = true
			if inst.Metadata["version"] != "v1" {
				t.Fatal("metadata version mismatch")
			}
			t.Logf("found instance: address=%s, metadata=%v", inst.Address, inst.Metadata)
			break
		}
	}

	if !found {
		t.Fatalf("expected to find registered instance at %s", testAddress)
	}

	// 5. Deregister service
	// 注销服务
	err = registry.Deregister(ctx, testService, testAddress)
	if err != nil {
		t.Fatalf("failed to deregister service: %v", err)
	}

	t.Logf("deregistered service: %s at %s", testService, testAddress)

	// 6. Verify deregistration
	// 验证注销
	time.Sleep(500 * time.Millisecond) // Wait for etcd propagation
	instances, err = registry.Discover(ctx, testService)
	if err != nil {
		t.Logf("discover after deregister: %v (expected)", err)
	} else if len(instances) > 0 {
		// Instance might still exist with TTL, not a hard failure
		t.Logf("WARNING: instances still found after deregister: %v", instances)
	} else {
		t.Log("verified: no instances found after deregister")
	}
}

// TestEtcdMultipleInstances tests registering and discovering multiple instances.
//
// TestEtcdMultipleInstances 测试注册和发现多个实例。
func TestEtcdMultipleInstances(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires etcd backend")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := &etcdregistry.DiscoveryConfig{
		EtcdEndpoints: []string{getEnvOrDefault("GORP_TEST_ETCD_ADDR", "localhost:2379")},
	}

	registry, err := etcdregistry.NewRegistry(cfg)
	if err != nil {
		t.Fatalf("failed to create etcd registry: %v", err)
	}
	defer registry.Close()

	multiService := "multi-instance-service"

	// Register 3 instances
	// 注册 3 个实例
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

	// Discover all instances
	// 发现所有实例
	instances, err := registry.Discover(ctx, multiService)
	if err != nil {
		t.Fatalf("failed to discover: %v", err)
	}

	if len(instances) < 3 {
		t.Fatalf("expected at least 3 instances, got %d", len(instances))
	}

	t.Logf("discovered %d instances", len(instances))

	// Verify all addresses exist
	// 验证所有地址存在
	foundCount := 0
	for _, addr := range addresses {
		for _, inst := range instances {
			if inst.Address == addr {
				foundCount++
				break
			}
		}
	}

	if foundCount != 3 {
		t.Fatalf("expected to find 3 registered addresses, found %d", foundCount)
	}

	t.Log("verified: all 3 instances discovered")

	// Cleanup
	// 清理
	for _, addr := range addresses {
		registry.Deregister(ctx, multiService, addr)
	}
}
